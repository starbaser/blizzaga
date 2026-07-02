package render

import "testing"

func TestEmbeddedFontMetrics(t *testing.T) {
	m, err := FontMetrics(Config{})
	if err != nil {
		t.Fatalf("FontMetrics: %v", err)
	}
	if m.AdvanceEm != 0.5 {
		t.Fatalf("advance = %v em, want Iosevka's designed 0.5", m.AdvanceEm)
	}
	if m.LineEm != 1.25 {
		t.Fatalf("line = %v em, want Iosevka's designed 1.25", m.LineEm)
	}
	if m.CellAspect() != 2.5 {
		t.Fatalf("cell aspect = %v, want 2.5", m.CellAspect())
	}
	advance, err := ColAdvance(Config{Font: Font{Size: 14}})
	if err != nil {
		t.Fatalf("ColAdvance: %v", err)
	}
	if advance != 7.0 {
		t.Fatalf("col advance at 14px = %v, want 7.0", advance)
	}
}
