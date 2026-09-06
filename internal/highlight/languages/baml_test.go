package languages

import (
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
)

func TestBAMLLexerUsesStructuralCaptures(t *testing.T) {
	lexer, err := BAML(BAMLConfig())
	if err != nil {
		t.Fatal(err)
	}

	source := "function Center(ctx: training.numeric.Context, example: Example) -> training.numeric.Program {\n  let x = example.target.x;\n  let scale = 1024.0;\n  ctx.finish(x)\n}\n"
	iterator, err := lexer.Tokenise(nil, source)
	if err != nil {
		t.Fatal(err)
	}
	tokens := iterator.Tokens()

	wants := []chroma.Token{
		{Type: chroma.Keyword, Value: "function"},
		{Type: chroma.NameFunction, Value: "Center"},
		{Type: chroma.NameClass, Value: "Context"},
		{Type: chroma.NameProperty, Value: "target"},
		{Type: chroma.NameProperty, Value: "x"},
		{Type: chroma.NameFunction, Value: "finish"},
		{Type: chroma.LiteralNumber, Value: "1024.0"},
	}
	for _, want := range wants {
		if !containsToken(tokens, want) {
			t.Errorf("missing token %#v in %#v", want, tokens)
		}
	}

	var roundTrip strings.Builder
	for _, token := range tokens {
		roundTrip.WriteString(token.Value)
	}
	if roundTrip.String() != source {
		t.Fatalf("token stream did not preserve source\ngot:  %q\nwant: %q", roundTrip.String(), source)
	}
}
