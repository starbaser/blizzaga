package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/starbaser/blizzaga/render"
)

func newFormTheme() *huh.Theme {
	accent := lipgloss.Color(srceryBrightGreen)
	highlight := lipgloss.Color(srceryBrightOrange)
	text := lipgloss.Color(srceryBrightWhite)
	soft := lipgloss.Color(srceryWhite)
	muted := lipgloss.Color(srceryBrightBlack)
	errorColor := lipgloss.Color(srceryRed)

	theme := huh.ThemeBase()
	theme.FieldSeparator = lipgloss.NewStyle()

	theme.Focused.Base = theme.Focused.Base.BorderForeground(accent)
	theme.Focused.Card = theme.Focused.Base
	theme.Focused.Title = theme.Focused.Title.Width(18).Foreground(accent).Bold(true)
	theme.Focused.NoteTitle = theme.Focused.NoteTitle.Foreground(accent).Bold(true).Margin(1, 0)
	theme.Focused.Description = theme.Focused.Description.Foreground(soft)
	theme.Focused.ErrorIndicator = theme.Focused.ErrorIndicator.Foreground(errorColor)
	theme.Focused.ErrorMessage = theme.Focused.ErrorMessage.Foreground(errorColor)
	theme.Focused.SelectSelector = theme.Focused.SelectSelector.Foreground(highlight)
	theme.Focused.NextIndicator = theme.Focused.NextIndicator.Foreground(highlight)
	theme.Focused.PrevIndicator = theme.Focused.PrevIndicator.Foreground(highlight)
	theme.Focused.Option = theme.Focused.Option.Foreground(text)
	theme.Focused.Directory = theme.Focused.Directory.Foreground(lipgloss.Color(srceryBrightBlue))
	theme.Focused.File = theme.Focused.File.Foreground(text)
	theme.Focused.MultiSelectSelector = theme.Focused.MultiSelectSelector.Foreground(highlight)
	theme.Focused.SelectedOption = lipgloss.NewStyle().Foreground(accent)
	theme.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(accent).SetString("✓ ")
	theme.Focused.UnselectedOption = theme.Focused.UnselectedOption.Foreground(text)
	theme.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(muted).SetString("• ")
	theme.Focused.FocusedButton = lipgloss.NewStyle().Foreground(text).PaddingRight(1)
	theme.Focused.BlurredButton = lipgloss.NewStyle().Foreground(muted).PaddingRight(1)
	theme.Focused.Next = theme.Focused.FocusedButton
	theme.Focused.TextInput.Cursor = theme.Focused.TextInput.Cursor.Foreground(accent)
	theme.Focused.TextInput.Placeholder = theme.Focused.TextInput.Placeholder.Foreground(muted)
	theme.Focused.TextInput.Prompt = theme.Focused.TextInput.Prompt.Foreground(highlight)
	theme.Focused.TextInput.Text = theme.Focused.TextInput.Text.Foreground(text)

	theme.Blurred = theme.Focused
	theme.Blurred.Base = theme.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	theme.Blurred.Card = theme.Blurred.Base
	theme.Blurred.Title = theme.Blurred.Title.Width(18).Foreground(soft).Bold(false)
	theme.Blurred.NoteTitle = theme.Blurred.NoteTitle.Foreground(soft).Bold(false).Margin(1, 0)
	theme.Blurred.Description = theme.Blurred.Description.Foreground(muted)
	theme.Blurred.SelectSelector = theme.Blurred.SelectSelector.Foreground(muted)
	theme.Blurred.MultiSelectSelector = theme.Blurred.MultiSelectSelector.Foreground(muted)
	theme.Blurred.SelectedOption = theme.Blurred.SelectedOption.Foreground(muted)
	theme.Blurred.TextInput.Text = theme.Blurred.TextInput.Text.Foreground(muted)
	theme.Blurred.NextIndicator = lipgloss.NewStyle()
	theme.Blurred.PrevIndicator = lipgloss.NewStyle()

	theme.Group.Title = theme.Focused.Title
	theme.Group.Description = theme.Focused.Description
	return theme
}

