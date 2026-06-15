// Package render is the shared rendering core for blizzaga: it decorates a
// measured SVG content document with terminal-window chrome (background,
// padding/margin, window controls, shadow, border, corner radius) and
// rasterizes the result to PNG. The blizzaga CLI and external consumers (e.g.
// filament's capture tooling) build content documents and hand them here.
package render

const (
	defaultFontSize   = 14.0
	defaultLineHeight = 1.2

	// FontHeightToWidthRatio is the monospace cell aspect ratio used to derive
	// per-column advance from the font size (calibrated for JetBrains Mono).
	FontHeightToWidthRatio = 1.68
)

// Config holds the window-decoration options for a rendered image. Its JSON
// tags define the portable blizzaga configuration schema, shared verbatim with
// downstream consumers so configuration files are interchangeable.
type Config struct {
	Background      string    `json:"background"`
	Margin          []float64 `json:"margin"`
	Padding         []float64 `json:"padding"`
	Window          bool      `json:"window"`
	Width           float64   `json:"width"`
	Height          float64   `json:"height"`
	Border          Border    `json:"border"`
	Shadow          Shadow    `json:"shadow"`
	Font            Font      `json:"font"`
	LineHeight      float64   `json:"line_height"`
	ShowLineNumbers bool      `json:"show_line_numbers"`
}

// Shadow is the configuration for a drop shadow.
type Shadow struct {
	Blur float64 `json:"blur"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

// Border is the configuration for a window border.
type Border struct {
	Radius float64 `json:"radius"`
	Width  float64 `json:"width"`
	Color  string  `json:"color"`
}

// Font is the configuration for the rendered font.
type Font struct {
	Family    string  `json:"family"`
	File      string  `json:"file"`
	Size      float64 `json:"size"`
	Ligatures bool    `json:"ligatures"`
}

type side int

const (
	top    side = 0
	right  side = 1
	bottom side = 2
	left   side = 3
)

// ExpandPadding normalizes a 1/2/4-length spec into a [top, right, bottom, left]
// slice, scaled by the given factor.
func ExpandPadding(p []float64, scale float64) []float64 {
	switch len(p) {
	case 1:
		return []float64{p[top] * scale, p[top] * scale, p[top] * scale, p[top] * scale}
	case 2:
		return []float64{p[top] * scale, p[right] * scale, p[top] * scale, p[right] * scale}
	case 4:
		return []float64{p[top] * scale, p[right] * scale, p[bottom] * scale, p[left] * scale}
	default:
		return []float64{0, 0, 0, 0}
	}
}

// ExpandMargin normalizes a 1/2/4-length margin spec, scaled by the given factor.
var ExpandMargin = ExpandPadding

// PaddingTop returns the top padding (Padding must be expanded to length 4).
func (c Config) PaddingTop() float64 { return c.Padding[top] }

// PaddingRight returns the right padding (Padding must be expanded to length 4).
func (c Config) PaddingRight() float64 { return c.Padding[right] }

// PaddingBottom returns the bottom padding (Padding must be expanded to length 4).
func (c Config) PaddingBottom() float64 { return c.Padding[bottom] }

// PaddingLeft returns the left padding (Padding must be expanded to length 4).
func (c Config) PaddingLeft() float64 { return c.Padding[left] }

// MarginTop returns the top margin (Margin must be expanded to length 4).
func (c Config) MarginTop() float64 { return c.Margin[top] }

// MarginRight returns the right margin (Margin must be expanded to length 4).
func (c Config) MarginRight() float64 { return c.Margin[right] }

// MarginBottom returns the bottom margin (Margin must be expanded to length 4).
func (c Config) MarginBottom() float64 { return c.Margin[bottom] }

// MarginLeft returns the left margin (Margin must be expanded to length 4).
func (c Config) MarginLeft() float64 { return c.Margin[left] }
