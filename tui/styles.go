package tui

import "github.com/charmbracelet/lipgloss"

// Palette — adaptive so it reads well on both light and dark terminals.
var (
	colorAccent  = lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#B69EFF"}
	colorFg      = lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#EAEAEA"}
	colorSubtle  = lipgloss.AdaptiveColor{Light: "#605F6B", Dark: "#9B9AA6"}
	colorFaint   = lipgloss.AdaptiveColor{Light: "#A2A1AD", Dark: "#5C5B66"}
	colorError   = lipgloss.AdaptiveColor{Light: "#C81E4A", Dark: "#FF6188"}
	colorBadgeBg = lipgloss.AdaptiveColor{Light: "#EDE9FE", Dark: "#2A2340"}
	colorOnAcc   = lipgloss.Color("#FFFFFF")
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorOnAcc).
			Background(colorAccent).
			Padding(0, 1)

	// List item rendering.
	selBarStyle      = lipgloss.NewStyle().Foreground(colorAccent)
	itemNameStyle    = lipgloss.NewStyle().Foreground(colorFg)
	itemNameSelStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	itemDetailStyle  = lipgloss.NewStyle().Foreground(colorSubtle)
	idStyle          = lipgloss.NewStyle().Foreground(colorFaint)
	badgeStyle       = lipgloss.NewStyle().Foreground(colorAccent).Background(colorBadgeBg).Padding(0, 1)

	// Form.
	formLabelStyle      = lipgloss.NewStyle().Foreground(colorSubtle)
	formLabelFocusStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	// Shared chrome.
	helpStyle   = lipgloss.NewStyle().Foreground(colorFaint)
	statusStyle = lipgloss.NewStyle().Foreground(colorAccent).Italic(true)
	errorStyle  = lipgloss.NewStyle().Foreground(colorError)
	boxStyle    = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 3)
)
