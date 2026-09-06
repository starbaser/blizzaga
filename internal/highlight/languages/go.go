// Package languages contains Blizzaga's statically linked Tree-sitter grammar
// registrations and their highlight queries.
package languages

import (
	_ "embed"

	"github.com/alecthomas/chroma/v2"
	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"

	"github.com/starbaser/blizzaga/internal/highlight/chromalexer"
	tshighlight "github.com/starbaser/blizzaga/internal/highlight/treesitter"
)

//go:embed queries/go/highlights.scm
var goHighlights string

// Go constructs the Tree-sitter-backed Go lexer with Chroma's existing
// registry metadata.
func Go(config chroma.Config) (*chromalexer.Lexer, error) {
	language := sitter.NewLanguage(tree_sitter_go.Language())
	engine, err := tshighlight.NewEngine(language, goHighlights)
	if err != nil {
		return nil, err
	}
	return chromalexer.New(config, engine, chromalexer.StandardCaptures())
}
