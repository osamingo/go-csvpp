package csvpp

import (
	"encoding/csv"
	"io"
	"strings"
)

// Writer writes CSV++ files according to the IETF CSV++ specification.
// It wraps encoding/csv.Writer and serializes CSV++ fields using the delimiters
// defined in the headers. The output is RFC 4180 compliant.
type Writer struct {
	// Comma is the field delimiter (default: ',').
	Comma rune
	// UseCRLF uses \r\n as the line terminator if true.
	UseCRLF bool

	w         io.Writer
	csvWriter *csv.Writer
	headers   []*ColumnHeader
}

// NewWriter creates a new Writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Comma: ',',
		w:     w,
	}
}

// SetHeaders sets the header information.
// This must be called before WriteHeader or Write.
func (w *Writer) SetHeaders(headers []*ColumnHeader) {
	w.headers = headers
}

// WriteHeader writes the header row.
func (w *Writer) WriteHeader() error {
	w.ensureWriter()

	if len(w.headers) == 0 {
		return ErrNoHeader
	}

	record := make([]string, len(w.headers))
	for i, h := range w.headers {
		record[i] = formatColumnHeader(h)
	}

	return w.csvWriter.Write(record)
}

// Write writes one record's worth of fields.
func (w *Writer) Write(record []*Field) error {
	w.ensureWriter()

	row := make([]string, len(record))
	for i, field := range record {
		var header *ColumnHeader
		if i < len(w.headers) {
			header = w.headers[i]
		}
		row[i] = w.formatField(header, field)
	}

	return w.csvWriter.Write(row)
}

// WriteAll writes all records.
// The header row is also written automatically.
func (w *Writer) WriteAll(records [][]*Field) error {
	if err := w.WriteHeader(); err != nil {
		return err
	}

	for _, record := range records {
		if err := w.Write(record); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}

// Flush flushes the buffer.
func (w *Writer) Flush() {
	if w.csvWriter != nil {
		w.csvWriter.Flush()
	}
}

// Error returns any error that occurred during writing.
func (w *Writer) Error() error {
	if w.csvWriter != nil {
		return w.csvWriter.Error()
	}
	return nil
}

// ensureWriter initializes the csv.Writer.
func (w *Writer) ensureWriter() {
	if w.csvWriter == nil {
		w.csvWriter = csv.NewWriter(w.w)
		w.csvWriter.Comma = w.Comma
		w.csvWriter.UseCRLF = w.UseCRLF
	}
}

// formatColumnHeader converts a ColumnHeader to its IETF CSV++ string representation.
// The format follows the ABNF grammar in Section 2.2 of the specification.
func formatColumnHeader(h *ColumnHeader) string {
	var sb strings.Builder
	sb.WriteString(h.Name)

	switch h.Kind {
	case SimpleField:
		// nothing to add
	case ArrayField:
		sb.WriteRune('[')
		if h.ArrayDelimiter != DefaultArrayDelimiter {
			sb.WriteRune(h.ArrayDelimiter)
		}
		sb.WriteRune(']')
	case StructuredField:
		if h.ComponentDelimiter != DefaultComponentDelimiter {
			sb.WriteRune(h.ComponentDelimiter)
		}
		sb.WriteRune('(')
		sb.WriteString(formatComponentList(h.Components, h.ComponentDelimiter))
		sb.WriteRune(')')
	case ArrayStructuredField:
		sb.WriteRune('[')
		if h.ArrayDelimiter != DefaultArrayDelimiter {
			sb.WriteRune(h.ArrayDelimiter)
		}
		sb.WriteRune(']')
		if h.ComponentDelimiter != DefaultComponentDelimiter {
			sb.WriteRune(h.ComponentDelimiter)
		}
		sb.WriteRune('(')
		sb.WriteString(formatComponentList(h.Components, h.ComponentDelimiter))
		sb.WriteRune(')')
	}

	return sb.String()
}

// formatComponentList converts a component list to a string.
func formatComponentList(components []*ColumnHeader, delim rune) string {
	if len(components) == 0 {
		return ""
	}

	parts := make([]string, len(components))
	for i, comp := range components {
		parts[i] = formatColumnHeader(comp)
	}

	return strings.Join(parts, string(delim))
}

// formatField converts a Field to a string.
func (w *Writer) formatField(header *ColumnHeader, field *Field) string {
	if header == nil {
		return field.Value
	}

	switch header.Kind {
	case SimpleField:
		return field.Value
	case ArrayField:
		return w.formatArrayField(header, field)
	case StructuredField:
		return w.formatStructuredField(header, field)
	case ArrayStructuredField:
		return w.formatArrayStructuredField(header, field)
	default:
		return field.Value
	}
}

// formatArrayField converts an array field to a string.
func (w *Writer) formatArrayField(header *ColumnHeader, field *Field) string {
	if len(field.Values) == 0 {
		return ""
	}
	return strings.Join(field.Values, string(header.ArrayDelimiter))
}

// formatStructuredField converts a structured field to a string.
func (w *Writer) formatStructuredField(header *ColumnHeader, field *Field) string {
	if len(field.Components) == 0 {
		return ""
	}
	return w.formatComponents(header.Components, header.ComponentDelimiter, field.Components)
}

// formatArrayStructuredField converts an array structured field to a string.
func (w *Writer) formatArrayStructuredField(header *ColumnHeader, field *Field) string {
	if len(field.Components) == 0 {
		return ""
	}

	parts := make([]string, len(field.Components))
	for i, comp := range field.Components {
		parts[i] = w.formatComponents(header.Components, header.ComponentDelimiter, comp.Components)
	}

	return strings.Join(parts, string(header.ArrayDelimiter))
}

// formatComponents converts a component list to a string (recursive).
func (w *Writer) formatComponents(headers []*ColumnHeader, delim rune, components []*Field) string {
	if len(components) == 0 {
		return ""
	}

	parts := make([]string, len(components))
	for i, comp := range components {
		var header *ColumnHeader
		if i < len(headers) {
			header = headers[i]
		}
		parts[i] = w.formatField(header, comp)
	}

	return strings.Join(parts, string(delim))
}
