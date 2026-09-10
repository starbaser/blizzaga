package highlight

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

type languageSpec struct {
	name       string
	aliases    []string
	filenames  []string
	mimeTypes  []string
	priority   float32
	grammar    *languageGrammar
	analyse    func(string) float32
	injections []injectionSpec
}

type injectionSpec struct {
	query     string
	languages map[string]*languageSpec
}

// Registry is a validated, immutable collection of statically linked
// Tree-sitter languages layered over Chroma's complete lexer catalog.
type Registry struct {
	languages []*languageSpec
	byName    map[string]*languageSpec
}

// NewRegistry validates and freezes statically linked Tree-sitter languages.
// Registered languages take precedence over Chroma for their names, aliases,
// filenames, MIME types, and equally ranked content analysis.
func NewRegistry(languages ...Language) (*Registry, error) {
	registry := &Registry{byName: make(map[string]*languageSpec)}
	pendingInjections := make([][]Injection, len(languages))
	for i, language := range languages {
		spec, err := freezeLanguage(language)
		if err != nil {
			return nil, fmt.Errorf("language %d: %w", i, err)
		}

		identifiers := append([]string{spec.name}, spec.aliases...)
		for _, identifier := range identifiers {
			key := normalizeIdentifier(identifier)
			if existing := registry.byName[key]; existing != nil {
				if existing == spec {
					continue
				}
				return nil, fmt.Errorf(
					"language identifier %q is shared by %q and %q",
					identifier,
					existing.name,
					spec.name,
				)
			}
			registry.byName[key] = spec
		}
		registry.languages = append(registry.languages, spec)
		pendingInjections[i] = cloneInjections(language.Injections)
	}

	for i, injections := range pendingInjections {
		for j, injection := range injections {
			frozen, err := registry.freezeInjection(registry.languages[i], injection)
			if err != nil {
				return nil, fmt.Errorf(
					"language %q injection %d: %w",
					registry.languages[i].name,
					j,
					err,
				)
			}
			registry.languages[i].injections = append(registry.languages[i].injections, frozen)
		}
	}
	if err := validateInjectionGraph(registry.languages); err != nil {
		return nil, err
	}

	return registry, nil
}

func freezeLanguage(language Language) (*languageSpec, error) {
	if strings.TrimSpace(language.Name) == "" {
		return nil, fmt.Errorf("name is empty")
	}
	if language.Grammar == nil {
		return nil, fmt.Errorf("Tree-sitter grammar is nil")
	}
	if strings.TrimSpace(language.Highlights) == "" {
		return nil, fmt.Errorf("highlight query is empty")
	}
	if len(language.Captures) == 0 {
		return nil, fmt.Errorf("capture map is empty")
	}

	aliases, err := cleanStrings("alias", language.Aliases)
	if err != nil {
		return nil, err
	}
	filenames, err := cleanStrings("filename", language.Filenames)
	if err != nil {
		return nil, err
	}
	for _, pattern := range filenames {
		if _, err := filepath.Match(pattern, "source"); err != nil {
			return nil, fmt.Errorf("invalid filename pattern %q: %w", pattern, err)
		}
	}
	mimeTypes, err := cleanStrings("MIME type", language.MIMETypes)
	if err != nil {
		return nil, err
	}

	captures := make(CaptureMap, len(language.Captures))
	for name, mapping := range language.Captures {
		if strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("capture name is empty")
		}
		if !mapping.TokenType.IsATokenType() {
			return nil, fmt.Errorf("capture %q has invalid Chroma token type %d", name, mapping.TokenType)
		}
		captures[name] = mapping
	}

	grammar, err := newLanguageGrammar(language.Grammar, language.Highlights, captures)
	if err != nil {
		return nil, err
	}
	return &languageSpec{
		name:      language.Name,
		aliases:   aliases,
		filenames: filenames,
		mimeTypes: mimeTypes,
		priority:  language.Priority,
		grammar:   grammar,
		analyse:   language.Analyse,
	}, nil
}

func (r *Registry) freezeInjection(parent *languageSpec, injection Injection) (injectionSpec, error) {
	if strings.TrimSpace(injection.Query) == "" {
		return injectionSpec{}, fmt.Errorf("query is empty")
	}
	if len(injection.Languages) == 0 {
		return injectionSpec{}, fmt.Errorf("language selector map is empty")
	}

	targets := make(map[string]*languageSpec, len(injection.Languages))
	for selector, name := range injection.Languages {
		selector = normalizeIdentifier(selector)
		if selector == "" {
			return injectionSpec{}, fmt.Errorf("language selector is empty")
		}
		target := r.byName[normalizeIdentifier(name)]
		if target == nil {
			return injectionSpec{}, fmt.Errorf("selector %q targets unknown Tree-sitter language %q", selector, name)
		}
		if _, exists := targets[selector]; exists {
			return injectionSpec{}, fmt.Errorf("duplicate language selector %q", selector)
		}
		targets[selector] = target
	}
	if err := validateInjectionQuery(parent.grammar.language, injection.Query, targets); err != nil {
		return injectionSpec{}, err
	}
	return injectionSpec{query: injection.Query, languages: targets}, nil
}

func validateInjectionGraph(languages []*languageSpec) error {
	const (
		unvisited = iota
		visiting
		visited
	)
	states := make(map[*languageSpec]int, len(languages))
	var visit func(*languageSpec) error
	visit = func(language *languageSpec) error {
		switch states[language] {
		case visiting:
			return fmt.Errorf("Tree-sitter injection cycle reaches language %q", language.name)
		case visited:
			return nil
		}
		states[language] = visiting
		for _, injection := range language.injections {
			for _, target := range injection.languages {
				if err := visit(target); err != nil {
					return err
				}
			}
		}
		states[language] = visited
		return nil
	}
	for _, language := range languages {
		if err := visit(language); err != nil {
			return err
		}
	}
	return nil
}

