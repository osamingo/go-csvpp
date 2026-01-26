package csvpp

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// parseColumnHeader parses a single column header string according to IETF CSV++ Section 2.2.
// ABNF (Section 2.2):
//
//	field = simple-field / array-field / struct-field / array-struct-field
func parseColumnHeader(s string) (*ColumnHeader, error) {
	return parseColumnHeaderWithDepth(s, 0, DefaultMaxNestingDepth)
}

// parseColumnHeaderWithDepth parses a column header with depth limit.
func parseColumnHeaderWithDepth(s string, depth, maxDepth int) (*ColumnHeader, error) {
	if depth > maxDepth {
		return nil, fmt.Errorf("%w: depth %d exceeds max %d", ErrNestingTooDeep, depth, maxDepth)
	}

	if s == "" {
		return nil, fmt.Errorf("%w: empty column header", ErrInvalidHeader)
	}

	h := &ColumnHeader{
		Kind: SimpleField,
	}

	remaining := s

	// 1. Extract name
	name, rest, err := parseName(remaining)
	if err != nil {
		return nil, err
	}
	h.Name = name
	remaining = rest

	if remaining == "" {
		return h, nil
	}

	// 2. Extract array delimiter if "[" is present
	// ABNF: array-field = name "[" [delimiter] "]"
	if strings.HasPrefix(remaining, "[") {
		delim, rest, err := parseArrayDelimiter(remaining)
		if err != nil {
			return nil, err
		}
		h.ArrayDelimiter = delim
		h.Kind = ArrayField
		remaining = rest
	}

	if remaining == "" {
		return h, nil
	}

	// 3. Extract component delimiter and component list if "(" is present
	// ABNF: struct-field = name [component-delim] "(" component-list ")"
	if len(remaining) > 0 {
		compDelim, compList, err := parseStructuredPartWithDepth(remaining, depth, maxDepth)
		if err != nil {
			return nil, err
		}
		h.ComponentDelimiter = compDelim
		h.Components = compList

		if h.Kind == ArrayField {
			h.Kind = ArrayStructuredField
		} else {
			h.Kind = StructuredField
		}
	}

	return h, nil
}

// parseName extracts the name part according to IETF CSV++ Section 2.2.
// ABNF (Section 2.2):
//
//	name = 1*field-char
//	field-char = ALPHA / DIGIT / "_" / "-"
//
// Note: Header names are restricted to ASCII characters per the IETF specification.
func parseName(s string) (name, rest string, err error) {
	var i int
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if !isFieldChar(r) {
			break
		}
		i += size
	}

	if i == 0 {
		return "", "", fmt.Errorf("%w: name is required", ErrInvalidHeader)
	}

	return s[:i], s[i:], nil
}

// isFieldChar checks if the rune is a valid field-char per IETF CSV++ Section 2.2.
// ABNF: field-char = ALPHA / DIGIT / "_" / "-"
// This restricts header names to ASCII alphanumeric characters, underscore, and hyphen.
func isFieldChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_' || r == '-'
}

// parseArrayDelimiter extracts the "[" delimiter "]" part per IETF CSV++ Section 2.2.2.
// ABNF: array-field = name "[" [delimiter] "]"
// If no delimiter is specified, DefaultArrayDelimiter (~) is used.
func parseArrayDelimiter(s string) (delim rune, rest string, err error) {
	if !strings.HasPrefix(s, "[") {
		return 0, s, nil
	}

	// Skip "["
	s = s[1:]

	// Find "]"
	idx := strings.Index(s, "]")
	if idx == -1 {
		return 0, "", fmt.Errorf("%w: missing closing bracket ']'", ErrInvalidHeader)
	}

	raw := s[:idx]
	rest = s[idx+1:]

	if raw == "" {
		// Use default delimiter
		delim = DefaultArrayDelimiter
	} else {
		r, size := utf8.DecodeRuneInString(raw)
		if size != len(raw) {
			return 0, "", fmt.Errorf("%w: array delimiter must be a single character", ErrInvalidHeader)
		}
		delim = r
	}

	return delim, rest, nil
}

