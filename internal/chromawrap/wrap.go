// Package chromawrap projects Chroma classifications across display wrapping.
package chromawrap

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2"
	"github.com/charmbracelet/x/cellbuf"
)

// Continuation marks a display row that continues the previous row's source
// line rather than starting a new one.
const Continuation = -1

// Wrapped is the display projection of a token stream after wrapping.
type Wrapped struct {
	// Iterator yields the wrapped text with every original Chroma style
	// projected onto it.
	Iterator chroma.Iterator

	// Text is the wrapped source, byte-identical to cellbuf.Wrap of the input.
	Text string

	// LineSources maps each display row of Text to the 0-based index of the
	// source line it begins, or Continuation when the row is the tail of a
	// source line that wrapped. A trailing newline in Text yields one extra
	// entry for the empty row after it.
	LineSources []int
}

// WrapTokens applies Blizzaga's display wrapping after syntax analysis and
// projects each original Chroma style onto the wrapped text.
func WrapTokens(iterator chroma.Iterator, width int) (Wrapped, error) {
	tokens := iterator.Tokens()
	source, tokenTypes := flattenTokens(tokens)
	wrapped := cellbuf.Wrap(source, width, "")
	projected, lineSources, err := projectTokenTypes(source, wrapped, tokenTypes)
	if err != nil {
		return Wrapped{}, err
	}
	return Wrapped{
		Iterator:    chroma.Literator(projected...),
		Text:        wrapped,
		LineSources: lineSources,
	}, nil
}

func flattenTokens(tokens []chroma.Token) (string, []chroma.TokenType) {
	length := 0
	for _, token := range tokens {
		length += len(token.Value)
	}

	source := make([]byte, 0, length)
	tokenTypes := make([]chroma.TokenType, 0, length)
	for _, token := range tokens {
		source = append(source, token.Value...)
		for range len(token.Value) {
			tokenTypes = append(tokenTypes, token.Type)
		}
	}
	return string(source), tokenTypes
}

// projectTokenTypes walks the wrapped text alongside the source it came from.
// Every byte either matches the source byte at the cursor or is a newline the
// wrapper inserted; a source newline that the wrapper reached only after
// trimming trailing whitespace is still a source newline. The second result
// records, per display row, which source line the row starts.
func projectTokenTypes(source, wrapped string, tokenTypes []chroma.TokenType) ([]chroma.Token, []int, error) {
	if len(source) != len(tokenTypes) {
		return nil, nil, fmt.Errorf("highlight token byte count %d does not match source byte count %d", len(tokenTypes), len(source))
	}

	output := make([]chroma.Token, 0)
	sourceLine := 0
	lineSources := []int{sourceLine}
	sourceOffset := 0
	wrappedOffset := 0
	for wrappedOffset < len(wrapped) {
		if sourceOffset < len(source) && wrapped[wrappedOffset] == source[sourceOffset] {
			if wrapped[wrappedOffset] == '\n' {
				sourceLine++
				lineSources = append(lineSources, sourceLine)
			}
			appendProjectedToken(&output, tokenTypes[sourceOffset], wrapped[wrappedOffset:wrappedOffset+1])
			wrappedOffset++
			sourceOffset++
			continue
		}

		if wrapped[wrappedOffset] != '\n' {
			return nil, nil, fmt.Errorf(
				"wrapped source diverged at output byte %d and source byte %d",
				wrappedOffset,
				sourceOffset,
			)
		}

		spaceEnd := skipHorizontalWhitespace(source, sourceOffset)
		newlineType := chroma.TextWhitespace
		if sourceOffset < len(tokenTypes) {
			newlineType = tokenTypes[sourceOffset]
		} else if len(output) > 0 {
			newlineType = output[len(output)-1].Type
		}

		if spaceEnd < len(source) && source[spaceEnd] == '\n' {
			newlineType = tokenTypes[spaceEnd]
			sourceOffset = spaceEnd + 1
			sourceLine++
			lineSources = append(lineSources, sourceLine)
		} else {
			sourceOffset = spaceEnd
			lineSources = append(lineSources, Continuation)
		}
		appendProjectedToken(&output, newlineType, "\n")
		wrappedOffset++
	}

	if skipHorizontalWhitespace(source, sourceOffset) != len(source) {
		return nil, nil, fmt.Errorf("wrapped source ended before source byte %d", sourceOffset)
	}
	return output, lineSources, nil
}

func skipHorizontalWhitespace(source string, offset int) int {
	for offset < len(source) {
		r, size := utf8.DecodeRuneInString(source[offset:])
		if r == '\n' || !unicode.IsSpace(r) {
			break
		}
		offset += size
	}
	return offset
}

func appendProjectedToken(tokens *[]chroma.Token, tokenType chroma.TokenType, value string) {
	if value == "" {
		return
	}
	if len(*tokens) > 0 && (*tokens)[len(*tokens)-1].Type == tokenType {
		(*tokens)[len(*tokens)-1].Value += value
		return
	}
	*tokens = append(*tokens, chroma.Token{Type: tokenType, Value: value})
}
