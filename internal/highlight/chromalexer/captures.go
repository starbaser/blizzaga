package chromalexer

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
)

// CaptureStyle maps a Tree-sitter capture to a Chroma token type. Priority
// breaks ties when several captures cover the exact same byte range.
type CaptureStyle struct {
	TokenType chroma.TokenType
	Priority  int
}

// CaptureMap maps conventional Tree-sitter highlight capture names to Chroma.
// Dotted captures fall back through their parents when no exact entry exists.
type CaptureMap map[string]CaptureStyle

// StandardCaptures returns a fresh mapping for the conventional highlight
// capture vocabulary used by Tree-sitter grammar repositories.
func StandardCaptures() CaptureMap {
	return CaptureMap{
		"attribute":             {TokenType: chroma.NameAttribute, Priority: 140},
		"boolean":               {TokenType: chroma.KeywordConstant, Priority: 150},
		"comment":               {TokenType: chroma.Comment, Priority: 100},
		"comment.directive":     {TokenType: chroma.CommentPreproc, Priority: 180},
		"comment.documentation": {TokenType: chroma.CommentSpecial, Priority: 160},
		"constant":              {TokenType: chroma.NameConstant, Priority: 120},
		"constant.builtin":      {TokenType: chroma.KeywordConstant, Priority: 180},
		"constant.numeric":      {TokenType: chroma.LiteralNumber, Priority: 180},
		"constructor":           {TokenType: chroma.NameClass, Priority: 180},
		"escape":                {TokenType: chroma.LiteralStringEscape, Priority: 250},
		"exception":             {TokenType: chroma.NameException, Priority: 170},
		"field":                 {TokenType: chroma.NameProperty, Priority: 160},
		"function":              {TokenType: chroma.NameFunction, Priority: 180},
		"function.builtin":      {TokenType: chroma.NameBuiltin, Priority: 220},
		"function.call":         {TokenType: chroma.NameFunction, Priority: 190},
		"function.macro":        {TokenType: chroma.NameFunctionMagic, Priority: 230},
		"function.method":       {TokenType: chroma.NameFunction, Priority: 210},
		"keyword":               {TokenType: chroma.Keyword, Priority: 150},
		"keyword.coroutine":     {TokenType: chroma.KeywordReserved, Priority: 160},
		"keyword.declaration":   {TokenType: chroma.KeywordDeclaration, Priority: 170},
		"keyword.directive":     {TokenType: chroma.KeywordNamespace, Priority: 170},
		"keyword.function":      {TokenType: chroma.KeywordDeclaration, Priority: 170},
		"keyword.import":        {TokenType: chroma.KeywordNamespace, Priority: 170},
		"keyword.operator":      {TokenType: chroma.Operator, Priority: 170},
		"keyword.type":          {TokenType: chroma.KeywordType, Priority: 170},
		"label":                 {TokenType: chroma.NameLabel, Priority: 150},
		"module":                {TokenType: chroma.NameNamespace, Priority: 150},
		"namespace":             {TokenType: chroma.NameNamespace, Priority: 150},
		"number":                {TokenType: chroma.LiteralNumber, Priority: 120},
		"operator":              {TokenType: chroma.Operator, Priority: 130},
		"parameter":             {TokenType: chroma.NameVariable, Priority: 140},
		"property":              {TokenType: chroma.NameProperty, Priority: 160},
		"punctuation":           {TokenType: chroma.Punctuation, Priority: 100},
		"string":                {TokenType: chroma.LiteralString, Priority: 100},
		"string.escape":         {TokenType: chroma.LiteralStringEscape, Priority: 250},
		"string.regex":          {TokenType: chroma.LiteralStringRegex, Priority: 130},
		"tag":                   {TokenType: chroma.NameTag, Priority: 150},
		"type":                  {TokenType: chroma.NameClass, Priority: 170},
		"type.builtin":          {TokenType: chroma.KeywordType, Priority: 190},
		"variable":              {TokenType: chroma.NameVariable, Priority: 100},
		"variable.builtin":      {TokenType: chroma.NameBuiltin, Priority: 180},
		"variable.member":       {TokenType: chroma.NameVariableInstance, Priority: 160},
		"variable.other.member": {TokenType: chroma.NameProperty, Priority: 170},
		"variable.parameter":    {TokenType: chroma.NameVariable, Priority: 150},
		"punctuation.bracket":   {TokenType: chroma.Punctuation, Priority: 110},
		"punctuation.delimiter": {TokenType: chroma.Punctuation, Priority: 110},
	}
}

func (m CaptureMap) lookup(name string) (CaptureStyle, bool) {
	for {
		if style, ok := m[name]; ok {
			return style, true
		}
		separator := strings.LastIndexByte(name, '.')
		if separator < 0 {
			return CaptureStyle{}, false
		}
		name = name[:separator]
	}
}
