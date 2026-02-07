package tui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"golang.design/x/clipboard"

	"github.com/osamingo/go-csvpp"
)

// filterQuery represents a parsed filter query.
type filterQuery struct {
	column string // empty means search all columns
	value  string // search string (lowercase)
}

// parseFilterQuery parses a filter query string.
// Supports "column:value" syntax for column-specific filtering.
// If the part before ":" is empty, it searches all columns.
func parseFilterQuery(s string) filterQuery {
	s = strings.TrimSpace(s)
	if s == "" {
		return filterQuery{}
	}

	if idx := strings.Index(s, ":"); idx >= 0 {
		col := strings.TrimSpace(s[:idx])
		val := strings.TrimSpace(s[idx+1:])
		if col != "" {
			return filterQuery{
				column: strings.ToLower(col),
				value:  strings.ToLower(val),
			}
		}
		// empty column part means search all columns
		return filterQuery{
			value: strings.ToLower(val),
		}
	}

	return filterQuery{
		value: strings.ToLower(s),
	}
}

// matchesFilter checks if a row matches the given filter query.
func matchesFilter(query filterQuery, headers []*csvpp.ColumnHeader, row table.Row) bool {
	if query.value == "" {
		return true
	}

	if query.column != "" {
		// Search specific column
		for i, h := range headers {
			if strings.ToLower(formatHeaderTitle(h)) == query.column {
				// row[0] is the selection marker, data starts at index 1
				if i+1 < len(row) && strings.Contains(strings.ToLower(row[i+1]), query.value) {
					return true
				}
				return false
			}
		}
		// Column not found
		return false
	}

	// Search all data columns (skip index 0 which is selection marker)
	for i := 1; i < len(row); i++ {
		if strings.Contains(strings.ToLower(row[i]), query.value) {
			return true
		}
	}
	return false
}

// Model represents the TUI model for viewing CSV++ data.
type Model struct {
	table    table.Model
	headers  []*csvpp.ColumnHeader
	records  [][]*csvpp.Field
	styles   Styles
	width    int
	height   int
	err      error
	selected map[int]bool // keyed by original record index
	copied   bool

	// Filter fields
	filtering   bool            // true when filter input is active
	filterInput textinput.Model // text input widget
	filterText  string          // committed filter text
	filteredIdx []int           // display position -> original record index
	allRows     []table.Row     // cache of all rows
}

