package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss"
)

const space = 18

var highlighter = regexp.MustCompile("{{(.+?)}}")

type usageStyles struct {
	codeBlock lipgloss.Style
	program   lipgloss.Style
	string    lipgloss.Style
	argument  lipgloss.Style
	flag      lipgloss.Style
	title     lipgloss.Style
}

type flagHelpStyles struct {
	dash    lipgloss.Style
	help    lipgloss.Style
	keyword lipgloss.Style
}

func newUsageStyles() usageStyles {
	codeBlock := lipgloss.NewStyle().
		Background(lipgloss.Color(srceryGray1)).
		MarginLeft(2).
		Padding(1, 2)
	background := codeBlock.GetBackground()

	return usageStyles{
		codeBlock: codeBlock,
		program:   lipgloss.NewStyle().Background(background).Foreground(lipgloss.Color(srceryBrightBlue)).PaddingLeft(1),
		string:    lipgloss.NewStyle().Background(background).Foreground(lipgloss.Color(srceryBrightGreen)).PaddingLeft(1),
		argument:  lipgloss.NewStyle().Background(background).Foreground(lipgloss.Color(srceryBrightWhite)).PaddingLeft(1),
		flag:      lipgloss.NewStyle().Background(background).Foreground(lipgloss.Color(srceryBrightBlack)).PaddingLeft(1),
		title: lipgloss.NewStyle().
			Bold(true).
			Transform(strings.ToUpper).
			Margin(1, 0, 0, 2).
			Foreground(lipgloss.Color(srceryBrightBlue)),
	}
}

func newFlagHelpStyles() flagHelpStyles {
	return flagHelpStyles{
		dash:    lipgloss.NewStyle().Foreground(lipgloss.Color(srceryBrightBlack)).MarginLeft(1),
		help:    lipgloss.NewStyle().Foreground(lipgloss.Color(srceryBrightBlack)),
		keyword: lipgloss.NewStyle().Foreground(lipgloss.Color(srceryBrightRed)),
	}
}

func helpPrinter(_ kong.HelpOptions, ctx *kong.Context) error {
	styles := newUsageStyles()

	fmt.Println()
	fmt.Println("  Generate images of code and terminal output. 📸")

	fmt.Println(styles.title.Render(strings.ToUpper("Usage")))
	fmt.Println()
	fmt.Println(
		styles.codeBlock.Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				lipgloss.JoinHorizontal(lipgloss.Left, styles.program.Render("blizzaga"), styles.argument.Render("main.go"), styles.flag.Render("[-o code.svg] [--flags]")),
				lipgloss.JoinHorizontal(lipgloss.Left, styles.program.Render("blizzaga"), styles.argument.Render("--execute"), styles.string.Render("\"ls -la\""), styles.flag.Render("[--flags]   ")),
			),
		),
	)

	flags := ctx.Flags()
	lastGroup := ""

	fmt.Println()
	for _, f := range flags {
		if f.Name == "interactive" {
			printFlag(f)
		}
	}

	fmt.Println(styles.title.Render("Settings"))

	for _, f := range flags {
		if f.Group != nil && f.Group.Title == "Settings" {
			if f.Hidden || f.Name == "help" {
				continue
			}
			printFlag(f)
		}
	}

	fmt.Print(styles.title.Render("Customization"))

	for _, f := range flags {
		if f.Hidden || f.Name == "help" || f.Group.Title == "Settings" {
			continue
		}

		if f.Group != nil && lastGroup != f.Group.Title {
			lastGroup = f.Group.Title
			fmt.Println()
		}

		printFlag(f)
	}
	fmt.Println()
	return nil
}

func printFlag(f *kong.Flag) {
	styles := newFlagHelpStyles()

	if f.Short > 0 {
		fmt.Print("    ", styles.dash.Render("-"), string(f.Short))
		fmt.Print(styles.dash.Render("--"), f.Name)
		fmt.Print(strings.Repeat(" ", space-len(f.Name)))
	} else {
		fmt.Print("    ", styles.dash.Render(" "), " ")
		fmt.Print(styles.dash.Render("--"), f.Name)
		fmt.Print(strings.Repeat(" ", space-len(f.Name)))
	}
	fmt.Println(renderHighlightedHelp(f.Help, styles.help, styles.keyword))
}

func renderHighlightedHelp(help string, helpStyle lipgloss.Style, keywordStyle lipgloss.Style) string {
	matches := highlighter.FindAllStringSubmatchIndex(help, -1)
	if len(matches) == 0 {
		return helpStyle.Render(help)
	}

	var b strings.Builder
	last := 0
	for _, match := range matches {
		b.WriteString(helpStyle.Render(help[last:match[0]]))
		b.WriteString(keywordStyle.Render(help[match[2]:match[3]]))
		last = match[1]
	}
	b.WriteString(helpStyle.Render(help[last:]))
	return b.String()
}
