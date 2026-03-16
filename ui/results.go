package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/jpcummins/tsk-lib/model"
	"github.com/jpcummins/tsk-lib/search"
)

type resultsModel struct {
	tasks   []*model.Task
	matches []search.Match // populated for fuzzy results, nil for DSL results
	cursor  int
	focused bool
	width   int
	height  int
	offset  int // scroll offset for visible window
	errMsg  string
}

func newResultsModel() resultsModel {
	return resultsModel{}
}

func (m *resultsModel) setTasks(tasks []*model.Task) {
	m.tasks = tasks
	m.matches = nil
	m.cursor = 0
	m.offset = 0
	m.errMsg = ""
}

func (m *resultsModel) setMatches(matches []search.Match) {
	m.matches = matches
	m.tasks = make([]*model.Task, len(matches))
	for i, match := range matches {
		m.tasks[i] = match.Task
	}
	m.cursor = 0
	m.offset = 0
	m.errMsg = ""
}

func (m *resultsModel) setError(err string) {
	m.errMsg = err
	m.tasks = nil
	m.matches = nil
	m.cursor = 0
	m.offset = 0
}

func (m *resultsModel) clear() {
	m.tasks = nil
	m.matches = nil
	m.cursor = 0
	m.offset = 0
	m.errMsg = ""
}

func (m *resultsModel) focus() {
	m.focused = true
}

func (m *resultsModel) blur() {
	m.focused = false
}

func (m *resultsModel) setSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *resultsModel) moveUp() {
	if m.cursor > 0 {
		m.cursor--
		if m.cursor < m.offset {
			m.offset = m.cursor
		}
	}
}

func (m *resultsModel) moveDown() {
	if m.cursor < len(m.tasks)-1 {
		m.cursor++
		visibleRows := m.visibleRows()
		if m.cursor >= m.offset+visibleRows {
			m.offset = m.cursor - visibleRows + 1
		}
	}
}

func (m *resultsModel) selectedTask() *model.Task {
	if len(m.tasks) == 0 || m.cursor >= len(m.tasks) {
		return nil
	}
	return m.tasks[m.cursor]
}

func (m *resultsModel) selectedMatch() *search.Match {
	if m.matches == nil || m.cursor >= len(m.matches) {
		return nil
	}
	return &m.matches[m.cursor]
}

// Each result row takes 2 lines (path+meta, summary/snippet) + 1 blank = 3 lines.
const rowHeight = 3

