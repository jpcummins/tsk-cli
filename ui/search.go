package ui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// searchMode tracks whether the search box is in DSL or fuzzy mode.
type searchMode int

const (
	modeQuery searchMode = iota
	modeFuzzy
)

type searchModel struct {
	input   textinput.Model
	focused bool
	width   int
	mode    searchMode
}

func newSearchModel() searchModel {
	ti := textinput.New()
	ti.Placeholder = "Enter a tsk query..."
	ti.Prompt = ""
	ti.Focus()

	return searchModel{
		input:   ti,
		focused: true,
		mode:    modeQuery,
	}
}

func (m *searchModel) toggleMode() {
	if m.mode == modeQuery {
		m.mode = modeFuzzy
		m.input.Placeholder = "Fuzzy search all tasks..."
	} else {
		m.mode = modeQuery
		m.input.Placeholder = "Enter a tsk query..."
	}
}

func (m *searchModel) focus() tea.Cmd {
	m.focused = true
	return m.input.Focus()
}

func (m *searchModel) blur() {
	m.focused = false
	m.input.Blur()
}

func (m *searchModel) value() string {
	return m.input.Value()
}

func (m *searchModel) setValue(s string) {
	m.input.SetValue(s)
}

func (m *searchModel) setWidth(w int) {
	m.width = w
	// Account for border (2) + padding (2) + mode label (variable ~12-14 chars)
	innerWidth := w - 4 - 14
	if innerWidth < 10 {
		innerWidth = 10
	}
	m.input.SetWidth(innerWidth)
}

func (m searchModel) update(msg tea.Msg) (searchModel, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m searchModel) view() string {
	var label string
	if m.mode == modeFuzzy {
		label = modeFuzzyStyle.Render("fuzzy") + searchLabelStyle.Render(": ")
	} else {
		label = modeQueryStyle.Render("query") + searchLabelStyle.Render(": ")
	}
	input := m.input.View()
	content := label + input

	style := searchBoxStyle
	if m.focused {
		style = searchBoxFocusedStyle
	}

	boxWidth := m.width - 2 // Account for outer margin
	if boxWidth < 20 {
		boxWidth = 20
	}
	return style.Width(boxWidth).Render(content)
}
