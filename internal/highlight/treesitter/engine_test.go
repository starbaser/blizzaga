package treesitter

import (
	"strings"
	"testing"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

func goLanguage() *sitter.Language {
	return sitter.NewLanguage(tree_sitter_go.Language())
}

func TestEngineHighlightsNamedCaptures(t *testing.T) {
	engine, err := NewEngine(goLanguage(), `
(function_declaration name: (identifier) @function)
(comment) @comment
`)
	if err != nil {
		t.Fatal(err)
	}

	source := []byte("package main\nfunc render() {} // hello\n")
	spans, err := engine.Highlight(source)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"function": "render",
		"comment":  "// hello",
	}
	for capture, text := range want {
		if !hasSpan(spans, capture, text, source) {
			t.Errorf("missing %s capture for %q in %#v", capture, text, spans)
		}
	}
}

func TestEngineAppliesPrioritySetting(t *testing.T) {
	engine, err := NewEngine(goLanguage(), `
((identifier) @variable
  (#set! priority "175"))
`)
	if err != nil {
		t.Fatal(err)
	}

	spans, err := engine.Highlight([]byte("var value = 1\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) == 0 {
		t.Fatal("expected an identifier capture")
	}
	if !spans[0].PrioritySet || spans[0].Priority != 175 {
		t.Fatalf("priority = (%d, %t), want (175, true)", spans[0].Priority, spans[0].PrioritySet)
	}
}

func TestEngineRejectsUnsupportedPredicates(t *testing.T) {
	_, err := NewEngine(goLanguage(), `
((identifier) @variable
  (#custom? @variable))
`)
	if err == nil {
		t.Fatal("expected unsupported predicate error")
	}
	if !strings.Contains(err.Error(), "unsupported predicate") {
		t.Fatalf("error = %q, want unsupported predicate", err)
	}
}

func TestEngineRejectsMalformedQuery(t *testing.T) {
	_, err := NewEngine(goLanguage(), `(not_a_go_node) @thing`)
	if err == nil {
		t.Fatal("expected query compilation error")
	}
}

func hasSpan(spans []Span, capture, text string, source []byte) bool {
	for _, span := range spans {
		if span.Capture == capture && string(source[span.StartByte:span.EndByte]) == text {
			return true
		}
	}
	return false
}
