package csvpp

import (
	"encoding/csv"
	"io"
)

// Reader reads CSV++ files according to the IETF CSV++ specification.
// It wraps encoding/csv.Reader and provides CSV++ header parsing and field parsing.
// The first row is always treated as the header row (IETF Section 2.1).
type Reader struct {
	// Comma is the field delimiter (default: ',').
	Comma rune
	// Comment is the comment character (disabled if 0).
	Comment rune
	// LazyQuotes relaxes strict quote checking if true.
	LazyQuotes bool
	// TrimLeadingSpace trims leading whitespace from fields if true.
	TrimLeadingSpace bool
	// MaxNestingDepth is the maximum nesting depth for structured fields (default: 10).
	// This limit prevents stack overflow from deeply nested input (IETF Section 5).
	// If 0, DefaultMaxNestingDepth is used.
	MaxNestingDepth int

	r             io.Reader
	csvReader     *csv.Reader
	headers       []*ColumnHeader
	headersParsed bool
	line          int // Current line number (1-based)
}

// NewReader creates a new Reader.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		Comma: ',',
		r:     r,
	}
}

// Headers returns the parsed header information.
// If headers have not been parsed yet, the first row is read and parsed.
func (r *Reader) Headers() ([]*ColumnHeader, error) {
	if err := r.ensureHeaders(); err != nil {
		return nil, err
	}
	return r.headers, nil
}

// Read reads and returns one record's worth of fields.
// The header row is automatically parsed on the first call.
// Returns io.EOF when the end of file is reached.
func (r *Reader) Read() ([]*Field, error) {
	if err := r.ensureHeaders(); err != nil {
		return nil, err
	}

	r.line++
	record, err := r.csvReader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, &ParseError{Line: r.line, Err: err}
	}

	fields, err := r.parseRecord(record)
	if err != nil {
		return nil, err // parseRecord already returns ParseError
	}

	return fields, nil
}

// ReadAll reads and returns all records.
// The header row is automatically parsed on the first call.
func (r *Reader) ReadAll() ([][]*Field, error) {
	if err := r.ensureHeaders(); err != nil {
		return nil, err
	}

	records, err := r.csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	result := make([][]*Field, 0, len(records))
	for _, record := range records {
		fields, err := r.parseRecord(record)
		if err != nil {
			return nil, err
		}
		result = append(result, fields)
	}

	return result, nil
}

// ensureHeaders ensures that headers have been parsed.
func (r *Reader) ensureHeaders() error {
	if r.headersParsed {
		return nil
	}

	// Initialize csv.Reader
	r.csvReader = csv.NewReader(r.r)
	r.csvReader.Comma = r.Comma
	r.csvReader.Comment = r.Comment
	r.csvReader.LazyQuotes = r.LazyQuotes
	r.csvReader.TrimLeadingSpace = r.TrimLeadingSpace

	// Read header row
	r.line = 1
	headerRow, err := r.csvReader.Read()
	if err != nil {
		if err == io.EOF {
			return ErrNoHeader
		}
		return &ParseError{Line: r.line, Err: err}
	}

	// Parse headers
	maxDepth := r.MaxNestingDepth
	if maxDepth == 0 {
		maxDepth = DefaultMaxNestingDepth
	}
	headers, err := parseHeaderRecordWithMaxDepth(headerRow, maxDepth)
	if err != nil {
		return &ParseError{Line: r.line, Err: err}
	}

	r.headers = headers
	r.headersParsed = true

	return nil
}

// parseRecord parses a data row and converts it to []*Field.
func (r *Reader) parseRecord(record []string) ([]*Field, error) {
	fields := make([]*Field, len(record))

	for i, value := range record {
		field, err := r.parseField(i, value)
		if err != nil {
			// Get field name
			fieldName := ""
			if i < len(r.headers) {
				fieldName = r.headers[i].Name
			}
			return nil, &ParseError{
				Line:   r.line,
				Column: i + 1,
				Field:  fieldName,
				Err:    err,
			}
		}
		fields[i] = field
	}

	return fields, nil
}

