package csvpp

import (
	"errors"
	"fmt"
)

// FieldKind represents the type of field as defined in IETF CSV++ Section 2.2.
// See: https://datatracker.ietf.org/doc/draft-mscaldas-csvpp/
type FieldKind int

const (
	SimpleField          FieldKind = iota // IETF Section 2.2.1: simple-field = name
	ArrayField                            // IETF Section 2.2.2: array-field = name "[" [delimiter] "]"
	StructuredField                       // IETF Section 2.2.3: struct-field = name [component-delim] "(" component-list ")"
	ArrayStructuredField                  // IETF Section 2.2.4: array-struct-field = name "[" [delimiter] "]" [component-delim] "(" component-list ")"
)

// String returns the string representation of FieldKind.
func (k FieldKind) String() string {
	switch k {
	case SimpleField:
		return "SimpleField"
	case ArrayField:
		return "ArrayField"
	case StructuredField:
		return "StructuredField"
	case ArrayStructuredField:
		return "ArrayStructuredField"
	default:
		return fmt.Sprintf("FieldKind(%d)", k)
	}
}

// Default delimiters as recommended in IETF CSV++ Section 2.3.2.
// The specification suggests delimiter progression: ~ → ^ → ; → : for nested structures.
const (
	DefaultArrayDelimiter     = '~' // IETF Section 2.3.2: recommended for array fields
	DefaultComponentDelimiter = '^' // IETF Section 2.3.2: recommended for structured fields
)

// DefaultMaxNestingDepth is the default maximum nesting depth.
// IETF Section 5 (Security Considerations) recommends limiting nesting depth to prevent
// stack overflow attacks from maliciously crafted input.
const DefaultMaxNestingDepth = 10

// ColumnHeader represents the declaration information for an individual field.
// It corresponds to the ABNF "field" rule in IETF CSV++ Section 2.2:
//
//	field = simple-field / array-field / struct-field / array-struct-field
//	name  = 1*field-char
//	field-char = ALPHA / DIGIT / "_" / "-"
type ColumnHeader struct {
	Name               string          // Field name (ABNF: name = 1*field-char)
	Kind               FieldKind       // Field type (IETF Section 2.2)
	ArrayDelimiter     rune            // Array delimiter (ABNF: delimiter)
	ComponentDelimiter rune            // Component delimiter (ABNF: component-delim)
	Components         []*ColumnHeader // Component list (ABNF: component-list)
}

// Field represents a parsed field value from a data row.
// The populated fields depend on the corresponding ColumnHeader.Kind:
//
//   - SimpleField: Value is set
//   - ArrayField: Values is set
//   - StructuredField: Components is set (each component is a Field)
//   - ArrayStructuredField: Components is set (each is a Field with its own Components)
type Field struct {
	Value      string   // Value for SimpleField
	Values     []string // Values for ArrayField (IETF Section 2.2.2)
	Components []*Field // Components for StructuredField/ArrayStructuredField (IETF Section 2.2.3/2.2.4)
}

// Error definitions.
var (
	ErrNoHeader       = errors.New("csvpp: header record is required")
	ErrInvalidHeader  = errors.New("csvpp: invalid column header format")
	ErrNestingTooDeep = errors.New("csvpp: nesting level exceeds limit")
)

// ParseError holds detailed information about an error that occurred during parsing.
type ParseError struct {
	Line   int    // Line number where the error occurred (1-based)
	Column int    // Column number where the error occurred (1-based)
	Field  string // Field name (if available)
	Err    error  // Original error
}

// Error returns the error message for ParseError.
func (e *ParseError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("csvpp: line %d, column %d (field %q): %v", e.Line, e.Column, e.Field, e.Err)
	}
	if e.Column > 0 {
		return fmt.Sprintf("csvpp: line %d, column %d: %v", e.Line, e.Column, e.Err)
	}
	return fmt.Sprintf("csvpp: line %d: %v", e.Line, e.Err)
}

// Unwrap returns the original error.
func (e *ParseError) Unwrap() error {
	return e.Err
}
