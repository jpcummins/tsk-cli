package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/jpcummins/tsk-lib/model"
	"github.com/jpcummins/tsk-lib/search"
)

type detailModel struct {
	viewport   viewport.Model
	task       *model.Task
	highlights []search.Highlight // optional, for fuzzy search results
	width      int
	height     int
	ready      bool
}

func newDetailModel() detailModel {
	return detailModel{}
}

func (m *detailModel) setSize(width, height int) {
	m.width = width
	m.height = height
	if m.ready {
		m.viewport.SetWidth(width)
		m.viewport.SetHeight(height)
	}
}

func (m *detailModel) setTask(task *model.Task) {
	m.task = task
	m.highlights = nil
	m.viewport = viewport.New(
		viewport.WithWidth(m.width),
		viewport.WithHeight(m.height),
	)
	m.viewport.SetContent(m.renderContent())
	m.ready = true
}

func (m *detailModel) setTaskWithHighlights(task *model.Task, highlights []search.Highlight) {
	m.task = task
	m.highlights = highlights
	m.viewport = viewport.New(
		viewport.WithWidth(m.width),
		viewport.WithHeight(m.height),
	)
	m.viewport.SetContent(m.renderContent())
	m.ready = true
}

func (m detailModel) update(msg tea.Msg) (detailModel, tea.Cmd) {
	if !m.ready {
		return m, nil
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m detailModel) view() string {
	if !m.ready || m.task == nil {
		return ""
	}
	return m.viewport.View()
}

func (m detailModel) renderContent() string {
	if m.task == nil {
		return ""
	}

	task := m.task
	var b strings.Builder

	// Header with key fields
	b.WriteString(detailPathStyle.Render(string(task.Path)))
	b.WriteString("\n")

	var fields []string

	if task.Status != "" {
		fields = append(fields, fmt.Sprintf("Status: %s", task.Status))
	} else if task.Category != "" {
		fields = append(fields, fmt.Sprintf("Status: %s", task.Category))
	}

	if task.Assignee != "" {
		fields = append(fields, fmt.Sprintf("Assignee: %s", task.Assignee))
	}

	if task.Due != nil {
		fields = append(fields, fmt.Sprintf("Due: %s", task.Due.Format("2006-01-02")))
	}

	if task.CreatedAt != nil {
		fields = append(fields, fmt.Sprintf("Created: %s", task.CreatedAt.Format("2006-01-02")))
	}

	if task.Estimate != nil {
		fields = append(fields, fmt.Sprintf("Estimate: %s", task.Estimate.String()))
	}

	if len(task.Labels) > 0 {
		fields = append(fields, fmt.Sprintf("Labels: %s", strings.Join(task.Labels, ", ")))
	}

	if len(task.Dependencies) > 0 {
		deps := make([]string, len(task.Dependencies))
		for i, d := range task.Dependencies {
			deps[i] = string(d)
		}
		fields = append(fields, fmt.Sprintf("Dependencies: %s", strings.Join(deps, ", ")))
	}

	if len(fields) > 0 {
		for _, f := range fields {
			b.WriteString(detailFieldStyle.Render(f))
			b.WriteString("\n")
		}
	}

	// Separator
	sep := strings.Repeat("─", m.width-4)
	if m.width-4 < 10 {
		sep = strings.Repeat("─", 10)
	}
	b.WriteString(hintStyle.Render(sep))
	b.WriteString("\n\n")

	// Body — with optional highlight support
	if task.Body != "" {
		bodyHighlight := m.findBodyHighlight()
		if bodyHighlight != nil {
			b.WriteString(m.renderHighlightedBody(task.Body, bodyHighlight.Positions))
		} else {
			b.WriteString(task.Body)
		}
	} else {
		b.WriteString(hintStyle.Render("(no content)"))
	}

	b.WriteString("\n\n")
	b.WriteString(hintStyle.Render("esc to go back | j/k to browse results | e to edit"))
	b.WriteString("\n")

	return b.String()
}

// findBodyHighlight returns the "body" highlight if present, nil otherwise.
func (m detailModel) findBodyHighlight() *search.Highlight {
	for i := range m.highlights {
		if m.highlights[i].Field == "body" {
			return &m.highlights[i]
		}
	}
	return nil
}

// renderHighlightedBody renders the markdown body with matched positions highlighted.
func (m detailModel) renderHighlightedBody(body string, positions []search.Range) string {
	if len(positions) == 0 {
		return body
	}

	var b strings.Builder
	prev := 0

	for _, pos := range positions {
		start := pos.Start
		end := pos.End
		if start < 0 {
			start = 0
		}
		if end > len(body) {
			end = len(body)
		}
		if start >= end {
			continue
		}

		if start > prev {
			b.WriteString(body[prev:start])
		}

		b.WriteString(matchHighlightStyle.Render(body[start:end]))

		prev = end
	}

	if prev < len(body) {
		b.WriteString(body[prev:])
	}

	return b.String()
}
