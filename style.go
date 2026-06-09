package main

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
)

const (
	srceryBlack         = "#121110"
	srceryRed           = "#EF2F27"
	srceryGreen         = "#519F50"
	srceryYellow        = "#FBB829"
	srceryBlue          = "#2C78BF"
	srceryMagenta       = "#E02C6D"
	srceryCyan          = "#0AAEB3"
	srceryWhite         = "#C5B088"
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

var charmStyle = styles.Register(chroma.MustNewStyle("charm", chroma.StyleEntries{
	chroma.Text:                "#C4C4C4",
	chroma.Error:               "#F1F1F1 bg:#F05B5B",
	chroma.Comment:             "#676767",
	chroma.CommentPreproc:      "#FF875F",
	chroma.Keyword:             "#00AAFF",
	chroma.KeywordReserved:     "#FF48DD",
	chroma.KeywordNamespace:    "#FF5F87",
	chroma.KeywordType:         "#635ADF",
	chroma.Operator:            "#FF7F83",
	chroma.Punctuation:         "#E8E8A8",
	chroma.Name:                "#C4C4C4",
	chroma.NameBuiltin:         "#FF7CDB",
	chroma.NameTag:             "#B083EA",
	chroma.NameAttribute:       "#7A7AE6",
	chroma.NameClass:           "#F1F1F1 underline bold",
	chroma.NameDecorator:       "#FFFF87",
	chroma.NameFunction:        "#00DC7F",
	chroma.LiteralNumber:       "#6EEFC0",
	chroma.LiteralString:       "#E38356",
	chroma.LiteralStringEscape: "#AFFFD7",
	chroma.GenericDeleted:      "#FD5B5B",
	chroma.GenericEmph:         "italic",
	chroma.GenericInserted:     "#00D787",
	chroma.GenericStrong:       "bold",
	chroma.GenericSubheading:   "#777777",
}))

var srceryStyle = styles.Register(chroma.MustNewStyle("srcery", chroma.StyleEntries{
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
	chroma.NameLabel:              srceryWhite,
	chroma.NameOperator:           srceryWhite,
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
	chroma.Operator:               srceryWhite,
	chroma.Punctuation:            srceryBrightBlack,
	chroma.TextPunctuation:        srceryBrightBlack,
	chroma.TextSymbol:             srceryYellow,
	chroma.Comment:                "italic " + srceryBrightBlack,
	chroma.CommentSpecial:         "italic " + srceryBrightCyan,
	chroma.CommentPreproc:         srceryCyan,
	chroma.GenericDeleted:         "bg:" + srceryDarkRed,
	chroma.GenericEmph:            "italic",
	chroma.GenericError:           srceryBrightRed,
	chroma.GenericHeading:         "bold " + srceryBrightBlue + " bg:" + srceryGray1,
	chroma.GenericInserted:        "bg:" + srceryDarkGreen,
	chroma.GenericOutput:          srceryBrightBlack,
	chroma.GenericPrompt:          srceryWhite,
	chroma.GenericStrong:          "bold",
	chroma.GenericSubheading:      "bold " + srceryYellow + " bg:" + srceryGray1,
	chroma.GenericTraceback:       srceryBrightRed,
	chroma.GenericUnderline:       "underline",
}))

var defaultStyle = srceryStyle
