package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/chroma/v2/formatters/svg"
	"github.com/starbaser/blizzaga/font"
	"github.com/starbaser/blizzaga/render"
)

func fontOptions(config *Config) ([]svg.Option, error) {
	if config.Font.File != "" {
		bts, err := os.ReadFile(config.Font.File)
		if err != nil {
			return nil, fmt.Errorf("invalid font file: %w", err)
		}

		var format svg.FontFormat
		switch ext := filepath.Ext(config.Font.File); ext {
		case ".ttf":
			format = svg.TRUETYPE
		case ".woff2":
			format = svg.WOFF2
		case ".woff":
			format = svg.WOFF
		default:
			return nil, fmt.Errorf("%s is not a supported font extension", ext)
		}

		return []svg.Option{
			svg.EmbedFont(
				config.Font.Family,
				base64.StdEncoding.EncodeToString(bts),
				format,
			),
			svg.FontFamily(config.Font.Family),
		}, nil
	}
	if config.Font.Family != render.DefaultFontFamily {
		return []svg.Option{
			svg.FontFamily(config.Font.Family),
		}, nil
	}
	fontBase64 := font.IosevkaCustom
	if !config.Font.Ligatures {
		fontBase64 = font.IosevkaCustomNL
	}
	return []svg.Option{
		svg.EmbedFont(config.Font.Family, fontBase64, svg.TRUETYPE),
		svg.FontFamily(config.Font.Family),
	}, nil
}
