package ui

import (
	"fmt"
	"strings"
)

type sampleQuery struct {
	Query       string
	Description string
}

var sampleQueries = []sampleQuery{
	{"status = todo", "Find all tasks in todo"},
	{"status.category = in_progress", "Tasks currently in progress"},
	{"has(labels, \"bug\")", "Tasks labeled as bugs"},
	{"assignee = me()", "Tasks assigned to you (requires --me)"},
	{"due < date(\"tomorrow\")", "Tasks due soon"},
	{"status.category != done", "All incomplete tasks"},
	{"path ~ \"backend\"", "Tasks under the backend path"},
}

var sampleFuzzyQueries = []sampleQuery{
	{"payment timeout", "Find tasks mentioning payment timeout"},
	{"cart bug", "Tasks related to cart bugs"},
	{"alice", "Tasks mentioning alice anywhere"},
}

func renderEmptyState(width int, fuzzy bool) string {
	var b strings.Builder

	b.WriteString("\n")

	if fuzzy {
		b.WriteString(emptyTitleStyle.Render("Fuzzy search"))
		b.WriteString("\n\n")

		for _, sq := range sampleFuzzyQueries {
			query := sampleQueryStyle.Render(sq.Query)
			desc := sampleDescStyle.Render(fmt.Sprintf("  %s", sq.Description))
			b.WriteString(fmt.Sprintf("  %s\n%s\n\n", query, desc))
		}
	} else {
		b.WriteString(emptyTitleStyle.Render("Try a query"))
		b.WriteString("\n\n")

		for _, sq := range sampleQueries {
			query := sampleQueryStyle.Render(sq.Query)
			desc := sampleDescStyle.Render(fmt.Sprintf("  %s", sq.Description))
			b.WriteString(fmt.Sprintf("  %s\n%s\n\n", query, desc))
		}
	}

	b.WriteString("\n")
	if fuzzy {
		b.WriteString(hintStyle.Render("Type to search (live) | ctrl+t to toggle mode"))
	} else {
		b.WriteString(hintStyle.Render("Type a query and press Enter | ctrl+t to toggle | ctrl+h for help"))
	}
	b.WriteString("\n")

	return b.String()
}
