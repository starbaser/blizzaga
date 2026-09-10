package highlight

import "github.com/alecthomas/chroma/v2"

const (
	srceryBlack         = "#121110"
	srceryRed           = "#EF2F27"
	srceryGreen         = "#519F50"
	srceryYellow        = "#FBB829"
	srceryBlue          = "#2C78BF"
	srceryBrightBlack   = "#917E6B"
	srceryBrightRed     = "#F75341"
	srceryBrightGreen   = "#98BC37"
	srceryBrightYellow  = "#FED06E"
	srceryBrightBlue    = "#68A8E4"
	srceryBrightMagenta = "#FF5C8F"
	srceryBrightCyan    = "#2BE4D0"
	srceryBrightWhite   = "#FCE8C3"
	srceryDarkRed       = "#4F2321"
	srceryDarkGreen     = "#294229"
	srceryOrange        = "#FF5F00"
	srceryBrightOrange  = "#FF8700"
	srceryGray1         = "#1C1B19"
)

var srceryStyle = chroma.MustNewStyle("srcery", chroma.StyleEntries{
	chroma.Text:                   srceryBrightWhite,
	chroma.LineNumbers:            srceryBrightBlack,
	chroma.LineNumbersTable:       srceryBrightBlack,
	chroma.Error:                  "bold " + srceryBrightWhite + " bg:" + srceryRed,
	chroma.Keyword:                srceryRed,
	chroma.KeywordConstant:        srceryBrightMagenta,
	chroma.KeywordDeclaration:     srceryOrange,
	chroma.KeywordNamespace:       srceryBrightRed,
	chroma.KeywordType:            "italic " + srceryBrightBlue,
	chroma.Name:                   srceryBrightWhite,
	chroma.NameAttribute:          srceryYellow,
	chroma.NameBuiltin:            srceryBrightBlue,
	chroma.NameClass:              "italic " + srceryBrightBlue,
	chroma.NameConstant:           srceryBrightMagenta,
	chroma.NameDecorator:          srceryBrightOrange,
	chroma.NameException:          srceryRed,
	chroma.NameFunction:           srceryYellow,
	chroma.NameKeyword:            srceryRed,
	chroma.NameLabel:              "#C5B088",
	chroma.NameOperator:           "#C5B088",
	chroma.NameProperty:           srceryBrightBlue,
	chroma.NameTag:                srceryBlue,
	chroma.Literal:                srceryBrightMagenta,
	chroma.LiteralString:          srceryBrightGreen,
	chroma.LiteralStringBoolean:   srceryBrightMagenta,
	chroma.LiteralStringChar:      srceryGreen,
	chroma.LiteralStringDelimiter: srceryGreen,
	chroma.LiteralStringEscape:    srceryYellow,
	chroma.LiteralStringRegex:     srceryYellow,
	chroma.LiteralNumber:          srceryBrightMagenta,
	chroma.Operator:               "#C5B088",
	chroma.Punctuation:            srceryBrightBlack,
	chroma.TextPunctuation:        srceryBrightBlack,
	chroma.TextSymbol:             srceryYellow,
	chroma.Comment:                "italic " + srceryBrightBlack,
	chroma.CommentSpecial:         "italic " + srceryBrightCyan,
	chroma.CommentPreproc:         "#0AAEB3",
	chroma.GenericDeleted:         "bg:" + srceryDarkRed,
	chroma.GenericEmph:            "italic",
	chroma.GenericError:           srceryBrightRed,
	chroma.GenericHeading:         "bold " + srceryBrightBlue + " bg:" + srceryGray1,
	chroma.GenericInserted:        "bg:" + srceryDarkGreen,
	chroma.GenericOutput:          srceryBrightBlack,
	chroma.GenericPrompt:          "#C5B088",
	chroma.GenericStrong:          "bold",
	chroma.GenericSubheading:      "bold " + srceryYellow + " bg:" + srceryGray1,
	chroma.GenericTraceback:       srceryBrightRed,
	chroma.GenericUnderline:       "underline",
})

// SrceryStyle returns Blizzaga's immutable Srcery Chroma style.
func SrceryStyle() *chroma.Style {
	return srceryStyle
}
