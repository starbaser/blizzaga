package main

import (
	"fmt"
	"strings"

	"charm.land/glamour/v2"
	glamouransi "charm.land/glamour/v2/ansi"
	glamourstyles "charm.land/glamour/v2/styles"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
)

const markdownChromaFormatter = "terminal16m"

func isMarkdownLanguage(language string) bool {
	switch strings.ToLower(language) {
	case "markdown", "md":
		return true
	default:
		return false
	}
}

func renderMarkdown(source string, theme string, wrap int) (string, error) {
	options := []glamour.TermRendererOption{
		glamour.WithStyles(markdownStyle(theme)),
		glamour.WithChromaFormatter(markdownChromaFormatter),
	}
	if wrap > 0 {
		options = append(options, glamour.WithWordWrap(wrap))
	}

	renderer, err := glamour.NewTermRenderer(options...)
	if err != nil {
		return "", fmt.Errorf("create markdown renderer: %w", err)
	}
	rendered, err := renderer.Render(source)
	if err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	rendered = strings.TrimSuffix(rendered, "\n")
	rendered = strings.ReplaceAll(rendered, "\t", "    ")
	return rendered, nil
}

func markdownStyle(theme string) glamouransi.StyleConfig {
	name := strings.ToLower(theme)
	if name == "" || name == "srcery" {
		return srceryMarkdownStyle()
	}
	if style, ok := glamourstyles.DefaultStyles[name]; ok {
		cfg := *style
		normalizeGlamourCodeBlockTheme(&cfg, "blizzaga-glamour-"+name)
		return cfg
	}
	return srceryMarkdownStyle()
}

func normalizeGlamourCodeBlockTheme(cfg *glamouransi.StyleConfig, name string) {
	if cfg == nil || cfg.CodeBlock.Chroma == nil {
		return
	}
	registerGlamourChromaStyle(name, cfg.CodeBlock.Chroma)
	cfg.CodeBlock.Theme = name
	cfg.CodeBlock.Chroma = nil
}

func registerGlamourChromaStyle(name string, cfg *glamouransi.Chroma) {
	if _, ok := styles.Registry[name]; ok {
		return
	}
	styles.Register(chroma.MustNewStyle(name, chroma.StyleEntries{
		chroma.Text:                glamourChromaStyle(cfg.Text),
		chroma.Error:               glamourChromaStyle(cfg.Error),
		chroma.Comment:             glamourChromaStyle(cfg.Comment),
		chroma.CommentPreproc:      glamourChromaStyle(cfg.CommentPreproc),
		chroma.Keyword:             glamourChromaStyle(cfg.Keyword),
		chroma.KeywordReserved:     glamourChromaStyle(cfg.KeywordReserved),
		chroma.KeywordNamespace:    glamourChromaStyle(cfg.KeywordNamespace),
		chroma.KeywordType:         glamourChromaStyle(cfg.KeywordType),
		chroma.Operator:            glamourChromaStyle(cfg.Operator),
		chroma.Punctuation:         glamourChromaStyle(cfg.Punctuation),
		chroma.Name:                glamourChromaStyle(cfg.Name),
		chroma.NameBuiltin:         glamourChromaStyle(cfg.NameBuiltin),
		chroma.NameTag:             glamourChromaStyle(cfg.NameTag),
		chroma.NameAttribute:       glamourChromaStyle(cfg.NameAttribute),
		chroma.NameClass:           glamourChromaStyle(cfg.NameClass),
		chroma.NameConstant:        glamourChromaStyle(cfg.NameConstant),
		chroma.NameDecorator:       glamourChromaStyle(cfg.NameDecorator),
		chroma.NameException:       glamourChromaStyle(cfg.NameException),
		chroma.NameFunction:        glamourChromaStyle(cfg.NameFunction),
		chroma.NameOther:           glamourChromaStyle(cfg.NameOther),
		chroma.Literal:             glamourChromaStyle(cfg.Literal),
		chroma.LiteralNumber:       glamourChromaStyle(cfg.LiteralNumber),
		chroma.LiteralDate:         glamourChromaStyle(cfg.LiteralDate),
		chroma.LiteralString:       glamourChromaStyle(cfg.LiteralString),
		chroma.LiteralStringEscape: glamourChromaStyle(cfg.LiteralStringEscape),
		chroma.GenericDeleted:      glamourChromaStyle(cfg.GenericDeleted),
		chroma.GenericEmph:         glamourChromaStyle(cfg.GenericEmph),
		chroma.GenericInserted:     glamourChromaStyle(cfg.GenericInserted),
		chroma.GenericStrong:       glamourChromaStyle(cfg.GenericStrong),
		chroma.GenericSubheading:   glamourChromaStyle(cfg.GenericSubheading),
		chroma.Background:          glamourChromaStyle(cfg.Background),
	}))
}

func glamourChromaStyle(style glamouransi.StylePrimitive) string {
	var parts []string
	if style.Color != nil {
		parts = append(parts, *style.Color)
	}
	if style.BackgroundColor != nil {
		parts = append(parts, "bg:"+*style.BackgroundColor)
	}
	if style.Italic != nil && *style.Italic {
		parts = append(parts, "italic")
	}
	if style.Bold != nil && *style.Bold {
		parts = append(parts, "bold")
	}
	if style.Underline != nil && *style.Underline {
		parts = append(parts, "underline")
	}
	return strings.Join(parts, " ")
}

func srceryMarkdownStyle() glamouransi.StyleConfig {
	cfg := glamourstyles.DarkStyleConfig
	cfg.Document.Color = stringPtr(srceryBrightWhite)
	cfg.Document.BackgroundColor = nil
	cfg.Document.Margin = uintPtr(1)
	cfg.BlockQuote.Color = stringPtr(srceryWhite)
	cfg.Heading.Color = stringPtr(srceryBrightBlue)
	cfg.H1.Color = stringPtr(srceryBlack)
	cfg.H1.BackgroundColor = stringPtr(srceryBrightYellow)
	cfg.H6.Color = stringPtr(srceryBrightBlack)
	cfg.Emph.Color = stringPtr(srceryWhite)
	cfg.Strong.Color = stringPtr(srceryBrightYellow)
	cfg.HorizontalRule.Color = stringPtr(srceryBrightBlack)
	cfg.Link.Color = stringPtr(srceryBrightCyan)
	cfg.LinkText.Color = stringPtr(srceryBrightCyan)
	cfg.Image.Color = stringPtr(srceryBrightCyan)
	cfg.ImageText.Color = stringPtr(srceryBrightBlack)
	cfg.Code.Color = stringPtr(srceryBrightMagenta)
	cfg.Code.BackgroundColor = stringPtr(srceryGray1)
	cfg.CodeBlock.Color = stringPtr(srceryBrightWhite)
	cfg.CodeBlock.BackgroundColor = stringPtr(srceryBlack)
	cfg.CodeBlock.Theme = "srcery"
	cfg.CodeBlock.Chroma = nil
	return cfg
}

func stringPtr(s string) *string { return &s }
func uintPtr(n uint) *uint       { return &n }