// Languages returns copied metadata in registration order.
func (r *Registry) Languages() []LanguageInfo {
	if r == nil {
		return nil
	}
	infos := make([]LanguageInfo, 0, len(r.languages))
	for _, language := range r.languages {
		infos = append(infos, LanguageInfo{
			Name:      language.name,
			Aliases:   append([]string(nil), language.aliases...),
			Filenames: append([]string(nil), language.filenames...),
			MIMETypes: append([]string(nil), language.mimeTypes...),
			Priority:  language.priority,
		})
	}
	return infos
}

type selection struct {
	name     string
	language *languageSpec
	lexer    chroma.Lexer
}

func (r *Registry) selectLanguage(query Query, source string) selection {
	if r == nil {
		return fallbackSelection()
	}
	if query.Language != "" {
		if language := r.byName[normalizeIdentifier(query.Language)]; language != nil {
			return selection{name: language.name, language: language}
		}
		if lexer := lexers.Get(query.Language); lexer != nil {
			return selection{name: lexer.Config().Name, lexer: lexer}
		}
		return fallbackSelection()
	}
	if query.Filename != "" {
		if language := r.matchFilename(query.Filename); language != nil {
			return selection{name: language.name, language: language}
		}
		if lexer := lexers.Match(query.Filename); lexer != nil {
			return selection{name: lexer.Config().Name, lexer: lexer}
		}
	}
	if query.MIMEType != "" {
		if language := r.matchMIMEType(query.MIMEType); language != nil {
			return selection{name: language.name, language: language}
		}
		if lexer := lexers.MatchMimeType(query.MIMEType); lexer != nil {
			return selection{name: lexer.Config().Name, lexer: lexer}
		}
	}

	var selected *languageSpec
	highest := float32(0)
	for _, language := range r.languages {
		if language.analyse == nil {
			continue
		}
		if score := language.analyse(source); score > highest {
			selected = language
			highest = score
		}
	}
	chromaLexer := lexers.Analyse(source)
	if chromaLexer != nil && chromaLexer.AnalyseText(source) > highest {
		return selection{name: chromaLexer.Config().Name, lexer: chromaLexer}
	}
	if selected != nil {
		return selection{name: selected.name, language: selected}
	}
	if chromaLexer != nil {
		return selection{name: chromaLexer.Config().Name, lexer: chromaLexer}
	}
	return fallbackSelection()
}

func (r *Registry) matchFilename(filename string) *languageSpec {
	base := filepath.Base(filename)
	var matches []*languageSpec
	for _, language := range r.languages {
		for _, pattern := range language.filenames {
			if matched, _ := filepath.Match(pattern, base); matched {
				matches = append(matches, language)
				break
			}
		}
	}
	return bestLanguage(matches)
}

func (r *Registry) matchMIMEType(mimeType string) *languageSpec {
	var matches []*languageSpec
	for _, language := range r.languages {
		for _, candidate := range language.mimeTypes {
			if strings.EqualFold(candidate, mimeType) {
				matches = append(matches, language)
				break
			}
		}
	}
	return bestLanguage(matches)
}

func bestLanguage(languages []*languageSpec) *languageSpec {
	if len(languages) == 0 {
		return nil
	}
	sort.SliceStable(languages, func(i, j int) bool {
		if languages[i].priority != languages[j].priority {
			return languages[i].priority > languages[j].priority
		}
		return languages[i].name < languages[j].name
	})
	return languages[0]
}

func fallbackSelection() selection {
	return selection{name: lexers.Fallback.Config().Name, lexer: lexers.Fallback}
}

func cleanStrings(kind string, values []string) ([]string, error) {
	cleaned := make([]string, len(values))
	seen := make(map[string]struct{}, len(values))
	for i, value := range values {
		if strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("%s %d is empty", kind, i)
		}
		key := normalizeIdentifier(value)
		if _, ok := seen[key]; ok {
			return nil, fmt.Errorf("duplicate %s %q", kind, value)
		}
		seen[key] = struct{}{}
		cleaned[i] = value
	}
	return cleaned, nil
}

func cloneInjections(injections []Injection) []Injection {
	cloned := make([]Injection, len(injections))
	for i, injection := range injections {
		languages := make(map[string]string, len(injection.Languages))
		for selector, language := range injection.Languages {
			languages[selector] = language
		}
		cloned[i] = Injection{Query: injection.Query, Languages: languages}
	}
	return cloned
}

func normalizeIdentifier(identifier string) string {
	return strings.ToLower(strings.TrimSpace(identifier))
}

var (
	defaultRegistryOnce sync.Once
	defaultRegistry     *Registry
	defaultRegistryErr  error
)

// DefaultRegistry returns the process-wide immutable registry containing the
// bundled Go and BAML grammars over Chroma's fallback catalog.
func DefaultRegistry() (*Registry, error) {
	defaultRegistryOnce.Do(func() {
		defaultRegistry, defaultRegistryErr = NewRegistry(BuiltinLanguages()...)
	})
	return defaultRegistry, defaultRegistryErr
}

// Highlight classifies source with DefaultRegistry.
func Highlight(source string, query Query) (Result, error) {
	registry, err := DefaultRegistry()
	if err != nil {
		return Result{}, err
	}
	return registry.Highlight(source, query)
}
