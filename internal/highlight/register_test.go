package highlight_test

import (
	"testing"

	"github.com/alecthomas/chroma/v2/lexers"

	"github.com/starbaser/blizzaga/internal/highlight"
	"github.com/starbaser/blizzaga/internal/highlight/chromalexer"
)

func TestRegisterTreeSitterLexersReplacesGoAndPreservesAnalysis(t *testing.T) {
	original := lexers.Get("go")
	if original == nil {
		t.Fatal("Chroma Go lexer is unavailable")
	}
	source := "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println() }\n"
	wantScore := original.AnalyseText(source)

	if err := highlight.RegisterTreeSitterLexers(); err != nil {
		t.Fatal(err)
	}
	replacement := lexers.Get("go")
	if _, ok := replacement.(*chromalexer.Lexer); !ok {
		t.Fatalf("registered Go lexer type = %T, want *chromalexer.Lexer", replacement)
	}
	if got := replacement.AnalyseText(source); got != wantScore {
		t.Fatalf("analysis score = %v, want %v", got, wantScore)
	}
	if lexers.Get("example.go") != replacement {
		t.Fatal("filename lookup did not resolve to the Tree-sitter Go lexer")
	}
	if err := highlight.RegisterTreeSitterLexers(); err != nil {
		t.Fatalf("idempotent registration failed: %v", err)
	}
	if lexers.Get("go") != replacement {
		t.Fatal("idempotent registration replaced the Tree-sitter Go lexer")
	}
}
