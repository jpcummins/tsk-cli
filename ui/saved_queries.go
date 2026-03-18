package ui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// savedQuery is a predefined query with a human-readable description.
type savedQuery struct {
	Query       string
	Description string
}

var predefinedQueries = []savedQuery{
	{"status.category = todo", "All tasks in todo"},
	{"status.category = in_progress", "All tasks currently in progress"},
	{"status.category != done", "All incomplete tasks"},
	{"assignee = me()", "Tasks assigned to me (requires --me)"},
	{"assignee = me() AND status.category != done", "My open tasks"},
	{"missing(assignee)", "Unassigned tasks"},
	{"missing(assignee) AND status.category = todo", "Unassigned todo tasks"},
	{"has(labels, \"bug\")", "Tasks labeled as bugs"},
	{"has(labels, \"bug\") AND status.category != done", "Open bugs"},
	{"due < date(\"tomorrow\") AND status.category != done", "Overdue or due today"},
	{"due < date(\"+7d\") AND status.category != done", "Due within a week"},
	{"missing(due) AND status.category != done", "Open tasks with no due date"},
	{"estimate > 2d AND missing(assignee)", "Large unassigned tasks"},
	{"exists(estimate) AND status.category = todo", "Estimated but not started"},
	{"path ~ \"backend\"", "Tasks under backend"},
	{"path ~ \"frontend\"", "Tasks under frontend"},
	{"iteration.status.category = in_progress", "Tasks in active iterations"},
	{"sla.status = \"breached\"", "Breached SLA tasks"},
	{"sla.status = \"at_risk\"", "At-risk SLA tasks"},
}

type savedQueriesModel struct {
	input    textinput.Model
	filtered []int // indices into predefinedQueries
	cursor   int
	width    int
	height   int
}

func newSavedQueriesModel() savedQueriesModel {
	ti := textinput.New()
	ti.Placeholder = "Filter saved queries..."
	ti.Prompt = ""
	ti.Focus()

	m := savedQueriesModel{
		input: ti,
	}
	m.applyFilter()
	return m
}

func (m *savedQueriesModel) setSize(width, height int) {
	m.width = width
	m.height = height
	innerWidth := width - 12
	if innerWidth < 20 {
		innerWidth = 20
	}
	m.input.SetWidth(innerWidth)
}

func (m *savedQueriesModel) applyFilter() {
	query := strings.ToLower(m.input.Value())
	m.filtered = nil

	for i, sq := range predefinedQueries {
		if query == "" ||
			strings.Contains(strings.ToLower(sq.Query), query) ||
			strings.Contains(strings.ToLower(sq.Description), query) {
			m.filtered = append(m.filtered, i)
		}
	}

	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *savedQueriesModel) moveUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

func (m *savedQueriesModel) moveDown() {
	if m.cursor < len(m.filtered)-1 {
		m.cursor++
	}
}

// selected returns the currently highlighted query, or nil if none.
func (m *savedQueriesModel) selected() *savedQuery {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	return &predefinedQueries[m.filtered[m.cursor]]
}

func (m *savedQueriesModel) reset() {
	m.input.SetValue("")
	m.cursor = 0
	m.applyFilter()
}

func (m savedQueriesModel) update(msg tea.Msg) (savedQueriesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Up):
			m.moveUp()
			return m, nil
		case key.Matches(msg, keys.Down):
			m.moveDown()
			return m, nil
		default:
			prev := m.input.Value()
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			if m.input.Value() != prev {
				m.applyFilter()
			}
			return m, cmd
		}
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m savedQueriesModel) view() string {
	contentWidth := m.width - 8
	if contentWidth < 40 {
		contentWidth = 40
	}
	contentHeight := m.height - 6
	if contentHeight < 10 {
		contentHeight = 10
	}

	title := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Render("Saved Queries")

	hint := hintStyle.Render("esc or ctrl+s to close | enter to select | j/k to navigate")

	// Filter input
	filterBox := searchBoxFocusedStyle.Width(contentWidth).Render(
		modeFuzzyStyle.Render("filter") + searchLabelStyle.Render(": ") + m.input.View(),
	)

	// Visible list area: total height minus title(1) + hint(1) + blank(1) + filterBox(3) + blank(1)
	listHeight := contentHeight - 7
	if listHeight < 3 {
		listHeight = 3
	}

	// Each item is 2 lines (query + description).
	itemHeight := 2
	visibleItems := listHeight / itemHeight
	if visibleItems < 1 {
		visibleItems = 1
	}

	// Compute scroll offset
	offset := 0
	if m.cursor >= visibleItems {
		offset = m.cursor - visibleItems + 1
	}

	var listBuf strings.Builder
	if len(m.filtered) == 0 {
		listBuf.WriteString(hintStyle.Render("  No matching queries"))
		listBuf.WriteString("\n")
	} else {
		end := offset + visibleItems
		if end > len(m.filtered) {
			end = len(m.filtered)
		}
		for vi := offset; vi < end; vi++ {
			idx := m.filtered[vi]
			sq := predefinedQueries[idx]
			selected := vi == m.cursor

			qStyle := sampleQueryStyle
			dStyle := sampleDescStyle
			if selected {
				qStyle = lipgloss.NewStyle().
					Foreground(colorHighlight).
					Background(colorHighlightBg).
					Bold(true)
				dStyle = lipgloss.NewStyle().
					Foreground(colorHighlight).
					Background(colorHighlightBg)
			}

			line1 := qStyle.Render("  " + sq.Query)
			line2 := dStyle.Render("    " + sq.Description)

			if selected {
				line1 = lipgloss.NewStyle().
					Background(colorHighlightBg).
					Width(contentWidth).
					Render(line1)
				line2 = lipgloss.NewStyle().
					Background(colorHighlightBg).
					Width(contentWidth).
					Render(line2)
			}

			listBuf.WriteString(line1 + "\n")
			listBuf.WriteString(line2 + "\n")
		}

		if offset > 0 {
			listBuf.WriteString(hintStyle.Render("  ... more above") + "\n")
		}
		if end < len(m.filtered) {
			listBuf.WriteString(hintStyle.Render("  ... more below") + "\n")
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		hint,
		"",
		filterBox,
		"",
		listBuf.String(),
	)

	bordered := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
