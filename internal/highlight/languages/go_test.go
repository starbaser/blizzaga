package languages

import (
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
)

func TestGoLexerUsesStructuralCaptures(t *testing.T) {
	lexer, err := Go(chroma.Config{
		Name:      "Go",
		Aliases:   []string{"go", "golang"},
		Filenames: []string{"*.go"},
		MimeTypes: []string{"text/x-gosrc"},
	})
	if err != nil {
		t.Fatal(err)
	}

	source := "package main //nolint:revive\n\ntype Thing struct { Value string }\n\nfunc render(t Thing) {\n\tfmt.Println(t.Value)\n\tprintln(\"hi\\n\")\n}\n"
	iterator, err := lexer.Tokenise(nil, source)
	if err != nil {
		t.Fatal(err)
	}
	tokens := iterator.Tokens()

	wants := []chroma.Token{
		{Type: chroma.KeywordNamespace, Value: "package"},
		{Type: chroma.KeywordDeclaration, Value: "func"},
		{Type: chroma.CommentPreproc, Value: "//nolint:revive"},
		{Type: chroma.NameFunction, Value: "render"},
		{Type: chroma.NameFunction, Value: "Println"},
		{Type: chroma.NameProperty, Value: "Value"},
		{Type: chroma.NameBuiltin, Value: "println"},
		{Type: chroma.NameClass, Value: "Thing"},
		{Type: chroma.LiteralStringEscape, Value: `\n`},
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

func TestGoLexerNormalizesLineEndingsByDefault(t *testing.T) {
	lexer, err := Go(chroma.Config{Name: "Go"})
	if err != nil {
		t.Fatal(err)
	}
	iterator, err := lexer.Tokenise(nil, "package main\r\n")
	if err != nil {
		t.Fatal(err)
	}

	var output strings.Builder
	for _, token := range iterator.Tokens() {
		output.WriteString(token.Value)
	}
	if output.String() != "package main\n" {
		t.Fatalf("normalized source = %q", output.String())
	}
}

func containsToken(tokens []chroma.Token, want chroma.Token) bool {
	for _, token := range tokens {
		if token == want {
			return true
		}
	}
	return false
}
