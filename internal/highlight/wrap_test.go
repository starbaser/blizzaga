package highlight

import (
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
	"github.com/charmbracelet/x/cellbuf"
)

func TestWrapTokensPreservesStylesAcrossInsertedNewlines(t *testing.T) {
	source := `"This is a long string literal"`
	iterator := chroma.Literator(chroma.Token{Type: chroma.LiteralString, Value: source})

	wrappedIterator, wrapped, err := WrapTokens(iterator, 10)
	if err != nil {
		t.Fatal(err)
	}
	if want := cellbuf.Wrap(source, 10, ""); wrapped != want {
		t.Fatalf("wrapped source = %q, want %q", wrapped, want)
	}

	for _, token := range wrappedIterator.Tokens() {
		if token.Type != chroma.LiteralString {
			t.Fatalf("wrapped token type = %s, want LiteralString", token.Type)
		}
	}
}

func TestWrapTokensPreservesMixedTokenTypes(t *testing.T) {
	sourceTokens := []chroma.Token{
		{Type: chroma.NameFunction, Value: "render"},
		{Type: chroma.Text, Value: "("},
		{Type: chroma.LiteralString, Value: `"one two three four"`},
		{Type: chroma.Text, Value: ")"},
	}
	wantWrapped := cellbuf.Wrap(joinTokenValues(sourceTokens), 9, "")

	iterator, wrapped, err := WrapTokens(chroma.Literator(sourceTokens...), 9)
	if err != nil {
		t.Fatal(err)
	}
	if wrapped != wantWrapped {
		t.Fatalf("wrapped source = %q, want %q", wrapped, wantWrapped)
	}

	output := iterator.Tokens()
	if got := joinTokenValues(output); got != wrapped {
		t.Fatalf("projected token source = %q, want %q", got, wrapped)
	}
	for _, token := range output {
		if strings.Contains(token.Value, "one") || strings.Contains(token.Value, "two") ||
			strings.Contains(token.Value, "three") || strings.Contains(token.Value, "four") {
			if token.Type != chroma.LiteralString {
				t.Errorf("string content %q has token type %s", token.Value, token.Type)
			}
		}
	}
}

func TestWrapTokensDropsOnlyWhitespaceRemovedByWrapper(t *testing.T) {
	sourceTokens := []chroma.Token{
		{Type: chroma.Keyword, Value: "word"},
		{Type: chroma.TextWhitespace, Value: "   "},
		{Type: chroma.Name, Value: "next"},
	}

	iterator, wrapped, err := WrapTokens(chroma.Literator(sourceTokens...), 5)
	if err != nil {
		t.Fatal(err)
	}
	if got := joinTokenValues(iterator.Tokens()); got != wrapped {
		t.Fatalf("projected token source = %q, want %q", got, wrapped)
	}
}

func TestWrapTokensMatchesCellBufferForPlainText(t *testing.T) {
	tests := []struct {
		name   string
		source string
		width  int
	}{
		{name: "hard wrap", source: "abcdefghijk", width: 4},
		{name: "hyphen", source: "alpha-beta-gamma", width: 7},
		{name: "wide runes", source: "日本語 sample", width: 6},
		{name: "emoji", source: "hello 👩🏽‍💻 world", width: 8},
		{name: "tabs", source: "one\t\ttwo", width: 5},
		{name: "leading whitespace", source: "    value", width: 5},
		{name: "trailing whitespace", source: "value    ", width: 5},
		{name: "explicit newline", source: "first   \nsecond", width: 6},
		{name: "unicode whitespace", source: "first\u2003second", width: 6},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iterator := chroma.Literator(chroma.Token{Type: chroma.Text, Value: test.source})
			wrappedIterator, wrapped, err := WrapTokens(iterator, test.width)
			if err != nil {
				t.Fatal(err)
			}
			if want := cellbuf.Wrap(test.source, test.width, ""); wrapped != want {
				t.Fatalf("wrapped source = %q, want %q", wrapped, want)
			}
			if got := joinTokenValues(wrappedIterator.Tokens()); got != wrapped {
				t.Fatalf("projected source = %q, want %q", got, wrapped)
			}
		})
	}
}

func joinTokenValues(tokens []chroma.Token) string {
	var output strings.Builder
	for _, token := range tokens {
		output.WriteString(token.Value)
	}
	return output.String()
}
