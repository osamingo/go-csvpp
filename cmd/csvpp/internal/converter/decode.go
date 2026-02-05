// Package converter provides conversion utilities for the csvpp CLI.
package converter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/goccy/go-yaml"

	"github.com/osamingo/go-csvpp"
)

// FromJSON reads JSON array and converts to CSVPP headers and records.
// The JSON must be an array of objects with consistent keys.
func FromJSON(r io.Reader) ([]*csvpp.ColumnHeader, [][]*csvpp.Field, error) {
	var data []map[string]any
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return nil, nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	if len(data) == 0 {
		return nil, nil, nil
	}

	headers := inferHeaders(data)
	records := convertRecords(headers, data)

	return headers, records, nil
}

// FromYAML reads YAML array and converts to CSVPP headers and records.
// The YAML must be an array of objects with consistent keys.
func FromYAML(r io.Reader) ([]*csvpp.ColumnHeader, [][]*csvpp.Field, error) {
	var data []map[string]any
	if err := yaml.NewDecoder(r).Decode(&data); err != nil {
		return nil, nil, fmt.Errorf("failed to decode YAML: %w", err)
	}

	if len(data) == 0 {
		return nil, nil, nil
	}

	headers := inferHeaders(data)
	records := convertRecords(headers, data)

	return headers, records, nil
}

// inferHeaders infers CSVPP headers from JSON/YAML data structure.
// Header inference rules:
//   - string → SimpleField
//   - []string → ArrayField
//   - map[string]any → StructuredField
//   - []map[string]any → ArrayStructuredField
func inferHeaders(data []map[string]any) []*csvpp.ColumnHeader {
	if len(data) == 0 {
		return nil
	}

	// Collect all unique keys from all records (first record defines order)
	keyOrder := collectKeyOrder(data[0])
	keyTypes := make(map[string]csvpp.FieldKind)
	keyComponents := make(map[string][]*csvpp.ColumnHeader)

	// Analyze all records to determine consistent types
	for _, record := range data {
		for key, value := range record {
			kind, components := inferFieldKind(value)
			if existing, ok := keyTypes[key]; ok {
				// Keep the more complex type if there's a mismatch
				if kind > existing {
					keyTypes[key] = kind
					keyComponents[key] = components
				}
			} else {
				keyTypes[key] = kind
				keyComponents[key] = components
			}
		}
	}

	// Build headers maintaining key order
	headers := make([]*csvpp.ColumnHeader, 0, len(keyOrder))
	for _, key := range keyOrder {
		header := &csvpp.ColumnHeader{
			Name:               key,
			Kind:               keyTypes[key],
			ArrayDelimiter:     csvpp.DefaultArrayDelimiter,
			ComponentDelimiter: csvpp.DefaultComponentDelimiter,
			Components:         keyComponents[key],
		}
		headers = append(headers, header)
	}

	return headers
}

// collectKeyOrder returns keys in iteration order (Go 1.12+ maps have random order).
// Uses first record to determine key order.
func collectKeyOrder(record map[string]any) []string {
	keys := make([]string, 0, len(record))
	for key := range record {
		keys = append(keys, key)
	}
	return keys
}

// inferFieldKind determines the FieldKind from a value.
func inferFieldKind(value any) (csvpp.FieldKind, []*csvpp.ColumnHeader) {
	switch v := value.(type) {
	case []any:
		if len(v) == 0 {
			return csvpp.ArrayField, nil
		}
		// Check first element to determine if it's array of strings or array of objects
		switch elem := v[0].(type) {
		case map[string]any:
			// ArrayStructuredField
			components := inferComponentHeaders(elem)
			return csvpp.ArrayStructuredField, components
		default:
			// ArrayField
			return csvpp.ArrayField, nil
		}
	case map[string]any:
		// StructuredField
		components := inferComponentHeaders(v)
		return csvpp.StructuredField, components
	default:
		// SimpleField
		return csvpp.SimpleField, nil
	}
}

// inferComponentHeaders creates headers for structured field components.
func inferComponentHeaders(m map[string]any) []*csvpp.ColumnHeader {
	headers := make([]*csvpp.ColumnHeader, 0, len(m))
	for key, value := range m {
		kind, components := inferFieldKind(value)
		header := &csvpp.ColumnHeader{
			Name:               key,
			Kind:               kind,
			ArrayDelimiter:     csvpp.DefaultArrayDelimiter,
			ComponentDelimiter: csvpp.DefaultComponentDelimiter,
			Components:         components,
		}
		headers = append(headers, header)
	}
	return headers
}

// convertRecords converts data records to CSVPP fields.
func convertRecords(headers []*csvpp.ColumnHeader, data []map[string]any) [][]*csvpp.Field {
	records := make([][]*csvpp.Field, 0, len(data))

	for _, record := range data {
		fields := make([]*csvpp.Field, len(headers))
		for i, header := range headers {
			value := record[header.Name]
			fields[i] = convertValue(header, value)
		}
		records = append(records, fields)
	}

	return records
}

// convertValue converts a single value to a CSVPP Field.
func convertValue(header *csvpp.ColumnHeader, value any) *csvpp.Field {
	if value == nil {
		return &csvpp.Field{}
	}

	switch header.Kind {
	case csvpp.SimpleField:
		return &csvpp.Field{Value: toString(value)}

	case csvpp.ArrayField:
		arr, ok := value.([]any)
		if !ok {
			return &csvpp.Field{Values: []string{toString(value)}}
		}
		values := make([]string, len(arr))
		for i, v := range arr {
			values[i] = toString(v)
		}
		return &csvpp.Field{Values: values}

	case csvpp.StructuredField:
		m, ok := value.(map[string]any)
		if !ok {
			return &csvpp.Field{}
		}
		components := make([]*csvpp.Field, len(header.Components))
		for i, compHeader := range header.Components {
			compValue := m[compHeader.Name]
			components[i] = convertValue(compHeader, compValue)
		}
		return &csvpp.Field{Components: components}

	case csvpp.ArrayStructuredField:
		arr, ok := value.([]any)
		if !ok {
			return &csvpp.Field{Components: []*csvpp.Field{}}
		}
		components := make([]*csvpp.Field, len(arr))
		for i, elem := range arr {
			m, ok := elem.(map[string]any)
			if !ok {
				components[i] = &csvpp.Field{}
				continue
			}
			compFields := make([]*csvpp.Field, len(header.Components))
			for j, compHeader := range header.Components {
				compValue := m[compHeader.Name]
				compFields[j] = convertValue(compHeader, compValue)
			}
			components[i] = &csvpp.Field{Components: compFields}
		}
		return &csvpp.Field{Components: components}

	default:
		return &csvpp.Field{Value: toString(value)}
	}
}

// toString converts any value to string.
func toString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		// JSON numbers are decoded as float64
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%v", val)
	case int, int64, int32:
		return fmt.Sprintf("%d", val)
	case bool:
		return fmt.Sprintf("%t", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
