package highlight

import (
	"fmt"

	"github.com/alecthomas/chroma/v2/lexers"

	"github.com/starbaser/blizzaga/internal/highlight/chromalexer"
	"github.com/starbaser/blizzaga/internal/highlight/languages"
)

// RegisterTreeSitterLexers installs the statically linked Tree-sitter lexers
// before Blizzaga performs any Chroma registry lookups.
func RegisterTreeSitterLexers() error {
	chromaGo := lexers.Get("go")
	if chromaGo == nil {
		return fmt.Errorf("Chroma Go lexer is unavailable")
	}
	if _, registered := chromaGo.(*chromalexer.Lexer); registered {
		return nil
	}

	goLexer, err := languages.Go(*chromaGo.Config())
	if err != nil {
		return fmt.Errorf("register Go Tree-sitter lexer: %w", err)
	}
	goLexer.SetAnalyser(chromaGo.AnalyseText)
	lexers.Register(goLexer)

	return nil
}