// parseField parses a single field.
func (r *Reader) parseField(index int, value string) (*Field, error) {
	// Treat as SimpleField if index is out of headers range
	if index >= len(r.headers) {
		return &Field{Value: value}, nil
	}

	header := r.headers[index]

	switch header.Kind {
	case SimpleField:
		return &Field{Value: value}, nil
	case ArrayField:
		return r.parseArrayField(header, value)
	case StructuredField:
		return r.parseStructuredField(header, value)
	case ArrayStructuredField:
		return r.parseArrayStructuredField(header, value)
	default:
		return &Field{Value: value}, nil
	}
}

// parseArrayField parses an array field per IETF CSV++ Section 2.2.2.
// Values are split by the array delimiter specified in the header.
// Example: "555-1234~555-5678" with delimiter '~' → Values: ["555-1234", "555-5678"]
func (r *Reader) parseArrayField(header *ColumnHeader, value string) (*Field, error) {
	if value == "" {
		return &Field{Values: []string{}}, nil
	}

	values := splitByRune(value, header.ArrayDelimiter)
	return &Field{Values: values}, nil
}

// parseStructuredField parses a structured field per IETF CSV++ Section 2.2.3.
// Components are split by the component delimiter and matched to header definitions.
// Example: "34.0522^-118.2437" (header: geo(lat^lon)) → Components: [{Value: "34.0522"}, {Value: "-118.2437"}]
func (r *Reader) parseStructuredField(header *ColumnHeader, value string) (*Field, error) {
	if value == "" {
		return &Field{Components: []*Field{}}, nil
	}

	return r.parseComponents(header.Components, header.ComponentDelimiter, value)
}

// parseArrayStructuredField parses an array structured field per IETF CSV++ Section 2.2.4.
// First splits by array delimiter, then each element is parsed as a structured field.
// Example: "home^123 Main~work^456 Oak" (header: address[](type^street))
// → Components: [[{Value: "home"}, {Value: "123 Main"}], [{Value: "work"}, {Value: "456 Oak"}]]
func (r *Reader) parseArrayStructuredField(header *ColumnHeader, value string) (*Field, error) {
	if value == "" {
		return &Field{Components: []*Field{}}, nil
	}

	// First split by array delimiter
	items := splitByRune(value, header.ArrayDelimiter)
	components := make([]*Field, 0, len(items))

	for _, item := range items {
		comp, err := r.parseComponents(header.Components, header.ComponentDelimiter, item)
		if err != nil {
			return nil, err
		}
		components = append(components, comp)
	}

	return &Field{Components: components}, nil
}

// parseComponents parses a component list (recursive).
func (r *Reader) parseComponents(headers []*ColumnHeader, delim rune, value string) (*Field, error) {
	parts := splitByRune(value, delim)
	components := make([]*Field, len(parts))

	for i, part := range parts {
		var comp *Field
		var err error

		// Parse according to header definition if available
		if i < len(headers) {
			compHeader := headers[i]
			switch compHeader.Kind {
			case SimpleField:
				comp = &Field{Value: part}
			case ArrayField:
				comp, err = r.parseArrayField(compHeader, part)
			case StructuredField:
				comp, err = r.parseStructuredField(compHeader, part)
			case ArrayStructuredField:
				comp, err = r.parseArrayStructuredField(compHeader, part)
			default:
				comp = &Field{Value: part}
			}
		} else {
			// Treat as SimpleField if no header definition
			comp = &Field{Value: part}
		}

		if err != nil {
			return nil, err
		}
		components[i] = comp
	}

	return &Field{Components: components}, nil
}

// splitByRune splits a string by the specified rune.
// Empty values are preserved (e.g., "a||b" → ["a", "", "b"]).
func splitByRune(s string, sep rune) []string {
	if s == "" {
		return []string{}
	}

	// Count separators to pre-allocate result slice
	n := 1
	for _, r := range s {
		if r == sep {
			n++
		}
	}

	result := make([]string, 0, n)
	start := 0

	for i, r := range s {
		if r == sep {
			result = append(result, s[start:i])
			start = i + len(string(sep))
		}
	}
	result = append(result, s[start:])

	return result
}
