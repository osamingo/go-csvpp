package csvpputil

import (
	"github.com/osamingo/go-csvpp"
)

// RecordToMap converts a single CSV++ record to a map[string]any.
// The resulting map uses field names as keys and converts:
//   - SimpleField: string value
//   - ArrayField: []string
//   - StructuredField: map[string]any with component names as keys
//   - ArrayStructuredField: []map[string]any
func RecordToMap(headers []*csvpp.ColumnHeader, record []*csvpp.Field) map[string]any {
	return fieldsToMap(headers, record)
}

// fieldsToMap converts fields to a map using headers as keys.
func fieldsToMap(headers []*csvpp.ColumnHeader, fields []*csvpp.Field) map[string]any {
	if len(headers) == 0 || len(fields) == 0 {
		return nil
	}

	n := min(len(headers), len(fields))
	result := make(map[string]any, n)
	for i := range n {
		result[headers[i].Name] = fieldToValue(headers[i], fields[i])
	}
	return result
}

// fieldToValue converts a single Field to its appropriate Go value.
func fieldToValue(header *csvpp.ColumnHeader, field *csvpp.Field) any {
	if header == nil || field == nil {
		return nil
	}

	switch header.Kind {
	case csvpp.SimpleField:
		return field.Value
	case csvpp.ArrayField:
		return field.Values
	case csvpp.StructuredField:
		return fieldsToMap(header.Components, field.Components)
	case csvpp.ArrayStructuredField:
		return arrayStructuredToSlice(header.Components, field.Components)
	default:
		return field.Value
	}
}

// arrayStructuredToSlice converts array-structured field to a slice of maps.
func arrayStructuredToSlice(headers []*csvpp.ColumnHeader, components []*csvpp.Field) []map[string]any {
	if len(components) == 0 {
		return nil
	}

	result := make([]map[string]any, 0, len(components))
	for _, comp := range components {
		if comp != nil {
			result = append(result, fieldsToMap(headers, comp.Components))
		}
	}
	return result
}
