package highlight

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
)

// StandardCaptures returns a fresh map for Tree-sitter's conventional
// highlight-capture vocabulary.
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
		"keyword.control":       {TokenType: chroma.Keyword, Priority: 170},
		"keyword.coroutine":     {TokenType: chroma.KeywordReserved, Priority: 160},
		"keyword.declaration":   {TokenType: chroma.KeywordDeclaration, Priority: 170},
		"keyword.directive":     {TokenType: chroma.KeywordNamespace, Priority: 170},
		"keyword.function":      {TokenType: chroma.KeywordDeclaration, Priority: 170},
		"keyword.import":        {TokenType: chroma.KeywordNamespace, Priority: 170},
		"keyword.operator":      {TokenType: chroma.Operator, Priority: 170},
		"keyword.type":          {TokenType: chroma.KeywordType, Priority: 170},
		"label":                 {TokenType: chroma.NameLabel, Priority: 150},
		"markup":                {TokenType: chroma.Text, Priority: 90},
		"module":                {TokenType: chroma.NameNamespace, Priority: 150},
		"namespace":             {TokenType: chroma.NameNamespace, Priority: 150},
		"number":                {TokenType: chroma.LiteralNumber, Priority: 120},
		"operator":              {TokenType: chroma.Operator, Priority: 130},
		"parameter":             {TokenType: chroma.NameVariable, Priority: 140},
		"property":              {TokenType: chroma.NameProperty, Priority: 160},
		"punctuation":           {TokenType: chroma.Punctuation, Priority: 100},
		"punctuation.bracket":   {TokenType: chroma.Punctuation, Priority: 110},
		"punctuation.delimiter": {TokenType: chroma.Punctuation, Priority: 110},
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
	}
}

func (m CaptureMap) lookup(name string) (CaptureMapping, bool) {
	for {
		if mapping, ok := m[name]; ok {
			return mapping, true
		}
		separator := strings.LastIndexByte(name, '.')
		if separator < 0 {
			return CaptureMapping{}, false
		}
		name = name[:separator]
	}
}
