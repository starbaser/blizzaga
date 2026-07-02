// Iosevka Custom is a first-party build of Iosevka
// (https://github.com/be5invis/Iosevka), Copyright 2015-2026 Renzhi Li
// (aka. Belleve Invis) and Iosevka contributors, licensed under the SIL Open
// Font License 1.1 (https://openfontlicense.org).
//
// The embedded files are terminal-oriented subsets of the iosevka-eigenmage
// build (Latin/Greek/Cyrillic, punctuation, arrows, math, box drawing,
// blocks, geometric shapes, braille, powerline). The NL variant strips the
// calt/liga ligature features; both preserve the design metrics
// (0.5 em advance, 1.25 em line, 2.5 cell aspect).
package font //nolint:revive

import (
	_ "embed"
	"encoding/base64"
)

// IosevkaCustomTTF contains the embedded IosevkaCustom-Regular.ttf subset.
//
//go:embed IosevkaCustom-Regular.ttf
var IosevkaCustomTTF []byte

// IosevkaCustomNLTTF contains the embedded IosevkaCustomNL-Regular.ttf
// no-ligatures subset.
//
//go:embed IosevkaCustomNL-Regular.ttf
var IosevkaCustomNLTTF []byte

var (
	// IosevkaCustom font, base64-encoded for SVG @font-face embedding.
	IosevkaCustom = base64.StdEncoding.EncodeToString(IosevkaCustomTTF)

	// IosevkaCustomNL font, base64-encoded for SVG @font-face embedding.
	IosevkaCustomNL = base64.StdEncoding.EncodeToString(IosevkaCustomNLTTF)
)
