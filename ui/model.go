package ui

import (
	"os"
	"os/exec"
	"path/filepath"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/jpcummins/tsk-lib/engine"
	"github.com/jpcummins/tsk-lib/model"
)

// editorFinishedMsg is sent when the external editor exits.
type editorFinishedMsg struct {
	err      error
	taskPath model.CanonicalPath
}

// viewState tracks which screen the user sees.
type viewState int

const (
	stateEmpty   viewState = iota // No query yet, show sample queries
	stateResults                  // Query results are displayed
	stateDetail                   // Viewing a single task
)

// focusArea tracks which component has keyboard focus.
type focusArea int

const (
	focusSearch focusArea = iota
	focusResults
)

// Model is the top-level Bubble Tea model for tsk-cli.
type Model struct {
	engine   *engine.Engine
	repoRoot string // absolute path to the repository root

	state viewState
	focus focusArea

	search       searchModel
	results      resultsModel
	detail       detailModel
	helpOverlay  helpOverlayModel
	savedQueries savedQueriesModel

	showHelp         bool
	showSavedQueries bool

	width  int
	height int

	lastQuery string
}

// NewModel creates the root model. The engine must already be initialized.
// repoRoot is the absolute path to the tsk repository root.
func NewModel(eng *engine.Engine, repoRoot string) Model {
	return Model{
		engine:       eng,
		repoRoot:     repoRoot,
		state:        stateEmpty,
		focus:        focusSearch,
		search:       newSearchModel(),
		results:      newResultsModel(),
		detail:       newDetailModel(),
		helpOverlay:  newHelpOverlayModel(),
		savedQueries: newSavedQueriesModel(),
		showHelp:     false,
	}
}

func (m Model) Init() tea.Cmd {
	return m.search.input.Focus()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case editorFinishedMsg:
		// Re-index the repository after the editor exits.
		m.engine.Index(m.repoRoot)

		// Refresh the current task in the detail view.
		if m.state == stateDetail && m.detail.task != nil {
			refreshed, err := m.engine.TaskByPath(msg.taskPath)
			if err == nil && refreshed != nil {
				m.detail.setTask(refreshed)
			}
		}

		// Also re-run the last query to refresh results.
		if m.lastQuery != "" {
			if m.search.mode == modeFuzzy {
				matches, err := m.engine.Search(m.lastQuery)
				if err == nil {
					m.results.setMatches(matches)
				}
			} else {
				tasks, _, err := m.engine.Query(m.lastQuery)
				if err == nil {
					m.results.setTasks(tasks)
				}
			}
		}

		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.search.setWidth(msg.Width)
		// Results/detail get remaining height after search box (approx 3 lines for border+padding)
		contentHeight := msg.Height - 4
		if contentHeight < 5 {
			contentHeight = 5
		}
		m.results.setSize(msg.Width, contentHeight)
		m.detail.setSize(msg.Width, contentHeight)
		m.helpOverlay.setSize(msg.Width, msg.Height)
		m.savedQueries.setSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyPressMsg:
		// If saved queries overlay is showing, handle it first
		if m.showSavedQueries {
			return m.updateSavedQueries(msg)
		}

		// If help overlay is showing, handle it first
		if m.showHelp {
			return m.updateHelpOverlay(msg)
		}

		// Global quit
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}

		// Global help toggle (only in query mode)
		if key.Matches(msg, keys.ToggleHelp) && m.search.mode == modeQuery {
			m.showHelp = true
			return m, nil
		}

		// Global saved queries toggle
		if key.Matches(msg, keys.SavedQueries) {
			m.showSavedQueries = true
			m.savedQueries.reset()
			m.savedQueries.input.Focus()
			return m, nil
		}

		switch m.state {
		case stateDetail:
			return m.updateDetail(msg)
		case stateResults:
			return m.updateResults(msg)
		default:
			return m.updateEmpty(msg)
		}
	}

	// If saved queries overlay is showing, route to it
	if m.showSavedQueries {
		var cmd tea.Cmd
		m.savedQueries, cmd = m.savedQueries.update(msg)
		return m, cmd
	}

	// If help overlay is showing, route to it
	if m.showHelp {
		var cmd tea.Cmd
		m.helpOverlay, cmd = m.helpOverlay.update(msg)
		return m, cmd
	}

	// Pass through to focused component
	switch m.focus {
	case focusSearch:
		var cmd tea.Cmd
		m.search, cmd = m.search.update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m Model) updateHelpOverlay(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape), key.Matches(msg, keys.ToggleHelp):
		m.showHelp = false
		return m, nil
	default:
		var cmd tea.Cmd
		m.helpOverlay, cmd = m.helpOverlay.update(msg)
		return m, cmd
	}
}

