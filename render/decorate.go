package render

import (
	"fmt"

	"github.com/beevik/etree"

	"github.com/starbaser/blizzaga/svg"
)

// DecorateParams carries the per-render inputs that are not part of Config.
type DecorateParams struct {
	// Doc is the content SVG document: a root element containing a terminal
	// background <rect> and a text <g> with one <text> per line.
	Doc *etree.Document

	// Scale multiplies all geometry (used for high-DPI PNG output).
	Scale float64

	// AutoWidth/AutoHeight indicate that Config.Width/Height were unset and the
	// dimension should be derived from content.
	AutoWidth  bool
	AutoHeight bool

	// IsAnsi clears the placeholder text of each line so colored <tspan>s can be
	// injected afterwards by the caller.
	IsAnsi bool

	// LongestLineCols is the display width (in cells) of the longest content
	// line, tab-expanded; used to size the terminal when AutoWidth is set.
	LongestLineCols int

	// OffsetLine is added to rendered line numbers (the first captured line).
	OffsetLine int

	// LineNumberColor is the fill color used for injected line numbers.
	LineNumberColor string
}

// Decorated is the result of Decorate: final image dimensions plus the text
// group and line elements the caller may post-process (e.g. ANSI tspan fill).
type Decorated struct {
	Width     float64
	Height    float64
	TextGroup *etree.Element
	TextLines []*etree.Element
	Scale     float64
}

// Decorate wraps a measured content document in terminal-window chrome. It
// mutates cfg in place (expanding padding for window controls and scaling line
// height) so callers that post-process content see consistent geometry.
func Decorate(cfg *Config, p DecorateParams) (Decorated, error) {
	elements := p.Doc.ChildElements()
	if len(elements) < 1 {
		return Decorated{}, fmt.Errorf("content document has no root element")
	}
	image := elements[0]
	scale := p.Scale

	hPadding := cfg.Padding[left] + cfg.Padding[right]
	hMargin := cfg.Margin[left] + cfg.Margin[right]
	vMargin := cfg.Margin[top] + cfg.Margin[bottom]
	vPadding := cfg.Padding[top] + cfg.Padding[bottom]

	terminal := image.SelectElement("rect")

	w, h := svg.GetDimensions(image)
	imageWidth := float64(w)
	imageHeight := float64(h)

	imageWidth *= scale
	imageHeight *= scale

	// chroma automatically calculates the height based on a font size of 14
	// and a line height of 1.2
	imageHeight *= (cfg.Font.Size / defaultFontSize)
	imageHeight *= (cfg.LineHeight / defaultLineHeight)

	terminalWidth := imageWidth
	terminalHeight := imageHeight

	if !p.AutoWidth {
		imageWidth = cfg.Width
		terminalWidth = cfg.Width - hMargin
	} else {
		imageWidth += hMargin + hPadding
		terminalWidth += hPadding
	}

	if !p.AutoHeight {
		imageHeight = cfg.Height
		terminalHeight = cfg.Height - vMargin
	} else {
		imageHeight += vMargin + vPadding
		terminalHeight += vPadding
	}

	if cfg.Window {
		windowControls := svg.NewWindowControls(5.5*float64(scale), 19.0*scale, 12.0*scale)
		svg.Move(windowControls, float64(cfg.Margin[left]), float64(cfg.Margin[top]))
		image.AddChild(windowControls)
		cfg.Padding[top] += (15 * scale)
	}

	if cfg.Border.Radius > 0 {
		svg.AddCornerRadius(terminal, cfg.Border.Radius*scale)
	}

	if cfg.Shadow.Blur > 0 || cfg.Shadow.X > 0 || cfg.Shadow.Y > 0 {
		id := "shadow"
		svg.AddShadow(image, id, cfg.Shadow.X*scale, cfg.Shadow.Y*scale, cfg.Shadow.Blur*scale)
		terminal.CreateAttr("filter", fmt.Sprintf("url(#%s)", id))
	}

	textGroup := image.SelectElement("g")
	textGroup.CreateAttr("font-size", fmt.Sprintf("%.2fpx", cfg.Font.Size*float64(scale)))
	textGroup.CreateAttr("clip-path", "url(#terminalMask)")
	text := textGroup.SelectElements("text")

	cfg.LineHeight *= float64(scale)

	for i, line := range text {
		if p.IsAnsi {
			line.SetText("")
		}
		if cfg.ShowLineNumbers {
			ln := etree.NewElement("tspan")
			ln.CreateAttr("xml:space", "preserve")
			ln.CreateAttr("fill", p.LineNumberColor)
			ln.SetText(fmt.Sprintf("%3d  ", i+1+p.OffsetLine))
			line.InsertChildAt(0, ln)
		}
		x := float64(cfg.Padding[left] + cfg.Margin[left])
		y := (float64(i+1))*(cfg.Font.Size*cfg.LineHeight) + float64(cfg.Padding[top]) + float64(cfg.Margin[top])

		svg.Move(line, x, y)

		// We are passed visible lines, remove the rest.
		if y > float64(imageHeight-cfg.Margin[bottom]-cfg.Padding[bottom]) {
			textGroup.RemoveChild(line)
		}
	}

	if p.AutoWidth {
		colAdvance, err := ColAdvance(*cfg)
		if err != nil {
			return Decorated{}, err
		}
		terminalWidth = float64(p.LongestLineCols+1) * colAdvance
		terminalWidth *= scale
		terminalWidth += hPadding
		imageWidth = terminalWidth + hMargin
	}

	if cfg.Border.Width > 0 {
		svg.AddOutline(terminal, cfg.Border.Width, cfg.Border.Color)

		// NOTE: necessary so that we don't clip the outline.
		terminalHeight -= (cfg.Border.Width * 2)
		terminalWidth -= (cfg.Border.Width * 2)
	}

	if cfg.ShowLineNumbers {
		if p.AutoWidth {
			terminalWidth += cfg.Font.Size * 3 * scale
			imageWidth += cfg.Font.Size * 3 * scale
		} else {
			terminalWidth -= cfg.Font.Size * 3
		}
	}

	if !p.AutoHeight || !p.AutoWidth {
		svg.AddClipPath(image, "terminalMask",
			cfg.Margin[left], cfg.Margin[top],
			terminalWidth, terminalHeight-cfg.Padding[bottom])
	}

	svg.Move(terminal, max(float64(cfg.Margin[left]), float64(cfg.Border.Width)/2), max(float64(cfg.Margin[top]), float64(cfg.Border.Width)/2))
	svg.SetDimensions(image, imageWidth, imageHeight)
	svg.SetDimensions(terminal, terminalWidth, terminalHeight)

	return Decorated{
		Width:     imageWidth,
		Height:    imageHeight,
		TextGroup: textGroup,
		TextLines: text,
		Scale:     scale,
	}, nil
}
