package highlight

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

func TestDefaultRegistryUsesTreeSitterBeforeChroma(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}

	source := "package main\r\nfunc render() { println(\"hello\") }"
	result, err := registry.Highlight(source, Query{Language: "golang"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Language != "Go" {
		t.Fatalf("language = %q, want Go", result.Language)
	}
	assertExactSource(t, result, source)
	assertClassifies(t, result, "render", chroma.NameFunction, "tree-sitter.function")
	assertClassifies(t, result, "println", chroma.NameBuiltin, "tree-sitter.function.builtin")
}

func TestChromaFallbackPreservesExactSource(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		language string
		source   string
	}{
		{language: "css", source: "a { color: red; }\r\n/* tail */"},
		{language: "json", source: "{\r\n  \"hello\": \"世界\"\r\n}"},
		{language: "haskell", source: "main = putStrLn \"hello\"\r\n"},
		{language: "not-a-language", source: "plain\r\ntext"},
	}
	for _, test := range tests {
		t.Run(test.language, func(t *testing.T) {
			t.Parallel()
			result, err := registry.Highlight(test.source, Query{Language: test.language})
			if err != nil {
				t.Fatal(err)
			}
			assertExactSource(t, result, test.source)
		})
	}
}

func TestDefaultRegistryBundlesBAML(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	source := "function Center(ctx: training.numeric.Context) -> training.numeric.Program {\r\n  ctx.finish(1024.0)\r\n}"
	result, err := registry.Highlight(source, Query{Filename: "program.baml"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Language != "BAML" {
		t.Fatalf("language = %q, want BAML", result.Language)
	}
	assertExactSource(t, result, source)
	assertClassifies(t, result, "Center", chroma.NameFunction, "tree-sitter.function")
	assertClassifies(t, result, "finish", chroma.NameFunction, "tree-sitter.function.method")
	assertClassifies(t, result, "1024.0", chroma.LiteralNumber, "tree-sitter.constant.numeric")
}

func TestIncrementalMatchesFreshAcrossEdits(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	session, err := registry.NewSession(
		"package main\n\nfunc greet() { println(\"héllo\") }\n",
		Query{Language: "go"},
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })

	edits := []Edit{
		{StartByte: len("package main\n"), OldEndByte: len("package main\n"), NewText: "// 世界\n"},
		{StartByte: len("package main\n// 世界\n\nfunc "), OldEndByte: len("package main\n// 世界\n\nfunc greet"), NewText: "render"},
		{StartByte: len("package main\n// 世界\n\nfunc render() { println(\"héllo\") "), OldEndByte: len("package main\n// 世界\n\nfunc render() { println(\"héllo\") }"), NewText: ""},
		{StartByte: len("package main\n// 世界\n\nfunc render() { println(\""), OldEndByte: len("package main\n// 世界\n\nfunc render() { println(\"héllo"), NewText: "hi 👩🏽‍💻"},
	}
	for index, edit := range edits {
		incremental, err := session.ApplyEdit(edit)
		if err != nil {
			t.Fatalf("edit %d: %v", index, err)
		}
		fresh, err := registry.Highlight(incremental.Source, Query{Language: "go"})
		if err != nil {
			t.Fatalf("fresh edit %d: %v", index, err)
		}
		if !reflect.DeepEqual(incremental, fresh) {
			t.Fatalf("edit %d incremental differs from fresh\nincremental: %#v\nfresh: %#v", index, incremental, fresh)
		}
		assertExactSource(t, incremental, incremental.Source)
	}
}

func TestSessionRejectsSplitUTF8EditAndCloses(t *testing.T) {
	t.Parallel()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	session, err := registry.NewSession("package π\n", Query{Language: "go"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.ApplyEdit(Edit{StartByte: 9, OldEndByte: 9, NewText: "x"}); err == nil {
		t.Fatal("expected split UTF-8 edit error")
	}
	first, err := session.Result()
	if err != nil {
		t.Fatal(err)
	}
	first.Spans[0].TokenType = chroma.Error
	second, err := session.Result()
	if err != nil {
		t.Fatal(err)
	}
	if second.Spans[0].TokenType == chroma.Error {
		t.Fatal("mutating a returned snapshot changed the retained session result")
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	if err := session.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if _, err := session.Result(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Result error = %v, want ErrClosed", err)
	}
}

func TestRegistryCopiesAndValidatesLanguageDeclarations(t *testing.T) {
	t.Parallel()
	aliases := []string{"test"}
	captures := CaptureMap{"function": {TokenType: chroma.NameFunction, Priority: 10}}
	language := Language{
		Name:       "Test",
		Aliases:    aliases,
		Filenames:  []string{"*.test"},
		Grammar:    goGrammar(),
		Highlights: `(function_declaration name: (identifier) @function)`,
		Captures:   captures,
	}
	registry, err := NewRegistry(language)
	if err != nil {
		t.Fatal(err)
	}
	aliases[0] = "mutated"
	captures["function"] = CaptureMapping{TokenType: chroma.NameVariable}

	result, err := registry.Highlight("package p\nfunc chosen() {}", Query{Language: "test"})
	if err != nil {
		t.Fatal(err)
	}
	assertClassifies(t, result, "chosen", chroma.NameFunction, "tree-sitter.function")
	infos := registry.Languages()
	infos[0].Aliases[0] = "also-mutated"
	if result, err := registry.Highlight("package p\nfunc chosen() {}", Query{Language: "test"}); err != nil || result.Language != "Test" {
		t.Fatalf("frozen alias lookup = (%q, %v)", result.Language, err)
	}

	invalid := []Language{
		{Name: "", Grammar: goGrammar(), Highlights: `(identifier) @variable`, Captures: StandardCaptures()},
		{Name: "Bad", Grammar: goGrammar(), Highlights: `(not_a_go_node) @variable`, Captures: StandardCaptures()},
		{Name: "Bad", Grammar: goGrammar(), Highlights: `(identifier) @variable`, Captures: nil},
	}
	for index, declaration := range invalid {
		if _, err := NewRegistry(declaration); err == nil {
			t.Errorf("invalid declaration %d succeeded", index)
		}
	}
	if _, err := NewRegistry(language, language); err == nil {
		t.Error("duplicate identifiers succeeded")
	}
}

func TestTextPredicatesAndOverlapPrecedence(t *testing.T) {
	t.Parallel()
	registry, err := NewRegistry(Language{
		Name:       "PredicateGo",
		Aliases:    []string{"predicate-go"},
		Grammar:    goGrammar(),
		Highlights: "((identifier) @variable (#match? @variable \"^pick\"))\n((identifier) @function (#eq? @function \"picked\"))",
		Captures:   StandardCaptures(),
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := registry.Highlight("package p\nvar picked = skipped", Query{Language: "predicate-go"})
	if err != nil {
		t.Fatal(err)
	}
	assertClassifies(t, result, "picked", chroma.NameFunction, "tree-sitter.function")
	assertClassifies(t, result, "skipped", chroma.Text, "")
}

func TestQueryPriorityOverridesCaptureMapping(t *testing.T) {
	t.Parallel()
	registry, err := NewRegistry(Language{
		Name:    "PriorityGo",
		Grammar: goGrammar(),
		Highlights: `
((identifier) @function)
((identifier) @variable
  (#set! priority "200"))
`,
		Captures: StandardCaptures(),
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := registry.Highlight("package p\nvar chosen = 1", Query{Language: "PriorityGo"})
	if err != nil {
		t.Fatal(err)
	}
	assertClassifies(t, result, "chosen", chroma.NameVariable, "tree-sitter.variable")
}

func TestRegisteredInjectionRebasesSpansAndSurvivesEdits(t *testing.T) {
	t.Parallel()
	inner := Language{
		Name:       "Inner",
		Aliases:    []string{"inner"},
		Grammar:    goGrammar(),
		Highlights: `(raw_string_literal) @function`,
		Captures:   StandardCaptures(),
	}
	outer := Language{
		Name:       "Outer",
		Aliases:    []string{"outer"},
		Grammar:    goGrammar(),
		Highlights: `(raw_string_literal) @string`,
		Captures:   StandardCaptures(),
		Injections: []Injection{{
			Query:     `((raw_string_literal) @injection.content (#set! injection.language "fml"))`,
			Languages: map[string]string{"fml": "inner"},
		}},
	}
	registry, err := NewRegistry(inner, outer)
	if err != nil {
		t.Fatal(err)
	}
	source := "package p\nvar template = `hello`\n"
	session, err := registry.NewSession(source, Query{Language: "outer"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })

	initial, err := session.Result()
	if err != nil {
		t.Fatal(err)
	}
	assertClassifies(t, initial, "`hello`", chroma.NameFunction, "tree-sitter.function")

	insert := strings.Index(source, "var template")
	incremental, err := session.ApplyEdit(Edit{StartByte: insert, OldEndByte: insert, NewText: "// moved\n"})
	if err != nil {
		t.Fatal(err)
	}
	assertClassifies(t, incremental, "`hello`", chroma.NameFunction, "tree-sitter.function")
	fresh, err := registry.Highlight(incremental.Source, Query{Language: "outer"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(incremental, fresh) {
		t.Fatalf("injected incremental differs from fresh\nincremental: %#v\nfresh: %#v", incremental, fresh)
	}
}

func TestInjectionValidation(t *testing.T) {
	t.Parallel()
	base := Language{
		Name:       "Base",
		Grammar:    goGrammar(),
		Highlights: `(identifier) @variable`,
		Captures:   StandardCaptures(),
		Injections: []Injection{{
			Query:     `(raw_string_literal) @other`,
			Languages: map[string]string{"go": "missing"},
		}},
	}
	if _, err := NewRegistry(base); err == nil {
		t.Fatal("invalid injection succeeded")
	}
	cyclic := Language{
		Name:       "Cyclic",
		Grammar:    goGrammar(),
		Highlights: `(identifier) @variable`,
		Captures:   StandardCaptures(),
		Injections: []Injection{{
			Query:     `((raw_string_literal) @injection.content (#set! injection.language "self"))`,
			Languages: map[string]string{"self": "cyclic"},
		}},
	}
	if _, err := NewRegistry(cyclic); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("cyclic injection error = %v", err)
	}
}

func TestChromaSourceContractError(t *testing.T) {
	t.Parallel()
	lexer := &changingLexer{config: chroma.Config{Name: "changing"}}
	_, err := highlightChroma("a\r\nb", selection{name: "changing", lexer: lexer})
	var contractErr *SourceContractError
	if !errors.As(err, &contractErr) {
		t.Fatalf("error = %v, want SourceContractError", err)
	}
	if contractErr.Offset != 1 || contractErr.SourceSize != 4 || contractErr.TokenSize != 3 {
		t.Fatalf("contract error = %#v", contractErr)
	}
}

func TestSrceryStyleIsPublic(t *testing.T) {
	t.Parallel()
	style := SrceryStyle()
	if style == nil || style.Name != "srcery" {
		t.Fatalf("SrceryStyle = %#v", style)
	}
	if style.Get(chroma.Keyword).Colour.IsSet() == false {
		t.Fatal("Srcery keyword colour is unset")
	}
}

func TestChromaClass(t *testing.T) {
	t.Parallel()
	if got := ChromaClass(chroma.KeywordDeclaration); got != "chroma.keyword.declaration" {
		t.Fatalf("class = %q", got)
	}
}

func goGrammar() *sitter.Language {
	return sitter.NewLanguage(tree_sitter_go.Language())
}

func assertExactSource(t *testing.T, result Result, want string) {
	t.Helper()
	if result.Source != want {
		t.Fatalf("source = %q, want %q", result.Source, want)
	}
	var joined strings.Builder
	for _, token := range result.Tokens() {
		joined.WriteString(token.Value)
	}
	if joined.String() != want {
		t.Fatalf("token concatenation = %q, want %q", joined.String(), want)
	}
	position := 0
	for index, span := range result.Spans {
		if span.StartByte != position || span.EndByte <= span.StartByte {
			t.Fatalf("span %d = %#v after byte %d", index, span, position)
		}
		position = span.EndByte
	}
	if position != len(want) {
		t.Fatalf("spans end at byte %d, want %d", position, len(want))
	}
}

func assertClassifies(t *testing.T, result Result, text string, tokenType chroma.TokenType, captureClass string) {
	t.Helper()
	start := strings.Index(result.Source, text)
	if start < 0 {
		t.Fatalf("source does not contain %q", text)
	}
	end := start + len(text)
	for _, span := range result.Spans {
		if span.StartByte <= start && span.EndByte >= end {
			if span.TokenType != tokenType || span.CaptureClass != captureClass {
				t.Fatalf("%q span = %#v, want token %s capture %q", text, span, tokenType, captureClass)
			}
			return
		}
	}
	t.Fatalf("no single span classifies %q in %#v", text, result.Spans)
}

type changingLexer struct {
	config chroma.Config
}

func (l *changingLexer) Config() *chroma.Config { return &l.config }
func (l *changingLexer) Tokenise(_ *chroma.TokeniseOptions, text string) (chroma.Iterator, error) {
	return chroma.Literator(chroma.Token{Type: chroma.Text, Value: strings.ReplaceAll(text, "\r\n", "\n")}), nil
}
func (l *changingLexer) SetRegistry(_ *chroma.LexerRegistry) chroma.Lexer { return l }
func (l *changingLexer) SetAnalyser(_ func(string) float32) chroma.Lexer  { return l }
func (l *changingLexer) AnalyseText(string) float32                       { return 0 }