// NewModel creates a new TUI model with the given data.
func NewModel(headers []*csvpp.ColumnHeader, records [][]*csvpp.Field) Model {
	styles := DefaultStyles()

	// Build table columns (first column is selection marker)
	columns := make([]table.Column, len(headers)+1)
	columns[0] = table.Column{Title: " ", Width: 2}
	for i, h := range headers {
		title := formatHeaderTitle(h)
		columns[i+1] = table.Column{
			Title: title,
			Width: max(len(title), 10),
		}
	}

	// Build table rows
	rows := make([]table.Row, len(records))
	for i, record := range records {
		row := make(table.Row, len(record)+1)
		row[0] = " " // Selection marker (empty initially)
		for j, field := range record {
			var header *csvpp.ColumnHeader
			if j < len(headers) {
				header = headers[j]
			}
			value := formatFieldValue(header, field)
			row[j+1] = value
			// Adjust column width
			if len(value) > columns[j+1].Width {
				columns[j+1].Width = min(len(value), 50) // Cap at 50
			}
		}
		rows[i] = row
	}

	// Create table
	t := table.New( //nostyle:funcfmt
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Apply styles
	s := table.DefaultStyles()
	s.Header = styles.Header
	s.Selected = styles.Selected
	s.Cell = styles.Cell
	t.SetStyles(s)

	// Initialize filter input
	fi := textinput.New()
	fi.Placeholder = "type to filter..."
	fi.CharLimit = 256

	// Build allRows cache
	allRows := make([]table.Row, len(rows))
	for i, row := range rows {
		r := make(table.Row, len(row))
		copy(r, row)
		allRows[i] = r
	}

	// Build initial filteredIdx (1:1 mapping)
	filteredIdx := make([]int, len(rows))
	for i := range filteredIdx {
		filteredIdx[i] = i
	}

	return Model{
		table:       t,
		headers:     headers,
		records:     records,
		styles:      styles,
		selected:    make(map[int]bool),
		filterInput: fi,
		filteredIdx: filteredIdx,
		allRows:     allRows,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { //nostyle:recvtype
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nostyle:recvtype
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filtering {
			return m.updateFilterMode(msg)
		}
		return m.updateNormalMode(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(msg.Height - 4) // Leave room for status
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// updateFilterMode handles key events when filter input is active.
func (m Model) updateFilterMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) { //nostyle:recvtype
	switch msg.Type {
	case tea.KeyEnter:
		// Commit filter
		m.filterText = m.filterInput.Value()
		m.filtering = false
		m.filterInput.Blur()
		m.table.Focus()
		m.applyFilter()
		return m, nil
	case tea.KeyEsc:
		// Cancel filter
		m.filtering = false
		m.filterInput.SetValue("")
		m.filterInput.Blur()
		m.table.Focus()
		m.clearFilter()
		return m, nil
	}

	// Forward key to textinput
	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)

	// Real-time filtering
	m.applyFilter()

	return m, cmd
}

// updateNormalMode handles key events in normal navigation mode.
func (m Model) updateNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) { //nostyle:recvtype
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "/":
		// Enter filter mode
		m.filtering = true
		m.filterInput.SetValue("")
		m.table.Blur()
		return m, m.filterInput.Focus()
	case "esc":
		if m.filterText != "" {
			// Clear active filter
			m.clearFilter()
		} else {
			// Clear selection
			m.selected = make(map[int]bool)
			m.copied = false
			m.rebuildRowMarkers()
		}
		return m, nil
	case " ":
		// Toggle selection using original index
		origIdx := m.originalIndex()
		if origIdx < 0 {
			return m, nil
		}
		if m.selected[origIdx] {
			delete(m.selected, origIdx)
		} else {
			m.selected[origIdx] = true
		}
		m.copied = false
		m.rebuildRowMarkers()
		return m, nil
	case "y", "c":
		// Copy selected rows
		if len(m.selected) > 0 {
			m.copyToClipboard()
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// originalIndex returns the original record index for the current cursor position.
// Returns -1 if the cursor position is out of range.
func (m *Model) originalIndex() int {
	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.filteredIdx) {
		return -1
	}
	return m.filteredIdx[cursor]
}

// rebuildRowMarkers updates selection markers in the currently displayed rows.
func (m *Model) rebuildRowMarkers() {
	rows := m.table.Rows()
	for i, origIdx := range m.filteredIdx {
		if i >= len(rows) {
			break
		}
		if m.selected[origIdx] {
			rows[i][0] = "✓"
		} else {
			rows[i][0] = " "
		}
	}
	m.table.SetRows(rows)
}

// applyFilter filters rows based on the current filter input value.
func (m *Model) applyFilter() {
	query := parseFilterQuery(m.filterInput.Value())

	if query.value == "" {
		// Show all rows without clearing committed filter state
		m.restoreAllRows()
		return
	}

	var filtered []table.Row
	var idx []int

	for i, row := range m.allRows {
		if matchesFilter(query, m.headers, row) {
			r := make(table.Row, len(row))
			copy(r, row)
			// Set selection marker
			if m.selected[i] {
				r[0] = "✓"
			} else {
				r[0] = " "
			}
			filtered = append(filtered, r)
			idx = append(idx, i)
		}
	}

	m.filteredIdx = idx
	m.table.SetRows(filtered)
	m.table.GotoTop()
}

// clearFilter resets the filter state and restores all rows.
func (m *Model) clearFilter() {
	m.filterText = ""
	m.restoreAllRows()
}

// restoreAllRows restores all rows to the table without modifying filter state.
func (m *Model) restoreAllRows() {
	rows := make([]table.Row, len(m.allRows))
	for i, row := range m.allRows {
		r := make(table.Row, len(row))
		copy(r, row)
		if m.selected[i] {
			r[0] = "✓"
		} else {
			r[0] = " "
		}
		rows[i] = r
	}

	filteredIdx := make([]int, len(m.allRows))
	for i := range filteredIdx {
		filteredIdx[i] = i
	}

	m.filteredIdx = filteredIdx
	m.table.SetRows(rows)
}

// copyToClipboard copies selected rows to clipboard in CSVPP format.
func (m *Model) copyToClipboard() {
	if err := clipboard.Init(); err != nil {
		m.err = fmt.Errorf("clipboard init: %w", err)
		return
	}

	var buf bytes.Buffer
	w := csvpp.NewWriter(&buf)
	w.SetHeaders(m.headers)

	// Write header
	if err := w.WriteHeader(); err != nil {
		m.err = fmt.Errorf("write header: %w", err)
		return
	}

	// Write selected records in order
	for i := 0; i < len(m.records); i++ {
		if m.selected[i] {
			if err := w.Write(m.records[i]); err != nil {
				m.err = fmt.Errorf("write record: %w", err)
				return
			}
		}
	}
	w.Flush()

	clipboard.Write(clipboard.FmtText, buf.Bytes())
	m.copied = true
}

// View implements tea.Model.
func (m Model) View() string { //nostyle:recvtype
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	var b strings.Builder

	// Table
	b.WriteString(m.table.View())
	b.WriteString("\n\n")

	// Filter line
	if m.filtering {
		b.WriteString(m.styles.FilterPrompt.Render("/"))
		b.WriteString(m.filterInput.View())
		b.WriteString("\n")
	} else if m.filterText != "" {
		b.WriteString(m.styles.FilterActive.Render(fmt.Sprintf("Filter: %s", m.filterText)))
		b.WriteString("\n")
	}

	// Status line
	status := fmt.Sprintf("%d columns, %d records", len(m.headers), len(m.records))
	if m.filterText != "" || m.filtering {
		status += fmt.Sprintf(" (%d shown)", len(m.filteredIdx))
	}
	if len(m.selected) > 0 {
		status += fmt.Sprintf(" | %d selected", len(m.selected))
	}
	if m.copied {
		status += " | Copied!"
	}
	b.WriteString(m.styles.Status.Render(status))
	b.WriteString("\n")

	// Help
	var help string
	if m.filtering {
		help = "Enter: apply filter • Esc: cancel • type to filter"
	} else if m.filterText != "" {
		help = "↑/↓: navigate • Space: select • y/c: copy • /: filter • Esc: clear filter • q: quit"
	} else {
		help = "↑/↓: navigate • Space: select • y/c: copy • /: filter • Esc: clear • q: quit"
	}
	b.WriteString(m.styles.Help.Render(help))

	return b.String()
}

// formatHeaderTitle formats a column header for display.
func formatHeaderTitle(h *csvpp.ColumnHeader) string {
	if h == nil {
		return ""
	}

	switch h.Kind {
	case csvpp.SimpleField:
		return h.Name
	case csvpp.ArrayField:
		return h.Name + "[]"
	case csvpp.StructuredField:
		comps := formatComponentNames(h.Components)
		return fmt.Sprintf("%s(%s)", h.Name, comps)
	case csvpp.ArrayStructuredField:
		comps := formatComponentNames(h.Components)
		return fmt.Sprintf("%s[](%s)", h.Name, comps)
	default:
		return h.Name
	}
}

// formatComponentNames formats component names for display.
func formatComponentNames(components []*csvpp.ColumnHeader) string {
	names := make([]string, len(components))
	for i, c := range components {
		names[i] = c.Name
	}
	return strings.Join(names, ",")
}

// formatFieldValue formats a field value for display.
func formatFieldValue(h *csvpp.ColumnHeader, f *csvpp.Field) string {
	if f == nil {
		return ""
	}

	if h == nil {
		return f.Value
	}

	switch h.Kind {
	case csvpp.SimpleField:
		return f.Value
	case csvpp.ArrayField:
		return strings.Join(f.Values, ", ")
	case csvpp.StructuredField:
		return formatStructuredValue(h.Components, f.Components)
	case csvpp.ArrayStructuredField:
		return formatArrayStructuredValue(h.Components, f.Components)
	default:
		return f.Value
	}
}

// formatStructuredValue formats a structured field value for display.
func formatStructuredValue(headers []*csvpp.ColumnHeader, components []*csvpp.Field) string {
	if len(components) == 0 {
		return ""
	}

	parts := make([]string, len(components))
	for i, c := range components {
		var name string
		if i < len(headers) {
			name = headers[i].Name
		} else {
			name = fmt.Sprintf("%d", i)
		}
		parts[i] = fmt.Sprintf("%s:%s", name, c.Value)
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// formatArrayStructuredValue formats an array structured field value for display.
func formatArrayStructuredValue(headers []*csvpp.ColumnHeader, components []*csvpp.Field) string {
	if len(components) == 0 {
		return ""
	}

	parts := make([]string, len(components))
	for i, item := range components {
		parts[i] = formatStructuredValue(headers, item.Components)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// PlainView returns a plain text representation without TUI.
func PlainView(headers []*csvpp.ColumnHeader, records [][]*csvpp.Field) string {
	var b strings.Builder

	// Print headers
	headerNames := make([]string, len(headers))
	for i, h := range headers {
		headerNames[i] = formatHeaderTitle(h)
	}
	b.WriteString(strings.Join(headerNames, "\t"))
	b.WriteString("\n")

	// Print separator
	b.WriteString(strings.Repeat("-", 40))
	b.WriteString("\n")

	// Print records
	for _, record := range records {
		values := make([]string, len(record))
		for i, field := range record {
			var header *csvpp.ColumnHeader
			if i < len(headers) {
				header = headers[i]
			}
			values[i] = formatFieldValue(header, field)
		}
		b.WriteString(strings.Join(values, "\t"))
		b.WriteString("\n")
	}

	return b.String()
}
