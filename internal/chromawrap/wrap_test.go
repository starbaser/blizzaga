package chromawrap

import (
	"slices"
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
	"github.com/charmbracelet/x/cellbuf"
)

func TestWrapTokensPreservesStylesAcrossInsertedNewlines(t *testing.T) {
	source := `"This is a long string literal"`
	iterator := chroma.Literator(chroma.Token{Type: chroma.LiteralString, Value: source})

	wrapped, err := WrapTokens(iterator, 10)
	if err != nil {
		t.Fatal(err)
	}
	if want := cellbuf.Wrap(source, 10, ""); wrapped.Text != want {
		t.Fatalf("wrapped source = %q, want %q", wrapped.Text, want)
	}

	for _, token := range wrapped.Iterator.Tokens() {
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

	wrapped, err := WrapTokens(chroma.Literator(sourceTokens...), 9)
	if err != nil {
		t.Fatal(err)
	}
	if wrapped.Text != wantWrapped {
		t.Fatalf("wrapped source = %q, want %q", wrapped.Text, wantWrapped)
	}

	output := wrapped.Iterator.Tokens()
	if got := joinTokenValues(output); got != wrapped.Text {
		t.Fatalf("projected token source = %q, want %q", got, wrapped.Text)
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

	wrapped, err := WrapTokens(chroma.Literator(sourceTokens...), 5)
	if err != nil {
		t.Fatal(err)
	}
	if got := joinTokenValues(wrapped.Iterator.Tokens()); got != wrapped.Text {
		t.Fatalf("projected token source = %q, want %q", got, wrapped.Text)
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
		{name: "unicode whitespace", source: "first second", width: 6},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iterator := chroma.Literator(chroma.Token{Type: chroma.Text, Value: test.source})
			wrapped, err := WrapTokens(iterator, test.width)
			if err != nil {
				t.Fatal(err)
			}
			if want := cellbuf.Wrap(test.source, test.width, ""); wrapped.Text != want {
				t.Fatalf("wrapped source = %q, want %q", wrapped.Text, want)
			}
			if got := joinTokenValues(wrapped.Iterator.Tokens()); got != wrapped.Text {
				t.Fatalf("projected source = %q, want %q", got, wrapped.Text)
			}
		})
	}
}

func TestWrapTokensMapsDisplayRowsToSourceLines(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		width       int
		wantText    string
		lineSources []int
	}{
		{
			name:        "no wrapping keeps one row per line",
			source:      "one\ntwo\nthree",
			width:       10,
			wantText:    "one\ntwo\nthree",
			lineSources: []int{0, 1, 2},
		},
		{
			name:        "soft wrapped rows continue their source line",
			source:      "abcdefghijk",
			width:       4,
			wantText:    "abcd\nefgh\nijk",
			lineSources: []int{0, Continuation, Continuation},
		},
		{
			name:        "source line after a wrapped line is numbered again",
			source:      "abcdefgh\nxy",
			width:       4,
			wantText:    "abcd\nefgh\nxy",
			lineSources: []int{0, Continuation, 1},
		},
		{
			name:        "wrap at a space then a source newline",
			source:      "alpha beta\ngamma",
			width:       5,
			wantText:    "alpha\nbeta\ngamma",
			lineSources: []int{0, Continuation, 1},
		},
		{
			name:        "trimmed trailing whitespace keeps the source newline",
			source:      "first   \nsecond",
			width:       6,
			wantText:    "first\nsecond",
			lineSources: []int{0, 1},
		},
		{
			name:        "blank source lines keep their own numbers",
			source:      "a\n\nb",
			width:       4,
			wantText:    "a\n\nb",
			lineSources: []int{0, 1, 2},
		},
		{
			name:        "trailing newline yields an empty final row",
			source:      "abcdef\n",
			width:       3,
			wantText:    "abc\ndef\n",
			lineSources: []int{0, Continuation, 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iterator := chroma.Literator(chroma.Token{Type: chroma.Text, Value: test.source})
			wrapped, err := WrapTokens(iterator, test.width)
			if err != nil {
				t.Fatal(err)
			}
			if wrapped.Text != test.wantText {
				t.Fatalf("wrapped source = %q, want %q", wrapped.Text, test.wantText)
			}
			if !slices.Equal(wrapped.LineSources, test.lineSources) {
				t.Fatalf("line sources = %v, want %v", wrapped.LineSources, test.lineSources)
			}
			if rows := strings.Count(wrapped.Text, "\n") + 1; rows != len(wrapped.LineSources) {
				t.Fatalf("line sources has %d entries for %d rows", len(wrapped.LineSources), rows)
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
