// Package csvpp implements the IETF CSV++ specification (draft-mscaldas-csvpp-01).
//
// CSV++ extends traditional CSV to support arrays and structured fields within cells,
// enabling complex data representation while maintaining CSV's simplicity.
// This package wraps encoding/csv and is fully compatible with RFC 4180.
//
// # Overview
//
// CSV++ introduces four field types beyond simple text values:
//
//   - Simple: "name" - plain text value
//   - Array: "tags[]" - multiple values separated by a delimiter (default: ~)
//   - Structured: "geo(lat^lon)" - named components separated by a delimiter (default: ^)
//   - ArrayStructured: "addresses[](street^city)" - array of structured values
//
// These field types are represented by the [FieldKind] constants:
// [SimpleField], [ArrayField], [StructuredField], and [ArrayStructuredField].
//
// # Basic Usage
//
// Reading CSV++ data:
//
//	r := csvpp.NewReader(file)
//
//	// Get parsed headers
//	headers, err := r.Headers()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Read records
//	for {
//	    record, err := r.Read()
//	    if err == io.EOF {
//	        break
//	    }
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    // process record
//	}
//
// Writing CSV++ data:
//
//	w := csvpp.NewWriter(file)
//	w.SetHeaders(headers)
//
//	if err := w.WriteHeader(); err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, record := range records {
//	    if err := w.Write(record); err != nil {
//	        log.Fatal(err)
//	    }
//	}
//	w.Flush()
//	if err := w.Error(); err != nil {
//	    log.Fatal(err)
//	}
//
// # Struct Mapping
//
// Use Marshal and Unmarshal for automatic struct mapping with struct tags:
//
//	type Person struct {
//	    Name   string   `csvpp:"name"`
//	    Phones []string `csvpp:"phone[]"`
//	    Geo    struct {
//	        Lat string
//	        Lon string
//	    } `csvpp:"geo(lat^lon)"`
//	}
//
//	// Read into structs
//	var people []Person
//	if err := csvpp.Unmarshal(file, &people); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Write from structs
//	var buf bytes.Buffer
//	if err := csvpp.Marshal(&buf, people); err != nil {
//	    log.Fatal(err)
//	}
//
// # Delimiter Conventions
//
// The IETF CSV++ specification recommends using specific delimiters for nested structures
// to avoid conflicts. The recommended progression is:
//
//   - Level 1 (arrays): ~ (tilde)
//   - Level 2 (components): ^ (caret)
//   - Level 3: ; (semicolon)
//   - Level 4: : (colon)
//
// This package uses ~ and ^ as defaults, matching the IETF recommendation.
//
// # Compatibility with encoding/csv
//
// This package wraps encoding/csv and inherits its RFC 4180 compliance.
// The Reader and Writer types expose the same configuration options:
//
//   - Comma: field delimiter (default: ',')
//   - Comment: comment character (Reader only)
//   - LazyQuotes: relaxed quote handling (Reader only)
//   - TrimLeadingSpace: trim leading whitespace (Reader only)
//   - UseCRLF: use \r\n line endings (Writer only)
//
// # Security Considerations
//
// The MaxNestingDepth option (default: 10) limits the depth of nested structures
// to prevent stack overflow attacks from maliciously crafted input.
//
// # Errors
//
// The package defines the following sentinel errors:
//
//   - [ErrNoHeader]: returned when attempting to read without a header row
//   - [ErrInvalidHeader]: returned when header format is invalid
//   - [ErrNestingTooDeep]: returned when nesting exceeds MaxNestingDepth
//
// Parse errors are wrapped in [ParseError], which provides line/column information.
//
// # Constants
//
// Default delimiters follow IETF recommendations:
//
//   - [DefaultArrayDelimiter]: ~ (tilde) for array fields
//   - [DefaultComponentDelimiter]: ^ (caret) for structured fields
//   - [DefaultMaxNestingDepth]: 10 (IETF recommended limit)
//
// # Specification Reference
//
// For the complete IETF CSV++ specification, see:
// https://datatracker.ietf.org/doc/draft-mscaldas-csvpp/
package csvpp
