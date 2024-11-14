package theme

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	Black    = lipgloss.Color("#000000")
	Teal     = lipgloss.Color("#2ea043")
	Orange   = lipgloss.Color("#f0883e")
	Blue     = lipgloss.Color("#58a6ff")
	Gray     = lipgloss.Color("#6e7681")
	DarkGray = lipgloss.Color("#21262d")

	// Base styles
	BaseStyle = lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	// Text styles
	Title = BaseStyle.
		Foreground(Blue).
		Bold(true).
		PaddingBottom(1)

	Subtitle = BaseStyle.
		Foreground(Gray).
		PaddingBottom(1)

	// Command styles
	CommandName = BaseStyle.
		Foreground(Orange).
		Bold(true)

	CommandDesc = BaseStyle.
		Foreground(Gray)

	// List styles
	ListItem = BaseStyle.
		PaddingLeft(2)

	ListItemKey = ListItem.
		Foreground(Teal).
		Bold(true)

	ListItemValue = ListItem.
		Foreground(Gray)

	// Status styles
	Success = BaseStyle.
		Foreground(Teal).
		Bold(true)

	Warning = BaseStyle.
		Foreground(Orange).
		Bold(true)

	Error = BaseStyle.
		Foreground(lipgloss.Color("#f85149")).
		Bold(true)

	// URL styles
	URL = BaseStyle.
		Foreground(Blue).
		Underline(true)

	DeleteURL = BaseStyle.
		Foreground(lipgloss.Color("#f85149")).
		Underline(true)

	// Table styles
	TableHeader = BaseStyle.
		Foreground(Blue).
		Bold(true).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(DarkGray)

	TableCell = BaseStyle.
		Foreground(Gray)

	// Help styles
	HelpCommand = BaseStyle.
		Foreground(Orange).
		Bold(true).
		PaddingRight(2)

	HelpDesc = BaseStyle.
		Foreground(Gray)

	HelpFlag = BaseStyle.
		Foreground(Teal).
		Bold(true).
		PaddingRight(2)

	// Box styles
	InfoBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Blue).
		Padding(1).
		MarginTop(1).
		MarginBottom(1)

	WarningBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Orange).
		Padding(1).
		MarginTop(1).
		MarginBottom(1)

	ErrorBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#f85149")).
		Padding(1).
		MarginTop(1).
		MarginBottom(1)
)

// Helper functions for common text formatting
func FormatCommand(name string) string {
	return CommandName.Render(name)
}

func FormatURL(url string) string {
	return URL.Render(url)
}

func FormatDeleteURL(url string) string {
	return DeleteURL.Render(url)
}

func FormatError(msg string) string {
	return Error.Render(msg)
}

func FormatSuccess(msg string) string {
	return Success.Render(msg)
}

func FormatWarning(msg string) string {
	return Warning.Render(msg)
}

func FormatKeyValue(key, value string) string {
	return ListItemKey.Render(key+":") + " " + ListItemValue.Render(value)
}

func RenderInfoBox(msg string) string {
	return InfoBox.Render(msg)
}

func RenderWarningBox(msg string) string {
	return WarningBox.Render(msg)
}

func RenderErrorBox(msg string) string {
	return ErrorBox.Render(msg)
}
