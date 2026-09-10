package highlight

import (
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2"
)

// Session owns retained highlighting state for one editable document. Methods
// are safe for sequential or concurrent callers; Close is idempotent.
type Session struct {
	mu       sync.Mutex
	closed   bool
	source   string
	selected selection
	tree     *treeState
	cached   Result
	hasCache bool
}

// Highlight classifies source in one shot.
func (r *Registry) Highlight(source string, query Query) (Result, error) {
	session, err := r.NewSession(source, query)
	if err != nil {
		return Result{}, err
	}
	defer session.Close()
	return session.Result()
}

// NewSession parses source and retains the selected parser/tree for edits.
func (r *Registry) NewSession(source string, query Query) (*Session, error) {
	if !utf8.ValidString(source) {
		return nil, fmt.Errorf("source is not valid UTF-8")
	}
	selected := r.selectLanguage(query, source)
	session := &Session{source: source, selected: selected}
	if selected.language != nil {
		tree, err := newTreeState(selected.language, source)
		if err != nil {
			return nil, err
		}
		session.tree = tree
	}
	if _, err := session.computeLocked(); err != nil {
		if session.tree != nil {
			session.tree.Close()
		}
		return nil, err
	}
	return session, nil
}

// Result returns an immutable snapshot for the session's current source.
func (s *Session) Result() (Result, error) {
	if s == nil {
		return Result{}, ErrClosed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.resultLocked()
}

func (s *Session) resultLocked() (Result, error) {
	if s.closed {
		return Result{}, ErrClosed
	}
	if !s.hasCache {
		if _, err := s.computeLocked(); err != nil {
			return Result{}, err
		}
	}
	return cloneResult(s.cached), nil
}

func (s *Session) computeLocked() (Result, error) {
	if s.tree == nil {
		result, err := highlightChroma(s.source, s.selected)
		if err != nil {
			return Result{}, err
		}
		s.cached = result
		s.hasCache = true
		return result, nil
	}
	order := uint(0)
	raw, err := s.tree.rawSpans(0, 0, &order)
	if err != nil {
		return Result{}, err
	}
	spans := renderTreeSpans(s.source, raw)
	s.cached = Result{Source: s.source, Language: s.selected.name, Spans: spans}
	s.hasCache = true
	return s.cached, nil
}

// ApplyEdit updates the retained syntax tree incrementally and returns the
// matching source/classification snapshot.
func (s *Session) ApplyEdit(edit Edit) (Result, error) {
	if s == nil {
		return Result{}, ErrClosed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return Result{}, ErrClosed
	}
	if err := validateEdit(s.source, edit); err != nil {
		return Result{}, err
	}

	var source strings.Builder
	source.Grow(len(s.source) - (edit.OldEndByte - edit.StartByte) + len(edit.NewText))
	source.WriteString(s.source[:edit.StartByte])
	source.WriteString(edit.NewText)
	source.WriteString(s.source[edit.OldEndByte:])
	next := source.String()
	if s.tree == nil {
		result, err := highlightChroma(next, s.selected)
		if err != nil {
			return Result{}, err
		}
		s.source = next
		s.cached = result
		s.hasCache = true
		return cloneResult(result), nil
	}
	if err := s.tree.applyEdit(edit, next); err != nil {
		return Result{}, err
	}
	s.source = next
	s.hasCache = false
	if _, err := s.computeLocked(); err != nil {
		return Result{}, err
	}
	return cloneResult(s.cached), nil
}

// Close releases every native parser, tree, query, and cursor owned by the
// session, including recursively injected trees.
func (s *Session) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	s.cached = Result{}
	s.hasCache = false
	if s.tree != nil {
		s.tree.Close()
		s.tree = nil
	}
	return nil
}

func cloneResult(result Result) Result {
	result.Spans = append([]Span(nil), result.Spans...)
	return result
}

func validateEdit(source string, edit Edit) error {
	if edit.StartByte < 0 || edit.OldEndByte < edit.StartByte || edit.OldEndByte > len(source) {
		return fmt.Errorf(
			"invalid edit range [%d,%d) for %d-byte source",
			edit.StartByte,
			edit.OldEndByte,
			len(source),
		)
	}
	if !runeBoundary(source, edit.StartByte) || !runeBoundary(source, edit.OldEndByte) {
		return fmt.Errorf("edit range [%d,%d) does not fall on UTF-8 boundaries", edit.StartByte, edit.OldEndByte)
	}
	if !utf8.ValidString(edit.NewText) {
		return fmt.Errorf("edit replacement is not valid UTF-8")
	}
	return nil
}

func runeBoundary(source string, offset int) bool {
	return offset == 0 || offset == len(source) || utf8.RuneStart(source[offset])
}

func highlightChroma(source string, selected selection) (Result, error) {
	iterator, err := selected.lexer.Tokenise(&chroma.TokeniseOptions{
		State:    "root",
		Nested:   true,
		EnsureLF: false,
	}, source)
	if err != nil {
		return Result{}, fmt.Errorf("tokenise %s source: %w", selected.name, err)
	}
	tokens := iterator.Tokens()
	spans := make([]Span, 0, len(tokens))
	offset := 0
	for _, token := range tokens {
		end := offset + len(token.Value)
		if end > len(source) || source[offset:end] != token.Value {
			return Result{}, sourceContractError(selected.name, source, tokens)
		}
		appendSpan(&spans, Span{
			StartByte:   offset,
			EndByte:     end,
			TokenType:   token.Type,
			ChromaClass: ChromaClass(token.Type),
		})
		offset = end
	}
	if offset != len(source) {
		return Result{}, sourceContractError(selected.name, source, tokens)
	}
	return Result{Source: source, Language: selected.name, Spans: spans}, nil
}

func sourceContractError(language, source string, tokens []chroma.Token) error {
	var tokenText strings.Builder
	for _, token := range tokens {
		tokenText.WriteString(token.Value)
	}
	actual := tokenText.String()
	limit := min(len(source), len(actual))
	offset := 0
	for offset < limit && source[offset] == actual[offset] {
		offset++
	}
	return &SourceContractError{
		Language:   language,
		Offset:     offset,
		SourceSize: len(source),
		TokenSize:  len(actual),
	}
}