// parseStructuredPartWithDepth parses the structured part with depth limit per IETF CSV++ Section 2.2.3.
// ABNF: struct-field = name [component-delim] "(" component-list ")"
// If no component delimiter is specified, DefaultComponentDelimiter (^) is used.
func parseStructuredPartWithDepth(s string, depth, maxDepth int) (compDelim rune, components []*ColumnHeader, err error) {
	if s == "" {
		return 0, nil, fmt.Errorf("%w: unexpected end of header", ErrInvalidHeader)
	}

	// Extract component delimiter (character before "(")
	parenIdx := strings.Index(s, "(")
	if parenIdx == -1 {
		return 0, nil, fmt.Errorf("%w: missing opening parenthesis '('", ErrInvalidHeader)
	}

	if parenIdx == 0 {
		// Use default component delimiter
		compDelim = DefaultComponentDelimiter
	} else {
		// Character before "(" is the component delimiter
		raw := s[:parenIdx]
		r, size := utf8.DecodeRuneInString(raw)
		if size != len(raw) {
			return 0, nil, fmt.Errorf("%w: component delimiter must be a single character", ErrInvalidHeader)
		}
		compDelim = r
	}

	// Parse after "("
	s = s[parenIdx+1:]

	// Find matching ")" (considering nesting)
	closeIdx := findClosingParen(s)
	if closeIdx == -1 {
		return 0, nil, fmt.Errorf("%w: missing closing parenthesis ')'", ErrInvalidHeader)
	}

	inner := s[:closeIdx]
	rest := s[closeIdx+1:]

	if rest != "" {
		return 0, nil, fmt.Errorf("%w: unexpected characters after closing parenthesis", ErrInvalidHeader)
	}

	// Parse component list (depth + 1)
	components, err = parseComponentListWithDepth(inner, compDelim, depth+1, maxDepth)
	if err != nil {
		return 0, nil, err
	}

	return compDelim, components, nil
}

// findClosingParen returns the index of the matching closing parenthesis.
func findClosingParen(s string) int {
	depth := 1
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch r {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
		i += size
	}
	return -1
}

// parseComponentListWithDepth parses a component list with depth limit per IETF CSV++ Section 2.2.3.
// ABNF: component-list = component *(component-delim component)
// Each component can itself be any field type, enabling nested structures.
func parseComponentListWithDepth(s string, delim rune, depth, maxDepth int) ([]*ColumnHeader, error) {
	if s == "" {
		return nil, fmt.Errorf("%w: component list is empty", ErrInvalidHeader)
	}

	parts := splitByDelimiter(s, delim)
	components := make([]*ColumnHeader, 0, len(parts))

	for _, part := range parts {
		comp, err := parseColumnHeaderWithDepth(part, depth, maxDepth)
		if err != nil {
			return nil, err
		}
		components = append(components, comp)
	}

	return components, nil
}

// splitByDelimiter splits a string by delimiter (ignoring content inside parentheses).
func splitByDelimiter(s string, delim rune) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == '(':
			depth++
			current.WriteRune(r)
		case r == ')':
			depth--
			current.WriteRune(r)
		case r == delim && depth == 0:
			parts = append(parts, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
		i += size
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// parseHeaderRecordWithMaxDepth parses an entire header row with depth limit.
func parseHeaderRecordWithMaxDepth(fields []string, maxDepth int) ([]*ColumnHeader, error) {
	if len(fields) == 0 {
		return nil, ErrNoHeader
	}

	headers := make([]*ColumnHeader, 0, len(fields))
	for _, field := range fields {
		h, err := parseColumnHeaderWithDepth(field, 0, maxDepth)
		if err != nil {
			return nil, err
		}
		headers = append(headers, h)
	}

	return headers, nil
}
