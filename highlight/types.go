package highlight

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Query describes how a registry should select a language. Non-empty fields
// are considered in Language, Filename, MIMEType order before content analysis.
type Query struct {
	Language string
	Filename string
	MIMEType string
}

// Span classifies one non-empty byte range in Result.Source. Result spans are
// ordered, non-overlapping, and cover the source exactly.
type Span struct {
	StartByte    int
	EndByte      int
	TokenType    chroma.TokenType
	ChromaClass  string
	CaptureClass string
}

// Result is an immutable highlighting snapshot. Source is the exact input,
// including line endings and trailing newlines.
type Result struct {
	Source   string
	Language string
	Spans    []Span
}

// Tokens projects a result into Chroma tokens without changing its source.
func (r Result) Tokens() []chroma.Token {
	tokens := make([]chroma.Token, 0, len(r.Spans))
	for _, span := range r.Spans {
		if span.StartByte < 0 || span.StartByte >= span.EndByte || span.EndByte > len(r.Source) {
			continue
		}
		value := r.Source[span.StartByte:span.EndByte]
		if len(tokens) > 0 && tokens[len(tokens)-1].Type == span.TokenType {
			tokens[len(tokens)-1].Value += value
			continue
		}
		tokens = append(tokens, chroma.Token{Type: span.TokenType, Value: value})
	}
	return tokens
}

// Edit replaces [StartByte, OldEndByte) with NewText. Offsets address the
// session's current UTF-8 source and must fall on rune boundaries.
type Edit struct {
	StartByte  int
	OldEndByte int
	NewText    string
}

// CaptureMapping assigns a neutral Tree-sitter capture to a Chroma token
// classification. Priority breaks ties between captures of identical range.
type CaptureMapping struct {
	TokenType chroma.TokenType
	Priority  int
}

// CaptureMap maps Tree-sitter highlight captures to token classifications.
// Dotted captures fall back through parent names when no exact mapping exists.
type CaptureMap map[string]CaptureMapping

// Injection declares a Tree-sitter injection query. Each matched
// @injection.content pattern must set injection.language; Languages translates
// that selector to a registered Tree-sitter language name or alias.
type Injection struct {
	Query     string
	Languages map[string]string
}

// Language declares one statically linked Tree-sitter grammar.
// NewRegistry validates and copies every mutable field.
type Language struct {
	Name       string
	Aliases    []string
	Filenames  []string
	MIMETypes  []string
	Priority   float32
	Grammar    *sitter.Language
	Highlights string
	Captures   CaptureMap
	Injections []Injection
	Analyse    func(source string) float32
}

// LanguageInfo is the immutable public metadata for a registered language.
type LanguageInfo struct {
	Name      string
	Aliases   []string
	Filenames []string
	MIMETypes []string
	Priority  float32
}

// SourceContractError reports a lexer that changed source text despite being
// called with line-ending and trailing-newline normalization disabled.
type SourceContractError struct {
	Language   string
	Offset     int
	SourceSize int
	TokenSize  int
}

func (e *SourceContractError) Error() string {
	return fmt.Sprintf(
		"%s lexer changed source at byte %d (source %d bytes, tokens %d bytes)",
		e.Language,
		e.Offset,
		e.SourceSize,
		e.TokenSize,
	)
}

// ErrClosed is returned after a Session has released its native resources.
var ErrClosed = errors.New("highlight session is closed")

func captureClass(name string) string {
	if name == "" {
		return ""
	}
	return "tree-sitter." + name
}

// ChromaClass returns a stable dotted lowercase class for a Chroma token type.
func ChromaClass(tokenType chroma.TokenType) string {
	name := tokenType.String()
	if name == "" {
		return "chroma.unknown"
	}

	var class strings.Builder
	class.WriteString("chroma.")
	for i, r := range name {
		if unicode.IsUpper(r) && i > 0 {
			previous, _ := utf8.DecodeLastRuneInString(name[:i])
			next, _ := utf8.DecodeRuneInString(name[i+utf8.RuneLen(r):])
			if unicode.IsLower(previous) || (next != utf8.RuneError && unicode.IsLower(next)) {
				class.WriteByte('.')
			}
		}
		class.WriteRune(unicode.ToLower(r))
	}
	return class.String()
}
