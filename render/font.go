package render

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/beevik/etree"

	"github.com/starbaser/blizzaga/font"
)

// DefaultFontFamily is the bundled monospace font family.
const DefaultFontFamily = "Iosevka Custom"

// EmbedFont injects an @font-face definition into the SVG root's <defs> and sets
// the root font-family, so a content document renders with the configured
// monospace font without relying on system fonts. It is the seam for content
// built outside chroma (e.g. terminal-grid snapshots); the blizzaga CLI embeds
// fonts through chroma's formatter instead.
func EmbedFont(root *etree.Element, cfg Config) error {
	family := cfg.Font.Family
	if family == "" {
		family = DefaultFontFamily
	}

	css, err := fontFaceCSS(family, cfg)
	if err != nil {
		return err
	}

	style := etree.NewElement("style")
	style.SetText(css)
	defs := etree.NewElement("defs")
	defs.AddChild(style)
	root.AddChild(defs)
	root.CreateAttr("font-family", family)
	return nil
}

func fontFaceCSS(family string, cfg Config) (string, error) {
	if cfg.Font.File != "" {
		raw, readErr := os.ReadFile(cfg.Font.File)
		if readErr != nil {
			return "", fmt.Errorf("read font file: %w", readErr)
		}
		var mime string
		switch filepath.Ext(cfg.Font.File) {
		case ".woff2":
			mime = "font/woff2"
		case ".woff":
			mime = "font/woff"
		case ".otf":
			mime = "font/otf"
		default:
			mime = "font/ttf"
		}
		return fmt.Sprintf(
			`@font-face{font-family:"%s";src:url("data:%s;base64,%s");font-weight:%s;font-style:%s;}`,
			family,
			mime,
			base64.StdEncoding.EncodeToString(raw),
			font.NormalWeight,
			font.NormalStyle,
		), nil
	}

	return font.FaceCSS(family, cfg.Font.Ligatures), nil
}
