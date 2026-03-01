package ui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#38BDF8")).
			MarginBottom(1)

	tableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#0369A1"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B")).
			MarginTop(1)

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7DD3FC")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F87171")).
			Bold(true)

	stoppedRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#52525B"))

	pausedRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#92400E"))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#334155")).
			MarginLeft(2).
			MarginTop(1)

	confirmStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FCD34D")).
			Bold(true)

	confirmNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F8FAFC")).
				Bold(true)

	logsDividerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0369A1"))

	logsTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#38BDF8")).
			Bold(true)

	logsLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CBD5E1"))

	inspectSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A78BFA")).
				Bold(true)

	inspectValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CBD5E1"))
)
