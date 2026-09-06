package languages

import (
	_ "embed"

	"github.com/alecthomas/chroma/v2"
	tree_sitter_baml "github.com/starbaser/tree-sitter-baml/bindings/go"
	sitter "github.com/tree-sitter/go-tree-sitter"

	"github.com/starbaser/blizzaga/internal/highlight/chromalexer"
	tshighlight "github.com/starbaser/blizzaga/internal/highlight/treesitter"
)

//go:embed queries/baml/highlights.scm
var bamlHighlights string

// BAMLConfig describes the statically linked BAML grammar to Chroma's lookup
// registry.
func BAMLConfig() chroma.Config {
	return chroma.Config{
		Name:      "BAML",
		Aliases:   []string{"baml"},
		Filenames: []string{"*.baml"},
		MimeTypes: []string{"text/x-baml"},
	}
}

// BAML constructs a Tree-sitter-backed lexer using Alloy's BAML grammar.
func BAML(config chroma.Config) (*chromalexer.Lexer, error) {
	language := sitter.NewLanguage(tree_sitter_baml.Language())
	engine, err := tshighlight.NewEngine(language, bamlHighlights)
	if err != nil {
		return nil, err
	}
	return chromalexer.New(config, engine, chromalexer.StandardCaptures())
}
