package tui

import "github.com/charmbracelet/lipgloss"

// contentWidth is the fixed column width quiz content wraps to, regardless
// of how wide the actual terminal is -- the outer app centers this block
// rather than letting prose stretch edge to edge.
const contentWidth = 76

var (
	colorAccent  = lipgloss.Color("#7D56F4")
	colorGood    = lipgloss.Color("#3ECF8E")
	colorBad     = lipgloss.Color("#E86671")
	colorWarn    = lipgloss.Color("#E8C547")
	colorMuted   = lipgloss.Color("#6E6E6E")
	colorFaint   = lipgloss.Color("#3C3C3C")
	colorFgOnAcc = lipgloss.Color("#FFFFFF")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorFgOnAcc).
			Background(colorAccent).
			Padding(0, 1)

	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Width(contentWidth).
			MarginBottom(1)

	explanationStyle = lipgloss.NewStyle().
				Width(contentWidth).
				MarginTop(1)

	helpStyle = lipgloss.NewStyle().Foreground(colorMuted)

	selectedStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	checkedBoxStyle = lipgloss.NewStyle().Foreground(colorGood).Bold(true)

	correctStyle   = lipgloss.NewStyle().Foreground(colorGood).Bold(true)
	incorrectStyle = lipgloss.NewStyle().Foreground(colorBad).Bold(true)
	typoWarnStyle  = lipgloss.NewStyle().Foreground(colorWarn).Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorFaint).
			Padding(1, 2)
)