func (m *resultsModel) visibleRows() int {
	if m.height <= 0 {
		return 5
	}
	rows := m.height / rowHeight
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (m resultsModel) view() string {
	if m.errMsg != "" {
		return "\n" + errorStyle.Render("  Error: "+m.errMsg) + "\n"
	}

	if len(m.tasks) == 0 {
		return "\n" + hintStyle.Render("  No results") + "\n"
	}

	var b strings.Builder

	header := hintStyle.Render(fmt.Sprintf("  %d result(s)", len(m.tasks)))
	if m.focused {
		header += hintStyle.Render("  (j/k to navigate, enter to view, tab to search)")
	}
	b.WriteString(header + "\n\n")

	visibleRows := m.visibleRows()
	end := m.offset + visibleRows
	if end > len(m.tasks) {
		end = len(m.tasks)
	}

	for i := m.offset; i < end; i++ {
		task := m.tasks[i]
		selected := m.focused && i == m.cursor

		var match *search.Match
		if m.matches != nil && i < len(m.matches) {
			match = &m.matches[i]
		}

		b.WriteString(m.renderRow(task, match, selected))
	}

	// Scroll indicators
	if m.offset > 0 {
		b.WriteString(hintStyle.Render("  ... more above") + "\n")
	}
	if end < len(m.tasks) {
		b.WriteString(hintStyle.Render("  ... more below") + "\n")
	}

	return b.String()
}

func (m resultsModel) renderRow(task *model.Task, match *search.Match, selected bool) string {
	maxWidth := m.width - 4
	if maxWidth < 20 {
		maxWidth = 20
	}

	pathStyle := resultPathStyle
	metaStyle := resultMetaStyle
	rowStyle := resultNormalStyle

	if selected {
		pathStyle = resultPathSelectedStyle
		metaStyle = resultMetaSelectedStyle
		rowStyle = resultSelectedStyle
	}

	// Line 1: path + status + assignee
	path := pathStyle.Render(string(task.CanonicalPath))

	var meta []string
	if task.Status != "" {
		meta = append(meta, task.Status)
	} else if task.StatusCategory != "" {
		meta = append(meta, string(task.StatusCategory))
	}
	if task.Assignee != "" {
		meta = append(meta, task.Assignee)
	}

	metaStr := ""
	if len(meta) > 0 {
		metaStr = metaStyle.Render("  [" + strings.Join(meta, " | ") + "]")
	}

	line1 := path + metaStr

	// Line 2: for fuzzy results show highlight snippet, for DSL show summary
	var line2 string
	if match != nil && len(match.Highlights) > 0 {
		line2 = renderHighlightSnippet(match.Highlights, maxWidth, selected)
	} else {
		summary := task.Summary
		if summary == "" && task.Body != "" {
			lines := strings.SplitN(strings.TrimSpace(task.Body), "\n", 2)
			if len(lines) > 0 {
				summary = lines[0]
			}
		}
		if len(summary) > maxWidth {
			summary = summary[:maxWidth-3] + "..."
		}
		if summary != "" {
			summaryStyle := resultSummaryStyle
			if selected {
				summaryStyle = resultSummarySelectedStyle
			}
			line2 = summaryStyle.Render("  " + summary)
		}
	}

	content := line1 + "\n" + line2
	row := rowStyle.Width(lipgloss.Width(content) + 2).Render(content)

	return row + "\n"
}

// renderHighlightSnippet produces a single-line snippet from the best
// highlight showing matched text in context with highlighting applied.
func renderHighlightSnippet(highlights []search.Highlight, maxWidth int, selected bool) string {
	// Pick the best highlight (they're already sorted by field weight).
	hl := highlights[0]

	if len(hl.Positions) == 0 {
		return ""
	}

	// Build a snippet around the first match position.
	firstPos := hl.Positions[0]
	text := hl.Text

	// Flatten to a single line for display.
	text = flattenToLine(text)

	// Recalculate position in the flattened text.
	// Since we replaced newlines with spaces (same byte length), positions are preserved.

	// Determine a context window around the first match.
	contextChars := 20
	snippetStart := firstPos.Start - contextChars
	if snippetStart < 0 {
		snippetStart = 0
	}
	snippetEnd := firstPos.End + contextChars
	if snippetEnd > len(text) {
		snippetEnd = len(text)
	}

	// Expand to word boundaries.
	for snippetStart > 0 && text[snippetStart] != ' ' {
		snippetStart--
	}
	for snippetEnd < len(text) && text[snippetEnd] != ' ' {
		snippetEnd++
	}

	snippet := text[snippetStart:snippetEnd]
	prefix := ""
	if snippetStart > 0 {
		prefix = "..."
	}
	suffix := ""
	if snippetEnd < len(text) {
		suffix = "..."
	}

	// Collect highlight positions from the same field that fall within
	// our snippet window. We only use positions from hl.Field since
	// positions from other fields refer to different text.
	var positions []search.Range
	for _, pos := range hl.Positions {
		// Adjust position relative to snippet.
		adjStart := pos.Start - snippetStart
		adjEnd := pos.End - snippetStart
		if adjEnd <= 0 || adjStart >= len(snippet) {
			continue
		}
		if adjStart < 0 {
			adjStart = 0
		}
		if adjEnd > len(snippet) {
			adjEnd = len(snippet)
		}
		positions = append(positions, search.Range{Start: adjStart, End: adjEnd})
	}

	// Sort and merge positions.
	positions = sortAndMerge(positions)

	// Render the snippet with highlights.
	fieldLabel := hintStyle.Render("  " + hl.Field + ": ")
	rendered := fieldLabel + prefix + renderTextWithHighlights(snippet, positions, selected) + suffix

	return rendered
}

// renderTextWithHighlights renders text with certain byte ranges highlighted.
func renderTextWithHighlights(text string, positions []search.Range, selected bool) string {
	if len(positions) == 0 {
		ctx := matchContextStyle
		if selected {
			ctx = resultSummarySelectedStyle
		}
		return ctx.Render(text)
	}

	var b strings.Builder
	prev := 0

	ctx := matchContextStyle
	hl := matchHighlightStyle
	if selected {
		ctx = resultSummarySelectedStyle
		hl = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true).Background(colorHighlightBg)
	}

	for _, pos := range positions {
		if pos.Start > prev {
			b.WriteString(ctx.Render(text[prev:pos.Start]))
		}
		if pos.Start < pos.End && pos.End <= len(text) {
			b.WriteString(hl.Render(text[pos.Start:pos.End]))
		}
		prev = pos.End
	}
	if prev < len(text) {
		b.WriteString(ctx.Render(text[prev:]))
	}

	return b.String()
}

// flattenToLine replaces newlines and consecutive whitespace with single spaces.
func flattenToLine(s string) string {
	// Replace all whitespace sequences with single space, preserving byte offsets
	// by replacing individual characters.
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' || s[i] == '\r' || s[i] == '\t' {
			result[i] = ' '
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}

// sortAndMerge sorts ranges by start and merges overlapping ones.
func sortAndMerge(positions []search.Range) []search.Range {
	if len(positions) <= 1 {
		return positions
	}

	// Simple insertion sort (positions are usually few).
	for i := 1; i < len(positions); i++ {
		for j := i; j > 0 && positions[j].Start < positions[j-1].Start; j-- {
			positions[j], positions[j-1] = positions[j-1], positions[j]
		}
	}

	merged := []search.Range{positions[0]}
	for _, r := range positions[1:] {
		last := &merged[len(merged)-1]
		if r.Start <= last.End {
			if r.End > last.End {
				last.End = r.End
			}
		} else {
			merged = append(merged, r)
		}
	}
	return merged
}
