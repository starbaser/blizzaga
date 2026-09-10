package highlight

import (
	"strings"
	"testing"
)

func benchmarkGoSource() string {
	var source strings.Builder
	source.WriteString("package benchmark\n\n")
	for range 200 {
		source.WriteString("func render() { println(\"hello, 世界\") }\n")
	}
	return source.String()
}

func BenchmarkHighlightOneShot(b *testing.B) {
	registry, err := DefaultRegistry()
	if err != nil {
		b.Fatal(err)
	}
	source := benchmarkGoSource()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := registry.Highlight(source, Query{Language: "go"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSessionResult(b *testing.B) {
	registry, err := DefaultRegistry()
	if err != nil {
		b.Fatal(err)
	}
	session, err := registry.NewSession(benchmarkGoSource(), Query{Language: "go"})
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = session.Close() })
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := session.Result(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSessionApplyEdit(b *testing.B) {
	registry, err := DefaultRegistry()
	if err != nil {
		b.Fatal(err)
	}
	source := benchmarkGoSource()
	offset := strings.Index(source, "render") + 1
	session, err := registry.NewSession(source, Query{Language: "go"})
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = session.Close() })
	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		replacement := "a"
		if i%2 == 1 {
			replacement = "e"
		}
		if _, err := session.ApplyEdit(Edit{
			StartByte:  offset,
			OldEndByte: offset + 1,
			NewText:    replacement,
		}); err != nil {
			b.Fatal(err)
		}
	}
}