func (m Model) updateSavedQueries(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape), key.Matches(msg, keys.SavedQueries):
		m.showSavedQueries = false
		return m, nil
	case key.Matches(msg, keys.Enter):
		sq := m.savedQueries.selected()
		if sq == nil {
			return m, nil
		}
		m.showSavedQueries = false

		// Switch to query mode and populate the search box
		m.search.mode = modeQuery
		m.search.input.Placeholder = "Enter a tsk query..."
		m.search.setValue(sq.Query)

		// Execute the query
		return m.executeDSLSearch(sq.Query)
	default:
		var cmd tea.Cmd
		m.savedQueries, cmd = m.savedQueries.update(msg)
		return m, cmd
	}
}

func (m Model) updateEmpty(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.ToggleMode):
		m.search.toggleMode()
		// If switching to fuzzy with existing text, run search immediately
		query := m.search.value()
		if query != "" && m.search.mode == modeFuzzy {
			m.lastQuery = query
			matches, err := m.engine.Search(query)
			if err != nil {
				m.results.setError(err.Error())
			} else {
				m.results.setMatches(matches)
			}
			m.results.blur()
			m.state = stateResults
		}
		return m, nil
	case key.Matches(msg, keys.Enter):
		return m.executeSearch()
	case key.Matches(msg, keys.Escape):
		// Quit if search is empty
		if m.search.value() == "" {
			return m, tea.Quit
		}
		m.search.setValue("")
		return m, nil
	default:
		prev := m.search.value()
		var cmd tea.Cmd
		m.search, cmd = m.search.update(msg)
		return m.maybeLiveFuzzy(prev, cmd)
	}
}

func (m Model) updateResults(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.ToggleMode):
		if m.focus == focusSearch {
			m.search.toggleMode()
			// Re-run search in new mode if there's a query
			query := m.search.value()
			if query != "" && m.search.mode == modeFuzzy {
				m.lastQuery = query
				matches, err := m.engine.Search(query)
				if err != nil {
					m.results.setError(err.Error())
				} else {
					m.results.setMatches(matches)
				}
				m.results.blur()
				m.state = stateResults
			} else if query == "" {
				m.results.clear()
				m.state = stateEmpty
			} else {
				// Switched to DSL mode; clear fuzzy results, user must press Enter
				m.results.clear()
				m.state = stateEmpty
			}
			return m, nil
		}
		return m, nil

	case key.Matches(msg, keys.Tab):
		// Toggle focus
		if m.focus == focusSearch {
			m.search.blur()
			m.results.focus()
			m.focus = focusResults
		} else {
			m.results.blur()
			cmd := m.search.focus()
			m.focus = focusSearch
			return m, cmd
		}
		return m, nil

	case key.Matches(msg, keys.Enter):
		if m.focus == focusSearch {
			return m.executeSearch()
		}
		// Open selected task
		return m.openSelectedTask()

	case key.Matches(msg, keys.Escape):
		if m.focus == focusResults {
			// Go back to search focus
			m.results.blur()
			cmd := m.search.focus()
			m.focus = focusSearch
			return m, cmd
		}
		// Clear results, go back to empty state
		m.results.clear()
		m.state = stateEmpty
		m.lastQuery = ""
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.focus == focusResults {
			m.results.moveUp()
			return m, nil
		}
		var cmd tea.Cmd
		m.search, cmd = m.search.update(msg)
		return m, cmd

	case key.Matches(msg, keys.Down):
		if m.focus == focusResults {
			m.results.moveDown()
			return m, nil
		}
		var cmd tea.Cmd
		m.search, cmd = m.search.update(msg)
		return m, cmd

	default:
		if m.focus == focusSearch {
			prev := m.search.value()
			var cmd tea.Cmd
			m.search, cmd = m.search.update(msg)
			return m.maybeLiveFuzzy(prev, cmd)
		}
		return m, nil
	}
}

