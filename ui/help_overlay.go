package ui

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type helpOverlayModel struct {
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

func newHelpOverlayModel() helpOverlayModel {
	return helpOverlayModel{}
}

func (m *helpOverlayModel) setSize(width, height int) {
	m.width = width
	m.height = height

	// Leave margin for the overlay border and title
	contentWidth := width - 8
	contentHeight := height - 6
	if contentWidth < 40 {
		contentWidth = 40
	}
	if contentHeight < 10 {
		contentHeight = 10
	}

	m.viewport = viewport.New(
		viewport.WithWidth(contentWidth),
		viewport.WithHeight(contentHeight),
	)
	m.viewport.SetContent(renderDSLReference(contentWidth))
	m.ready = true
}

func (m helpOverlayModel) update(msg tea.Msg) (helpOverlayModel, tea.Cmd) {
	if !m.ready {
		return m, nil
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m helpOverlayModel) view() string {
	if !m.ready {
		return ""
	}

	title := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Render("TSK Query Language Reference")

	hint := hintStyle.Render("esc or ctrl+h to close | j/k to scroll")

	header := lipgloss.JoinVertical(lipgloss.Left, title, hint)

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		m.viewport.View(),
	)

	// Border around the whole thing
	bordered := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Render(content)

	// Center it on screen
	overlay := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)

	return overlay
}

// renderDSLReference produces the help content for the DSL.
func renderDSLReference(width int) string {
	var b strings.Builder

	sectionStyle := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	codeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D787"))
	descStyle := lipgloss.NewStyle().Foreground(colorSecondary)

	section := func(title string) {
		b.WriteString("\n")
		b.WriteString(sectionStyle.Render(title))
		b.WriteString("\n\n")
	}

	item := func(code, desc string) {
		b.WriteString("  ")
		b.WriteString(codeStyle.Render(code))
		b.WriteString("  ")
		b.WriteString(descStyle.Render(desc))
		b.WriteString("\n")
	}

	example := func(query, desc string) {
		b.WriteString("  ")
		b.WriteString(codeStyle.Render(query))
		b.WriteString("\n")
		b.WriteString("    ")
		b.WriteString(descStyle.Render(desc))
		b.WriteString("\n\n")
	}

	section("OPERATORS")
	item("=", "Equality")
	item("!=", "Inequality")
	item("<, <=, >, >=", "Ordering (dates, numbers, durations)")
	item("~", "Contains (substring, case-insensitive)")
	item("IN", "Set membership (comma-separated list)")

	section("BOOLEAN LOGIC")
	item("AND", "Logical AND (higher precedence than OR)")
	item("OR", "Logical OR")
	item("NOT", "Negation (highest precedence)")
	item("( ... )", "Grouping")

	section("TASK FIELDS")
	item("status", "Custom status value")
	item("status.category", "Base category: todo, in_progress, done")
	item("assignee", "Email or team:name")
	item("due", "Due date (RFC3339)")
	item("date", "Creation date (RFC3339)")
	item("updated_at", "Last modified date (RFC3339)")
	item("estimate", "Estimate duration (e.g., 2h, 1.5d)")
	item("path", "Canonical path (e.g., launch/phase-1/cli)")
	item("summary", "Task summary text")
	item("dependency", "Matches any dependency")
	item("labels", "Matches any label")

	section("ITERATION FIELDS")
	item("iteration.team", "Team directory name")
	item("iteration.status", "Iteration custom status")
	item("iteration.status.category", "Iteration category")
	item("iteration.start", "Start date (RFC3339)")
	item("iteration.end", "End date (RFC3339)")
	item("iteration.path", "Iteration canonical path")

	section("SLA FIELDS (reporting only)")
	item("sla.id", "SLA rule id")
	item("sla.status", "Status: ok, at_risk, breached")
	item("sla.target", "Target duration")
	item("sla.elapsed", "Elapsed time")
	item("sla.remaining", "Remaining time")

	section("FUNCTIONS")
	item("exists(field)", "Field is present")
	item("missing(field)", "Field is absent")
	item("has(field, value)", "List membership (e.g., labels, dependency)")
	item("date(value)", "Convert relative date (e.g., \"today\", \"yesterday\")")
	item("me()", "Current user (requires --me flag)")
	item("my_team()", "All teams current user belongs to")
	item("team(name)", "Expand to team:name or any team member")

	section("EXAMPLES")
	example(
		`status.category = in_progress`,
		`All in-progress tasks`,
	)
	example(
		`assignee = me() AND status.category != done`,
		`My open tasks`,
	)
	example(
		`has(labels, "bug") AND status.category = todo`,
		`Unfixed bugs`,
	)
	example(
		`due < date("tomorrow") AND status.category != done`,
		`Overdue or due soon`,
	)
	example(
		`path ~ "backend" AND assignee = team("platform")`,
		`Backend tasks for platform team`,
	)
	example(
		`estimate > 2d AND missing(assignee)`,
		`Large unassigned tasks`,
	)
	example(
		`iteration.status.category = in_progress AND iteration.start <= date("today")`,
		`Tasks in active iterations`,
	)
	example(
		`sla.status = "breached"`,
		`All breached SLA tasks`,
	)
	example(
		`sla.id = "security-30d" AND sla.status = "breached"`,
		`Breached security SLA tasks`,
	)

	b.WriteString("\n")
	b.WriteString(descStyle.Render("For full spec, see SPEC.md section 12.1"))
	b.WriteString("\n")

	return b.String()
}
