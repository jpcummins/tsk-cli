package ui

import (
	"strings"
)

func renderEmptyState(width int, fuzzy bool) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(emptyTitleStyle.Render("  Keyboard Shortcuts"))
	b.WriteString("\n\n")

	shortcut := func(key, desc string) {
		b.WriteString("  ")
		b.WriteString(sampleQueryStyle.Render(key))
		padding := 16 - len(key)
		if padding < 2 {
			padding = 2
		}
		b.WriteString(strings.Repeat(" ", padding))
		b.WriteString(sampleDescStyle.Render(desc))
		b.WriteString("\n")
	}

	shortcut("enter", "Execute query")
	shortcut("esc", "Clear / go back / quit")
	shortcut("tab", "Switch focus between search and results")
	shortcut("j / k", "Navigate up / down")
	b.WriteString("\n")
	shortcut("e", "Edit task in $EDITOR (detail view)")
	b.WriteString("\n")
	shortcut("ctrl+t", "Toggle query / fuzzy mode")
	shortcut("ctrl+s", "Saved queries")
	shortcut("ctrl+h", "Query language reference")
	shortcut("ctrl+c", "Quit")

	b.WriteString("\n")

	if fuzzy {
		b.WriteString(hintStyle.Render("  Type to search (results update live)"))
	} else {
		b.WriteString(hintStyle.Render("  Type a query and press enter to search"))
	}
	b.WriteString("\n")

	return b.String()
}
