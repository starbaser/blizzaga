// Package highlight classifies source text with statically linked Tree-sitter
// grammars and falls back to Chroma's lexer catalog.
//
// Results contain exact source text and byte-addressed classifications. Styling
// is deliberately separate: consumers may use SrceryStyle or any other Chroma
// style without reparsing source.
package highlight
