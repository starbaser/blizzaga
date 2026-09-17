package highlight

import (
	_ "embed"
	"regexp"

	"github.com/alecthomas/chroma/v2/lexers"
	tree_sitter_baml "github.com/starbaser/tree-sitter-baml/bindings/go"
	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

//go:embed queries/go/highlights.scm
var goHighlights string

//go:embed queries/baml/highlights.scm
var bamlHighlights string

//go:embed queries/baml/injections.scm
var bamlInjections string

//go:embed queries/baml/template.scm
var bamlTemplateHighlights string

//go:embed queries/fml/highlights.scm
var fmlHighlights string

// BuiltinLanguages returns fresh descriptors for Blizzaga's bundled
// Tree-sitter languages.
func BuiltinLanguages() []Language {
	chromaGo := lexers.Get("go")
	goLanguage := Language{
		Name:       "Go",
		Aliases:    []string{"go", "golang"},
		Filenames:  []string{"*.go"},
		MIMETypes:  []string{"text/x-gosrc"},
		Grammar:    sitter.NewLanguage(tree_sitter_go.Language()),
		Highlights: goHighlights,
		Captures:   StandardCaptures(),
	}
	if chromaGo != nil {
		config := chromaGo.Config()
		goLanguage.Name = config.Name
		goLanguage.Aliases = append([]string(nil), config.Aliases...)
		goLanguage.Filenames = append([]string(nil), config.Filenames...)
		goLanguage.MIMETypes = append([]string(nil), config.MimeTypes...)
		goLanguage.Priority = config.Priority
		goLanguage.Analyse = chromaGo.AnalyseText
	}

	return []Language{
		goLanguage,
		{
			Name:       "BAML",
			Aliases:    []string{"baml"},
			Filenames:  []string{"*.baml"},
			MIMETypes:  []string{"text/x-baml"},
			Grammar:    sitter.NewLanguage(tree_sitter_baml.Language()),
			Highlights: bamlHighlights,
			Captures:   StandardCaptures(),
			Injections: []Injection{{
				Query:     bamlInjections,
				Languages: map[string]string{"fml": "fml", "baml-template": "baml-template"},
			}},
			Analyse: analyseBAML,
		},
		{
			Name:       "FML",
			Aliases:    []string{"fml"},
			Grammar:    sitter.NewLanguage(tree_sitter_baml.LanguageFML()),
			Highlights: fmlHighlights,
			Captures:   StandardCaptures(),
		},
		{
			// The FML grammar over a BAML raw string body. The template query
			// runs first so its string capture outranks FML's markup capture on
			// the same text nodes; everything else comes from the FML query.
			Name:       "BAML Template",
			Aliases:    []string{"baml-template"},
			Grammar:    sitter.NewLanguage(tree_sitter_baml.LanguageFML()),
			Highlights: bamlTemplateHighlights + "\n" + fmlHighlights,
			Captures:   StandardCaptures(),
		},
	}
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
