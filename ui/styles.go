package ui

import "charm.land/lipgloss/v2"

// Colors — using plain hex colors that work well on dark terminals.
var (
	colorPrimary     = lipgloss.Color("#7B78EE")
	colorSecondary   = lipgloss.Color("#A0A0A0")
	colorMuted       = lipgloss.Color("#666666")
	colorHighlight   = lipgloss.Color("#FFFFFF")
	colorHighlightBg = lipgloss.Color("#5A56E0")
	colorError       = lipgloss.Color("#FF6666")
	colorBorder      = lipgloss.Color("#444444")
	colorFocusBorder = lipgloss.Color("#7B78EE")
)

// Search box styles.
var (
	searchBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	searchBoxFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorFocusBorder).
				Padding(0, 1)

	searchLabelStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)
)

// Results list styles.
var (
	resultNormalStyle = lipgloss.NewStyle().
				Padding(0, 1)

	resultSelectedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Foreground(colorHighlight).
				Background(colorHighlightBg).
				Bold(true)

	resultPathStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	resultPathSelectedStyle = lipgloss.NewStyle().
				Foreground(colorHighlight).
				Bold(true)

	resultMetaStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)

	resultMetaSelectedStyle = lipgloss.NewStyle().
				Foreground(colorHighlight)

	resultSummaryStyle = lipgloss.NewStyle().
				Foreground(colorMuted)

	resultSummarySelectedStyle = lipgloss.NewStyle().
					Foreground(colorHighlight)
)

// Detail view styles.
var (
	detailHeaderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(colorBorder).
				Padding(0, 1).
				MarginBottom(1)

	detailPathStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	detailFieldStyle = lipgloss.NewStyle().
				Foreground(colorSecondary)

	detailBodyStyle = lipgloss.NewStyle().
			Padding(0, 2)
)

// Help / empty state styles.
var (
	emptyTitleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			MarginBottom(1)

	sampleQueryStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	sampleDescStyle = lipgloss.NewStyle().
			Foreground(colorMuted)
)

// Match highlight style for fuzzy search results.
var (
	matchHighlightStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")).
				Bold(true)

	matchContextStyle = lipgloss.NewStyle().
				Foreground(colorMuted)
)

// Search mode indicator styles.
var (
	modeQueryStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	modeFuzzyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Bold(true)
)

// Status bar / hints.
var (
	hintStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)
)
