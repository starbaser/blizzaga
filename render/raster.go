package render

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/beevik/etree"
	"github.com/kanrichan/resvg-go"

	"github.com/starbaser/blizzaga/font"
)

// Rasterize converts an SVG document to PNG bytes. It prefers rsvg-convert
// (librsvg) when available for speed, falling back to the embedded resvg-go
// WASM rasterizer otherwise.
func Rasterize(doc *etree.Document, width, height float64) ([]byte, error) {
	svgBytes, err := doc.WriteToBytes()
	if err != nil {
		return nil, fmt.Errorf("serialize SVG: %w", err)
	}
	if png, err := rasterizeLibrsvg(svgBytes); err == nil {
		return png, nil
	}
	return rasterizeResvg(svgBytes, width, height)
}

func rasterizeLibrsvg(svgBytes []byte) ([]byte, error) {
	if _, err := exec.LookPath("rsvg-convert"); err != nil {
		return nil, err //nolint:wrapcheck
	}
	cmd := exec.Command("rsvg-convert")
	cmd.Stdin = bytes.NewReader(svgBytes)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("rsvg-convert: %w", err)
	}
	return out, nil
}

func rasterizeResvg(svgBytes []byte, width, height float64) ([]byte, error) {
	worker, err := resvg.NewDefaultWorker(context.Background())
	if err != nil {
		return nil, fmt.Errorf("resvg worker: %w", err)
	}
	defer worker.Close() //nolint:errcheck

	fontdb, err := worker.NewFontDBDefault()
	if err != nil {
		return nil, fmt.Errorf("resvg font db: %w", err)
	}
	defer fontdb.Close() //nolint:errcheck
	if err := fontdb.LoadFontData(font.JetBrainsMonoTTF); err != nil {
		return nil, fmt.Errorf("load font: %w", err)
	}
	if err := fontdb.LoadFontData(font.JetBrainsMonoNLTTF); err != nil {
		return nil, fmt.Errorf("load font: %w", err)
	}

	pixmap, err := worker.NewPixmap(uint32(width), uint32(height))
	if err != nil {
		return nil, fmt.Errorf("resvg pixmap: %w", err)
	}
	defer pixmap.Close() //nolint:errcheck

	tree, err := worker.NewTreeFromData(svgBytes, &resvg.Options{
		Dpi:                192,
		ShapeRenderingMode: resvg.ShapeRenderingModeGeometricPrecision,
		TextRenderingMode:  resvg.TextRenderingModeOptimizeLegibility,
		ImageRenderingMode: resvg.ImageRenderingModeOptimizeQuality,
		DefaultSizeWidth:   float32(width),
		DefaultSizeHeight:  float32(height),
	})
	if err != nil {
		return nil, fmt.Errorf("resvg parse: %w", err)
	}
	defer tree.Close() //nolint:errcheck

	if err := tree.ConvertText(fontdb); err != nil {
		return nil, fmt.Errorf("resvg convert text: %w", err)
	}
	if err := tree.Render(resvg.TransformIdentity(), pixmap); err != nil {
		return nil, fmt.Errorf("resvg render: %w", err)
	}
	png, err := pixmap.EncodePNG()
	if err != nil {
		return nil, fmt.Errorf("encode PNG: %w", err)
	}
	return png, nil
}
