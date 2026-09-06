package highlight

import (
	"fmt"
	"regexp"

	"github.com/alecthomas/chroma/v2/lexers"

	"github.com/starbaser/blizzaga/internal/highlight/chromalexer"
	"github.com/starbaser/blizzaga/internal/highlight/languages"
)

// RegisterTreeSitterLexers installs the statically linked Tree-sitter lexers
// before Blizzaga performs any Chroma registry lookups.
func RegisterTreeSitterLexers() error {
	if err := registerGo(); err != nil {
		return err
	}
	if err := registerBAML(); err != nil {
		return err
	}
	return nil
}

func registerGo() error {
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

func registerBAML() error {
	existing := lexers.Get("baml")
	if _, registered := existing.(*chromalexer.Lexer); registered {
		return nil
	}

	config := languages.BAMLConfig()
	if existing != nil {
		config = *existing.Config()
	}
	bamlLexer, err := languages.BAML(config)
	if err != nil {
		return fmt.Errorf("register BAML Tree-sitter lexer: %w", err)
	}
	if existing != nil {
		bamlLexer.SetAnalyser(existing.AnalyseText)
	} else {
		bamlLexer.SetAnalyser(analyseBAML)
	}
	lexers.Register(bamlLexer)
	return nil
}

var bamlDeclaration = regexp.MustCompile(`(?m)^\s*(?:class|enum|function|template_string|client|generator|retry_policy|printer|test|type_builder)\s+[A-Za-z_]`)

func analyseBAML(source string) float32 {
	declarations := bamlDeclaration.FindAllStringIndex(source, 3)
	switch {
	case len(declarations) >= 2:
		return 0.8
	case len(declarations) == 1:
		return 0.4
	default:
		return 0
	}
}
