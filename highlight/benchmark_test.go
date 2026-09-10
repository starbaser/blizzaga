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

func benchmarkFMLSource(member string) string {
	var source strings.Builder
	source.WriteString("class Props { label string title string }\n")
	source.WriteString("function View(props: Props) -> filament.Template {\n  return ##\"")
	for range 100 {
		source.WriteString("<text content=\"Unicode ✦ {{ props.")
		source.WriteString(member)
		source.WriteString(" }}\" />\n")
	}
	source.WriteString("\"##\n}\n")
	return source.String()
}

func BenchmarkFMLHighlightFresh(b *testing.B) {
	registry, err := DefaultRegistry()
	if err != nil {
		b.Fatal(err)
	}
	sources := [2]string{benchmarkFMLSource("label"), benchmarkFMLSource("title")}
	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		if _, err := registry.Highlight(sources[i%2], Query{Language: "baml"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFMLSessionApplyEdit(b *testing.B) {
	registry, err := DefaultRegistry()
	if err != nil {
		b.Fatal(err)
	}
	source := benchmarkFMLSource("label")
	offset := strings.Index(source, "props.label") + len("props.")
	session, err := registry.NewSession(source, Query{Language: "baml"})
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = session.Close() })
	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		replacement := "title"
		if i%2 == 1 {
			replacement = "label"
		}
		if _, err := session.ApplyEdit(Edit{
			StartByte:  offset,
			OldEndByte: offset + len(replacement),
			NewText:    replacement,
		}); err != nil {
			b.Fatal(err)
		}
	}
}
