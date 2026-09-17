package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/alecthomas/chroma/v2"
	formatter "github.com/alecthomas/chroma/v2/formatters/svg"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/alecthomas/kong"
	"github.com/beevik/etree"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
	"github.com/mattn/go-isatty"

	"github.com/starbaser/blizzaga/highlight"
	in "github.com/starbaser/blizzaga/input"
	"github.com/starbaser/blizzaga/internal/chromawrap"
	"github.com/starbaser/blizzaga/render"
)

var (
	// Version contains the application version number. It's set via ldflags
	// when building.
	Version = ""

	// CommitSHA contains the SHA of the commit that this application was built
	// against. It's set via ldflags when building.
	CommitSHA = ""
)

func main() {
	const shaLen = 7

	var (
		input  string
		err    error
		config Config
		scale  float64
	)

	k, err := kong.New(&config, kong.Help(helpPrinter))
	if err != nil {
		printErrorFatal("Something went wrong", err)
	}
	ctx, err := k.Parse(os.Args[1:])
	if err != nil || ctx.Error != nil {
		printErrorFatal("Invalid Usage", err)
	}

	if config.Version {
		if Version == "" {
			info, ok := debug.ReadBuildInfo()
			if ok && info.Main.Sum != "" {
				Version = info.Main.Version
			} else {
				Version = "unknown (built from source)"
			}
		}
		version := fmt.Sprintf("blizzaga version %s", Version)
		if len(CommitSHA) >= shaLen {
			version += " (" + CommitSHA[:shaLen] + ")"
		}
		fmt.Println(version)
		os.Exit(0)
	}

	// Copy the pty output to buffer
	if config.Execute != "" {
		input, err = executeCommand(config)
		if err != nil {
			if input != "" {
				err = fmt.Errorf("%w\n%s", err, input)
			}
			printErrorFatal("Something went wrong", err)
		}
		if input == "" {
			printErrorFatal("Something went wrong", errors.New("no command output"))
		}
	}

	isDefaultConfig := config.Config == "default"
	configFile, err := configs.Open("configurations/" + config.Config + ".json")
	if config.Config == "user" {
		configFile, err = loadUserConfig()
	}
	if err != nil {
		configFile, err = os.Open(config.Config)
	}
	if err != nil {
		configFile, _ = configs.Open("configurations/base.json")
	}
	r, err := kong.JSON(configFile)
	if err != nil {
		printErrorFatal("Invalid JSON", err)
	}
	k, err = kong.New(&config, kong.Help(helpPrinter), kong.Resolvers(r))
	if err != nil {
		printErrorFatal("Something went wrong", err)
	}
	ctx, err = k.Parse(os.Args[1:])
	if err != nil {
		printErrorFatal("Invalid Usage", err)
	}

	if config.Interactive {
		cfg, interactiveErr := runForm(&config)
		config = *cfg
		if interactiveErr != nil {
			printErrorFatal("", interactiveErr)
		}
		if isDefaultConfig {
			_ = saveUserConfig(*cfg)
		}
	}

	autoHeight := config.Height == 0
	autoWidth := config.Width == 0

	if config.Output == "" {
		config.Output = defaultOutputFilename
	}

	scale = 1
	if autoHeight && autoWidth && strings.HasSuffix(config.Output, ".png") {
		scale = 4
	}

	if config.Input == "" && !in.IsPipe(os.Stdin) && len(ctx.Args) <= 0 {
		_ = helpPrinter(kong.HelpOptions{}, ctx)
		os.Exit(0)
	}

	highlighter, err := highlight.DefaultRegistry()
	if err != nil {
		printErrorFatal("Could not initialize syntax highlighting", err)
	}

	// An explicit file argument wins over a piped stdin: scripts and agents
	// often run with stdin attached to a pipe that never closes.
	if config.Input == "-" || (config.Input == "" && in.IsPipe(os.Stdin)) {
		input, err = in.ReadInput(os.Stdin)
	} else if config.Execute != "" {
		config.Language = "ansi"
	} else {
		input, err = in.ReadFile(config.Input)
		if err != nil {
			printErrorFatal("File not found", err)
		}
	}

	// adjust for 1-indexing
	for i := range config.Lines {
		config.Lines[i]--
	}

	if isMarkdownLanguage(config.Language) {
		input, err = renderMarkdown(input, config.Theme, config.Wrap)
		if err != nil {
			printErrorFatal("Markdown render", err)
		}
		config.Language = "ansi"
		config.Wrap = 0
	}

	strippedInput := ansi.Strip(input)
	isAnsi := strings.ToLower(config.Language) == "ansi" || strippedInput != input
	strippedInput = cut(strippedInput, config.Lines)
	input = cut(input, config.Lines)

	// lineSources maps rendered rows back to source lines once wrapping has
	// split some of them; nil means every row is its own source line.
	var lineSources []int

	// wrap to character limit.
	if config.Wrap > 0 && isAnsi {
		// The stripped text only sizes the SVG, but wrapping it through the
		// projection also yields the row map the line-number gutter needs.
		wrapped, wrapErr := chromawrap.WrapTokens(
			chroma.Literator(chroma.Token{Type: chroma.Text, Value: strippedInput}),
			config.Wrap,
		)
		if wrapErr != nil {
			printErrorFatal("Could not wrap input", wrapErr)
		}
		strippedInput = wrapped.Text
		lineSources = wrapped.LineSources
		input = cellbuf.Wrap(input, config.Wrap, "")
	}

	if input == "" {
		if err != nil {
			printErrorFatal("No input", err)
		} else {
			printErrorFatal("No input", errors.New("check --lines is within bounds"))
		}
	}

	s, ok := styles.Registry[strings.ToLower(config.Theme)]
	if s == nil || !ok {
		s = defaultStyle
	}
	if !s.Has(chroma.Background) {
		s, err = s.Builder().Add(chroma.Background, "bg:"+config.Background).Build()
		if err != nil {
			printErrorFatal("Could not add background", err)
		}
	}

	// Create a token iterator.
	var it chroma.Iterator
	if isAnsi {
		// For ANSI output, we'll inject our own SVG. For now, let's just strip the ANSI
		// codes and print the text to properly size the input.
		it = chroma.Literator(chroma.Token{Type: chroma.Text, Value: strippedInput})
	} else {
		result, highlightErr := highlighter.Highlight(input, highlight.Query{
			Language: config.Language,
			Filename: config.Input,
		})
		err = highlightErr
		if err != nil {
			printErrorFatal("Could not lex file", err)
		}
		it = chroma.Literator(result.Tokens()...)
		if config.Wrap > 0 {
			wrapped, wrapErr := chromawrap.WrapTokens(it, config.Wrap)
			if wrapErr != nil {
				printErrorFatal("Could not wrap highlighted file", wrapErr)
			}
			it = wrapped.Iterator
			strippedInput = wrapped.Text
			lineSources = wrapped.LineSources
		}
	}

	// Format the code to an SVG.
	options, err := fontOptions(&config)
	if err != nil {
		printErrorFatal("Invalid font options", err)
	}

	f := formatter.New(options...)
	if err != nil {
		printErrorFatal("Malformed text", err)
	}

	buf := &bytes.Buffer{}
	err = f.Format(buf, s, it)
	if err != nil {
		log.Fatal(err)
	}

	// Parse SVG (XML document)
	doc := etree.NewDocument()
	_, err = doc.ReadFrom(buf)
	if err != nil {
		printErrorFatal("Bad SVG", err)
	}
	embedDefaultFontFaces(doc.Root(), &config)

	rcfg := config.renderConfig()
	rcfg.Margin = render.ExpandMargin(rcfg.Margin, scale)
	rcfg.Padding = render.ExpandPadding(rcfg.Padding, scale)

	tabWidth := 4
	if isAnsi {
		tabWidth = 6
	}
	longestLineCols := lipgloss.Width(strings.ReplaceAll(strippedInput, "\t", strings.Repeat(" ", tabWidth)))

	offsetLine := 0
	if len(config.Lines) > 0 {
		offsetLine = config.Lines[0]
	}

	decorated, err := render.Decorate(&rcfg, render.DecorateParams{
		Doc:             doc,
		Scale:           scale,
		AutoWidth:       autoWidth,
		AutoHeight:      autoHeight,
		IsAnsi:          isAnsi,
		LongestLineCols: longestLineCols,
		OffsetLine:      offsetLine,
		LineSources:     lineSources,
		LineNumberColor: s.Get(chroma.LineNumbers).Colour.String(),
	})
	if err != nil {
		printErrorFatal("Bad Output", err)
	}
	imageWidth := decorated.Width
	imageHeight := decorated.Height

	if isAnsi {
		colAdvance, err := render.ColAdvance(rcfg)
		if err != nil {
			printErrorFatal("Font metrics", err)
		}
		d := dispatcher{lines: decorated.TextLines, svg: decorated.TextGroup, config: &rcfg, scale: decorated.Scale, colAdvance: colAdvance}
		parser := ansi.NewParser()
		parser.SetHandler(ansi.Handler{
			Print:     d.Print,
			HandleCsi: d.CsiDispatch,
			Execute:   d.Execute,
		})
		for _, line := range strings.Split(input, "\n") {
			parser.Parse([]byte(line))
			d.Execute(ansi.LF) // simulate a newline
		}
	}

	istty := isatty.IsTerminal(os.Stdout.Fd())

	switch {
	case strings.HasSuffix(config.Output, ".png"):
		pngBytes, rerr := render.Rasterize(doc, imageWidth, imageHeight)
		if rerr != nil {
			printErrorFatal("Unable to convert SVG to PNG", rerr)
		}
		if werr := os.WriteFile(config.Output, pngBytes, 0o600); werr != nil {
			printErrorFatal("Unable to write output", werr)
		}
		printFilenameOutput(config.Output)

	default:
		// output file specified.
		if config.Output != "" {
			err = doc.WriteToFile(config.Output)
			if err != nil {
				printErrorFatal("Unable to write output", err)
			}
			printFilenameOutput(config.Output)
			return
		}

		// reading from stdin.
		if config.Input == "" || config.Input == "-" {
			if istty {
				err = doc.WriteToFile(defaultOutputFilename)
				printFilenameOutput(defaultOutputFilename)
			} else {
				_, err = doc.WriteTo(os.Stdout)
			}
			if err != nil {
				printErrorFatal("Unable to write output", err)
			}
			return
		}

		// reading from file.
		if istty {
			config.Output = strings.TrimSuffix(filepath.Base(config.Input), filepath.Ext(config.Input)) + ".svg"
			err = doc.WriteToFile(config.Output)
			printFilenameOutput(config.Output)
		} else {
			_, err = doc.WriteTo(os.Stdout)
		}
		if err != nil {
			printErrorFatal("Unable to write output", err)
		}
	}
}

var outputHeader = lipgloss.NewStyle().
	Foreground(lipgloss.Color(srceryBrightWhite)).
	Background(lipgloss.Color(srceryBlue)).
	Bold(true).
	Padding(0, 1).
	MarginRight(1).
	SetString("WROTE")

func printFilenameOutput(filename string) {
	fmt.Println(lipgloss.JoinHorizontal(lipgloss.Center, outputHeader.String(), filename))
}
