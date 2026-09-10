// Package chromawrap projects Chroma classifications across display wrapping.
package chromawrap

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2"
	"github.com/charmbracelet/x/cellbuf"
)

// WrapTokens applies Blizzaga's display wrapping after syntax analysis and
// projects each original Chroma style onto the wrapped text.
func WrapTokens(iterator chroma.Iterator, width int) (chroma.Iterator, string, error) {
	tokens := iterator.Tokens()
	source, tokenTypes := flattenTokens(tokens)
	wrapped := cellbuf.Wrap(source, width, "")
	projected, err := projectTokenTypes(source, wrapped, tokenTypes)
	if err != nil {
		return nil, "", err
	}
	return chroma.Literator(projected...), wrapped, nil
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

func projectTokenTypes(source, wrapped string, tokenTypes []chroma.TokenType) ([]chroma.Token, error) {
	if len(source) != len(tokenTypes) {
		return nil, fmt.Errorf("highlight token byte count %d does not match source byte count %d", len(tokenTypes), len(source))
	}

	output := make([]chroma.Token, 0)
	sourceOffset := 0
	wrappedOffset := 0
	for wrappedOffset < len(wrapped) {
		if sourceOffset < len(source) && wrapped[wrappedOffset] == source[sourceOffset] {
			appendProjectedToken(&output, tokenTypes[sourceOffset], wrapped[wrappedOffset:wrappedOffset+1])
			wrappedOffset++
			sourceOffset++
			continue
		}

		if wrapped[wrappedOffset] != '\n' {
			return nil, fmt.Errorf(
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
		} else {
			sourceOffset = spaceEnd
		}
		appendProjectedToken(&output, newlineType, "\n")
		wrappedOffset++
	}

	if skipHorizontalWhitespace(source, sourceOffset) != len(source) {
		return nil, fmt.Errorf("wrapped source ended before source byte %d", sourceOffset)
	}
	return output, nil
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
