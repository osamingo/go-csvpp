package csvpputil

import (
	"github.com/goccy/go-yaml"

	"github.com/osamingo/go-csvpp"
)

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

// fieldsToMapSlice converts fields to yaml.MapSlice preserving headers order.
func fieldsToMapSlice(headers []*csvpp.ColumnHeader, fields []*csvpp.Field) yaml.MapSlice {
	if len(headers) == 0 || len(fields) == 0 {
		return nil
	}

	n := min(len(headers), len(fields))
	result := make(yaml.MapSlice, n)
	for i := range n {
		result[i] = yaml.MapItem{
			Key:   headers[i].Name,
			Value: fieldToValueYAML(headers[i], fields[i]),
		}
	}
	return result
}

// fieldToValueYAML converts a single Field to its appropriate Go value for YAML.
// Uses yaml.MapSlice for structured fields to preserve key order.
func fieldToValueYAML(header *csvpp.ColumnHeader, field *csvpp.Field) any {
	if header == nil || field == nil {
		return nil
	}

	switch header.Kind {
	case csvpp.SimpleField:
		return field.Value
	case csvpp.ArrayField:
		return field.Values
	case csvpp.StructuredField:
		return fieldsToMapSlice(header.Components, field.Components)
	case csvpp.ArrayStructuredField:
		return arrayStructuredToSliceYAML(header.Components, field.Components)
	default:
		return field.Value
	}
}

// arrayStructuredToSliceYAML converts array-structured field to a slice of yaml.MapSlice.
func arrayStructuredToSliceYAML(headers []*csvpp.ColumnHeader, components []*csvpp.Field) []yaml.MapSlice {
	if len(components) == 0 {
		return nil
	}

	result := make([]yaml.MapSlice, 0, len(components))
	for _, comp := range components {
		if comp != nil {
			result = append(result, fieldsToMapSlice(headers, comp.Components))
		}
	}
	return result
}
