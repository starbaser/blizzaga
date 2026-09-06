package chromalexer

import (
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"

	tshighlight "github.com/starbaser/blizzaga/internal/highlight/treesitter"
)

func TestRenderTokensResolvesExactAndNestedCaptures(t *testing.T) {
	source := `call("a\n")`
	spans := []tshighlight.Span{
		{StartByte: 0, EndByte: 4, Capture: "variable", PatternIndex: 2},
		{StartByte: 0, EndByte: 4, Capture: "function", PatternIndex: 0},
		{StartByte: 5, EndByte: 10, Capture: "string", PatternIndex: 3},
		{StartByte: 7, EndByte: 9, Capture: "escape", PatternIndex: 4},
	}

	tokens := renderTokens(source, spans, StandardCaptures())
	want := []chroma.Token{
		{Type: chroma.NameFunction, Value: "call"},
		{Type: chroma.Text, Value: "("},
		{Type: chroma.LiteralString, Value: `"a`},
		{Type: chroma.LiteralStringEscape, Value: `\n`},
		{Type: chroma.LiteralString, Value: `"`},
		{Type: chroma.Text, Value: ")"},
	}
	assertTokens(t, tokens, want)
}

func TestRenderTokensHonoursQueryPriority(t *testing.T) {
	spans := []tshighlight.Span{
		{StartByte: 0, EndByte: 4, Capture: "function", Priority: 100, PrioritySet: true},
		{StartByte: 0, EndByte: 4, Capture: "variable", Priority: 200, PrioritySet: true},
	}
	tokens := renderTokens("name", spans, StandardCaptures())
	assertTokens(t, tokens, []chroma.Token{{Type: chroma.NameVariable, Value: "name"}})
}

func TestRenderTokensUsesConventionalDefaultQueryPriority(t *testing.T) {
	spans := []tshighlight.Span{
		{StartByte: 0, EndByte: 4, Capture: "function"},
		{StartByte: 0, EndByte: 4, Capture: "variable", Priority: 90, PrioritySet: true},
	}
	tokens := renderTokens("name", spans, StandardCaptures())
	assertTokens(t, tokens, []chroma.Token{{Type: chroma.NameFunction, Value: "name"}})
}

func TestRenderTokensPreservesUnicodeAndUncapturedText(t *testing.T) {
	source := "π := value\n"
	spans := []tshighlight.Span{
		{StartByte: uint(strings.Index(source, "value")), EndByte: uint(strings.Index(source, "value") + len("value")), Capture: "variable"},
		{StartByte: 0, EndByte: 2, Capture: "unknown.capture"},
	}

	tokens := renderTokens(source, spans, StandardCaptures())
	var roundTrip strings.Builder
	for _, token := range tokens {
		roundTrip.WriteString(token.Value)
	}
	if roundTrip.String() != source {
		t.Fatalf("rendered source = %q, want %q", roundTrip.String(), source)
	}
	if tokens[len(tokens)-1].Type != chroma.TextWhitespace {
		t.Fatalf("trailing token type = %s, want TextWhitespace", tokens[len(tokens)-1].Type)
	}
}

func TestCaptureMapFallsBackToParentCapture(t *testing.T) {
	style, ok := CaptureMap{
		"function": {TokenType: chroma.NameFunction, Priority: 1},
	}.lookup("function.specialized.call")
	if !ok || style.TokenType != chroma.NameFunction {
		t.Fatalf("style = (%v, %t), want NameFunction", style, ok)
	}
}

func assertTokens(t *testing.T, got, want []chroma.Token) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("token count = %d, want %d\ngot:  %#v\nwant: %#v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}
