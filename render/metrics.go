package render

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"

	blizzagafont "github.com/starbaser/blizzaga/font"
)

// Metrics are the design metrics of the font a render embeds, expressed as
// fractions of the em size so they scale with any configured font size.
type Metrics struct {
	// AdvanceEm is the monospace column advance per em ("0" glyph).
	AdvanceEm float64
	// LineEm is the design line pitch per em (ascent - descent + line gap).
	LineEm float64
}

// CellAspect is the height/width ratio of the font's terminal cell.
func (m Metrics) CellAspect() float64 {
	return m.LineEm / m.AdvanceEm
}

var (
	metricsMu    sync.Mutex
	metricsCache = map[string]Metrics{}
)

// FontMetrics parses the font the config actually embeds — Font.File when
// set, the bundled font otherwise — and returns its design metrics. Results
// are cached per font file.
func FontMetrics(cfg Config) (Metrics, error) {
	key := cfg.Font.File
	metricsMu.Lock()
	defer metricsMu.Unlock()
	if m, ok := metricsCache[key]; ok {
		return m, nil
	}
	data := blizzagafont.IosevkaCustomTTF
	if cfg.Font.File != "" {
		raw, err := os.ReadFile(cfg.Font.File)
		if err != nil {
			return Metrics{}, fmt.Errorf("read font file: %w", err)
		}
		data = raw
	}
	m, err := parseMetrics(data)
	if err != nil {
		return Metrics{}, err
	}
	metricsCache[key] = m
	return m, nil
}

// ColAdvance is the per-column pixel advance for the configured font.
func ColAdvance(cfg Config) (float64, error) {
	m, err := FontMetrics(cfg)
	if err != nil {
		return 0, err
	}
	return cfg.Font.Size * m.AdvanceEm, nil
}

func parseMetrics(data []byte) (Metrics, error) {
	f, err := sfnt.Parse(data)
	if err != nil {
		return Metrics{}, fmt.Errorf("parse font: %w", err)
	}
	var buf sfnt.Buffer
	upm := fixed.Int26_6(f.UnitsPerEm()) << 6
	glyph, err := f.GlyphIndex(&buf, '0')
	if err != nil || glyph == 0 {
		return Metrics{}, fmt.Errorf("font has no '0' glyph: %w", err)
	}
	advance, err := f.GlyphAdvance(&buf, glyph, upm, font.HintingNone)
	if err != nil {
		return Metrics{}, fmt.Errorf("glyph advance: %w", err)
	}
	metrics, err := f.Metrics(&buf, upm, font.HintingNone)
	if err != nil {
		return Metrics{}, fmt.Errorf("font metrics: %w", err)
	}
	em := float64(upm)
	return Metrics{
		AdvanceEm: float64(advance) / em,
		LineEm:    float64(metrics.Height) / em,
	}, nil
}