func runForm(config *Config) (*Config, error) {
	var (
		padding      = strings.Trim(fmt.Sprintf("%v", config.Padding), "[]")
		margin       = strings.Trim(fmt.Sprintf("%v", config.Margin), "[]")
		fontSize     = fmt.Sprintf("%d", int(config.Font.Size))
		lineHeight   = fmt.Sprintf("%.1f", config.LineHeight)
		borderRadius = fmt.Sprintf("%.0f", config.Border.Radius)
		borderWidth  = fmt.Sprintf("%.0f", config.Border.Width)
		shadowBlur   = fmt.Sprintf("%.0f", config.Shadow.Blur)
		shadowX      = fmt.Sprintf("%.0f", config.Shadow.X)
		shadowY      = fmt.Sprintf("%.0f", config.Shadow.Y)
	)

	theme := newFormTheme()

	f := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title("\nCapture file"),

			huh.NewFilePicker().
				Title("").
				Picking(true).
				Height(10).
				Value(&config.Input),

			huh.NewNote().Description("Choose a code file to screenshot."),
		).WithHide(config.Input != "" || config.Execute != ""),
		huh.NewGroup(
			huh.NewNote().Title("Settings"),

			huh.NewInput().
				Title("Output").
				Placeholder(defaultOutputFilename).
				// Description("Output location for image.").
				Inline(true).
				Prompt("").
				Value(&config.Output),

			huh.NewSelect[string]().Title("Theme ").
				// Description("Theme for syntax highlighting.").
				Inline(true).
				Options(huh.NewOptions(styles.Names()...)...).
				Value(&config.Theme),

			// huh.NewInput().Title("Background ").
			// 	// Description("Apply a background fill.").
			// 	Placeholder("#FFF").
			// 	Value(&config.Background).
			// 	Inline(true).
			// 	Prompt("").
			// 	Validate(validateColor),

			huh.NewNote().Title("Window"),

			huh.NewInput().Title("Padding ").
				// Description("Apply padding to the code.").
				Placeholder("20 40").
				Inline(true).
				Value(&padding).
				Prompt("").
				Validate(validatePadding),

			huh.NewInput().Title("Margin ").
				// Description("Apply margin to the window.").
				Placeholder("20").
				Inline(true).
				Value(&margin).
				Prompt("").
				Validate(validatePadding),

			huh.NewConfirm().Title("Controls").
				Inline(true).
				Value(&config.Window),

			huh.NewNote().Title("Font"),

			huh.NewInput().Title("Font Family ").
				// Description("Font family to use for code").
				Placeholder("Iosevka Custom").
				Inline(true).
				Prompt("").
				Value(&config.Font.Family),

			huh.NewInput().Title("Font Size ").
				// Description("Font size to use for code.").
				Placeholder("14").
				Inline(true).
				Prompt("").
				Value(&fontSize).
				Validate(validateInteger),

			huh.NewInput().Title("Line Height ").
				// Description("Line height relative to size.").
				Placeholder("1.2").
				Inline(true).
				Prompt("").
				Value(&lineHeight).
				Validate(validateFloat),

			huh.NewNote().Title("Border"),

			huh.NewInput().Title("Border Radius ").
				// Description("Corner radius of the window.").
				Placeholder("0").
				Inline(true).
				Prompt("").
				Value(&borderRadius).
				Validate(validateInteger),

			huh.NewInput().Title("Border Width ").
				// Description("Border width thickness.").
				Placeholder("1").
				Inline(true).
				Prompt("").
				Value(&borderWidth).
				Validate(validateInteger),

			huh.NewInput().Title("Border Color ").
				// Description("Color of outline stroke.").
				Validate(validateColor).
				Inline(true).
				Prompt("").
				Value(&config.Border.Color).
				Placeholder("#515151"),

			huh.NewNote().Title("Shadow"),

			huh.NewInput().Title("Blur ").
				// Description("Shadow Gaussian Blur.").
				Placeholder("0").
				Inline(true).
				Prompt("").
				Value(&shadowBlur).
				Validate(validateInteger),

			huh.NewInput().Title("X Offset ").
				// Description("Shadow offset x coordinate").
				Placeholder("0").
				Inline(true).
				Prompt("").
				Value(&shadowX).
				Validate(validateInteger),

			huh.NewInput().Title("Y Offset ").
				// Description("Shadow offset y coordinate").
				Placeholder("0").
				Inline(true).
				Prompt("").
				Value(&shadowY).
				Validate(validateInteger),
		).WithHeight(33),
	).WithTheme(theme).WithWidth(40)

	err := f.Run()

	if config.Output == "" {
		config.Output = defaultOutputFilename
	}

	config.Padding = parsePadding(padding)
	config.Margin = parseMargin(margin)
	config.Font.Size, _ = strconv.ParseFloat(fontSize, 64)
	config.LineHeight, _ = strconv.ParseFloat(lineHeight, 64)
	config.Border.Radius, _ = strconv.ParseFloat(borderRadius, 64)
	config.Border.Width, _ = strconv.ParseFloat(borderWidth, 64)
	config.Shadow.Blur, _ = strconv.ParseFloat(shadowBlur, 64)
	config.Shadow.X, _ = strconv.ParseFloat(shadowX, 64)
	config.Shadow.Y, _ = strconv.ParseFloat(shadowY, 64)
	return config, err //nolint: wrapcheck
}

func validateMargin(s string) error {
	tokens := strings.Fields(s)
	if len(tokens) > 4 {
		return errors.New("maximum four values")
	}
	for _, t := range tokens {
		_, err := strconv.Atoi(t)
		if err != nil {
			return errors.New("must be valid space-separated integers")
		}
	}
	return nil
}

func validatePadding(s string) error {
	return validateMargin(s)
}

func validateInteger(s string) error {
	if len(s) <= 0 {
		return nil
	}

	_, err := strconv.Atoi(s)
	if err != nil {
		return errors.New("must be valid integer")
	}
	return nil
}

func validateFloat(s string) error {
	if len(s) <= 0 {
		return nil
	}

	_, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return errors.New("must be valid float")
	}
	return nil
}

var colorRegex = regexp.MustCompile("^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$")

func validateColor(s string) error {
	if len(s) <= 0 {
		return nil
	}

	if !colorRegex.MatchString(s) {
		return errors.New("must be valid color")
	}
	return nil
}

func parsePadding(v string) []float64 {
	//nolint: prealloc
	var values []float64
	for _, p := range strings.Fields(v) {
		pi, _ := strconv.ParseFloat(p, 64) // already validated
		values = append(values, pi)
	}
	return render.ExpandPadding(values, 1)
}

var parseMargin = parsePadding
