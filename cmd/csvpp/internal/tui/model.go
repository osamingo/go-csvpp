package tui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"golang.design/x/clipboard"

	"github.com/osamingo/go-csvpp"
)

// Model represents the TUI model for viewing CSV++ data.
type Model struct {
	table    table.Model
	headers  []*csvpp.ColumnHeader
	records  [][]*csvpp.Field
	styles   Styles
	width    int
	height   int
	err      error
	selected map[int]bool
	copied   bool
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

	return Model{
		table:    t,
		headers:  headers,
		records:  records,
		styles:   styles,
		selected: make(map[int]bool),
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
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			// Clear selection
			m.selected = make(map[int]bool)
			m.copied = false
			m.updateRowMarkers()
			return m, nil
		case " ":
			// Toggle selection
			cursor := m.table.Cursor()
			if m.selected[cursor] {
				delete(m.selected, cursor)
			} else {
				m.selected[cursor] = true
			}
			m.copied = false
			m.updateRowMarkers()
			return m, nil
		case "y", "c":
			// Copy selected rows
			if len(m.selected) > 0 {
				m.copyToClipboard()
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(msg.Height - 4) // Leave room for status
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// updateRowMarkers updates the selection markers in the table rows.
func (m *Model) updateRowMarkers() {
	rows := m.table.Rows()
	for i := range rows {
		if m.selected[i] {
			rows[i][0] = "✓"
		} else {
			rows[i][0] = " "
		}
	}
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

	// Status line
	status := fmt.Sprintf("%d columns, %d records", len(m.headers), len(m.records))
	if len(m.selected) > 0 {
		status += fmt.Sprintf(" | %d selected", len(m.selected))
	}
	if m.copied {
		status += " | Copied!"
	}
	b.WriteString(m.styles.Status.Render(status))
	b.WriteString("\n")

	// Help
	help := "↑/↓: navigate • Space: select • y/c: copy • Esc: clear • q: quit"
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