func (m Model) updateDetail(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.state = stateResults
		return m, nil

	case key.Matches(msg, keys.Edit):
		return m.editCurrentTask()

	case key.Matches(msg, keys.Up):
		if m.results.cursor > 0 {
			m.results.cursor--
			m.openCurrentTask()
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		if m.results.cursor < len(m.results.tasks)-1 {
			m.results.cursor++
			m.openCurrentTask()
		}
		return m, nil

	default:
		var cmd tea.Cmd
		m.detail, cmd = m.detail.update(msg)
		return m, cmd
	}
}

// editCurrentTask opens the current task in $EDITOR.
func (m Model) editCurrentTask() (tea.Model, tea.Cmd) {
	task := m.detail.task
	if task == nil {
		return m, nil
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	// Resolve canonical path to filesystem path.
	var filePath string
	if task.IsReadme {
		filePath = filepath.Join(m.repoRoot, "tasks", string(task.Path), "README.md")
	} else {
		filePath = filepath.Join(m.repoRoot, "tasks", string(task.Path)+".md")
	}

	taskPath := task.Path
	c := exec.Command(editor, filePath)
	cmd := tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err, taskPath: taskPath}
	})

	return m, cmd
}

// openSelectedTask opens the currently selected task in the detail view.
func (m Model) openSelectedTask() (tea.Model, tea.Cmd) {
	task := m.results.selectedTask()
	if task == nil {
		return m, nil
	}
	match := m.results.selectedMatch()
	if match != nil {
		m.detail.setTaskWithHighlights(task, match.Highlights)
	} else {
		m.detail.setTask(task)
	}
	m.state = stateDetail
	return m, nil
}

// openCurrentTask updates the detail view for the current cursor position.
// Used when navigating with j/k in detail view.
func (m *Model) openCurrentTask() {
	task := m.results.selectedTask()
	if task == nil {
		return
	}
	match := m.results.selectedMatch()
	if match != nil {
		m.detail.setTaskWithHighlights(task, match.Highlights)
	} else {
		m.detail.setTask(task)
	}
}

func (m Model) executeSearch() (tea.Model, tea.Cmd) {
	query := m.search.value()
	if query == "" {
		return m, nil
	}

	m.lastQuery = query

	if m.search.mode == modeFuzzy {
		return m.executeFuzzySearch(query)
	}
	return m.executeDSLSearch(query)
}

func (m Model) executeDSLSearch(query string) (tea.Model, tea.Cmd) {
	tasks, diags, err := m.engine.Query(query)
	if err != nil {
		m.results.setError(err.Error())
		m.state = stateResults
		return m, nil
	}

	if diags.HasErrors() {
		m.results.setError(diags.Errors()[0].Message)
		m.state = stateResults
		return m, nil
	}

	m.results.setTasks(tasks)
	m.state = stateResults

	// Auto-focus results after search
	m.search.blur()
	m.results.focus()
	m.focus = focusResults

	return m, nil
}

func (m Model) executeFuzzySearch(query string) (tea.Model, tea.Cmd) {
	matches, err := m.engine.Search(query)
	if err != nil {
		m.results.setError(err.Error())
		m.state = stateResults
		return m, nil
	}

	m.results.setMatches(matches)
	m.state = stateResults

	// Auto-focus results after search
	m.search.blur()
	m.results.focus()
	m.focus = focusResults

	return m, nil
}

// maybeLiveFuzzy checks if the search text changed and, in fuzzy mode,
// re-runs the search immediately. Focus stays on the search box.
func (m Model) maybeLiveFuzzy(prevQuery string, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	if m.search.mode != modeFuzzy {
		return m, cmd
	}

	query := m.search.value()
	if query == prevQuery {
		return m, cmd
	}

	if query == "" {
		m.results.clear()
		m.state = stateEmpty
		m.lastQuery = ""
		return m, cmd
	}

	m.lastQuery = query
	matches, err := m.engine.Search(query)
	if err != nil {
		m.results.setError(err.Error())
	} else {
		m.results.setMatches(matches)
	}
	m.results.blur()
	m.state = stateResults
	return m, cmd
}

func (m Model) View() tea.View {
	var content string

	searchView := m.search.view()

	switch m.state {
	case stateDetail:
		content = searchView + "\n" + m.detail.view()
	case stateResults:
		content = searchView + "\n" + m.results.view()
	default:
		content = searchView + "\n" + renderEmptyState(m.width, m.search.mode == modeFuzzy)
	}

	// If an overlay is active, render it on top
	if m.showSavedQueries {
		content = m.savedQueries.view()
	} else if m.showHelp {
		content = m.helpOverlay.view()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
