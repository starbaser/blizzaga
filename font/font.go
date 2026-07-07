// Package font embeds the Iosevka Custom font assets used by Blizzaga.
//
// Iosevka Custom is a first-party build of Iosevka
// (https://github.com/be5invis/Iosevka), Copyright 2015-2026 Renzhi Li (aka.
// Belleve Invis) and Iosevka contributors, licensed under the SIL Open Font
// License 1.1 (https://openfontlicense.org).
//
// The embedded files are terminal-oriented subsets of the iosevka-eigenmage
// build (Latin/Greek/Cyrillic, punctuation, arrows, math, box drawing,
// blocks, geometric shapes, braille, powerline). The NL variant strips the
// calt/liga ligature features; both preserve the design metrics
// (0.5 em advance, 1.25 em line, 2.5 cell aspect).
package font

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"strings"
)

const (
	// NormalWeight is the CSS/OpenType weight for the regular bundled face.
	NormalWeight = "400"
	// BoldWeight is the CSS/OpenType weight for the bold bundled face.
	BoldWeight = "700"
	// NormalStyle is the CSS style for upright bundled faces.
	NormalStyle = "normal"
	// ItalicStyle is the CSS style for italic bundled faces.
	ItalicStyle = "italic"
)

// Face is one embedded Iosevka Custom font face.
type Face struct {
	Weight string
	Style  string
	TTF    []byte
	Base64 string
}

// IosevkaCustomTTF contains the embedded IosevkaCustom-Regular.ttf subset.
//
//go:embed IosevkaCustom-Regular.ttf
var IosevkaCustomTTF []byte

// IosevkaCustomItalicTTF contains the embedded IosevkaCustom-Italic.ttf subset.
//
//go:embed IosevkaCustom-Italic.ttf
var IosevkaCustomItalicTTF []byte

// IosevkaCustomBoldTTF contains the embedded IosevkaCustom-Bold.ttf subset.
//
//go:embed IosevkaCustom-Bold.ttf
var IosevkaCustomBoldTTF []byte

// IosevkaCustomBoldItalicTTF contains the embedded
// IosevkaCustom-BoldItalic.ttf subset.
//
//go:embed IosevkaCustom-BoldItalic.ttf
var IosevkaCustomBoldItalicTTF []byte

// IosevkaCustomNLTTF contains the embedded IosevkaCustomNL-Regular.ttf
// no-ligatures subset.
//
//go:embed IosevkaCustomNL-Regular.ttf
var IosevkaCustomNLTTF []byte

// IosevkaCustomNLItalicTTF contains the embedded IosevkaCustomNL-Italic.ttf
// no-ligatures subset.
//
//go:embed IosevkaCustomNL-Italic.ttf
var IosevkaCustomNLItalicTTF []byte

// IosevkaCustomNLBoldTTF contains the embedded IosevkaCustomNL-Bold.ttf
// no-ligatures subset.
//
//go:embed IosevkaCustomNL-Bold.ttf
var IosevkaCustomNLBoldTTF []byte

// IosevkaCustomNLBoldItalicTTF contains the embedded
// IosevkaCustomNL-BoldItalic.ttf no-ligatures subset.
//
//go:embed IosevkaCustomNL-BoldItalic.ttf
var IosevkaCustomNLBoldItalicTTF []byte

var (
	// IosevkaCustom font, base64-encoded for SVG @font-face embedding.
	IosevkaCustom = base64.StdEncoding.EncodeToString(IosevkaCustomTTF)
	// IosevkaCustomItalic font, base64-encoded for SVG @font-face embedding.
	IosevkaCustomItalic = base64.StdEncoding.EncodeToString(IosevkaCustomItalicTTF)
	// IosevkaCustomBold font, base64-encoded for SVG @font-face embedding.
	IosevkaCustomBold = base64.StdEncoding.EncodeToString(IosevkaCustomBoldTTF)
	// IosevkaCustomBoldItalic font, base64-encoded for SVG @font-face embedding.
	IosevkaCustomBoldItalic = base64.StdEncoding.EncodeToString(IosevkaCustomBoldItalicTTF)

	// IosevkaCustomNL font, base64-encoded for SVG @font-face embedding.
	IosevkaCustomNL = base64.StdEncoding.EncodeToString(IosevkaCustomNLTTF)
	// IosevkaCustomNLItalic font, base64-encoded for SVG @font-face embedding.
	IosevkaCustomNLItalic = base64.StdEncoding.EncodeToString(IosevkaCustomNLItalicTTF)
	// IosevkaCustomNLBold font, base64-encoded for SVG @font-face embedding.
	IosevkaCustomNLBold = base64.StdEncoding.EncodeToString(IosevkaCustomNLBoldTTF)
	// IosevkaCustomNLBoldItalic font, base64-encoded for SVG @font-face embedding.
	IosevkaCustomNLBoldItalic = base64.StdEncoding.EncodeToString(IosevkaCustomNLBoldItalicTTF)
)

// Faces returns the embedded face set matching the ligature preference.
func Faces(ligatures bool) []Face {
	if !ligatures {
		return []Face{
			{Weight: NormalWeight, Style: NormalStyle, TTF: IosevkaCustomNLTTF, Base64: IosevkaCustomNL},
			{Weight: NormalWeight, Style: ItalicStyle, TTF: IosevkaCustomNLItalicTTF, Base64: IosevkaCustomNLItalic},
			{Weight: BoldWeight, Style: NormalStyle, TTF: IosevkaCustomNLBoldTTF, Base64: IosevkaCustomNLBold},
			{Weight: BoldWeight, Style: ItalicStyle, TTF: IosevkaCustomNLBoldItalicTTF, Base64: IosevkaCustomNLBoldItalic},
		}
	}
	return []Face{
		{Weight: NormalWeight, Style: NormalStyle, TTF: IosevkaCustomTTF, Base64: IosevkaCustom},
		{Weight: NormalWeight, Style: ItalicStyle, TTF: IosevkaCustomItalicTTF, Base64: IosevkaCustomItalic},
		{Weight: BoldWeight, Style: NormalStyle, TTF: IosevkaCustomBoldTTF, Base64: IosevkaCustomBold},
		{Weight: BoldWeight, Style: ItalicStyle, TTF: IosevkaCustomBoldItalicTTF, Base64: IosevkaCustomBoldItalic},
	}
}

// FaceCSS emits the bundled font-face rules for a configured family name.
func FaceCSS(family string, ligatures bool) string {
	var b strings.Builder
	for _, face := range Faces(ligatures) {
		fmt.Fprintf(
			&b,
			`@font-face{font-family:"%s";src:url("data:application/x-font-truetype;charset=utf-8;base64,%s") format("truetype");font-weight:%s;font-style:%s;}`+"\n",
			family,
			face.Base64,
			face.Weight,
			face.Style,
		)
	}
	return b.String()
}
