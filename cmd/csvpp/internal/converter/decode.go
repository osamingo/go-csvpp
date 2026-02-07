// Package converter provides conversion utilities for the csvpp CLI.
package converter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/goccy/go-yaml"

	"github.com/osamingo/go-csvpp"
)

// keyOrderInfo holds the ordered keys extracted from the first record.
type keyOrderInfo struct {
	keys   []string
	nested map[string]*keyOrderInfo
}

// FromJSON reads JSON array and converts to CSVPP headers and records.
// The JSON must be an array of objects with consistent keys.
func FromJSON(r io.Reader) ([]*csvpp.ColumnHeader, [][]*csvpp.Field, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read input: %w", err)
	}

	var records []map[string]any
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	if len(records) == 0 {
		return nil, nil, nil
	}

	order, err := extractJSONKeyOrder(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract JSON key order: %w", err)
	}

	headers := inferHeaders(records, order)
	fields := convertRecords(headers, records)

	return headers, fields, nil
}

// FromYAML reads YAML array and converts to CSVPP headers and records.
// The YAML must be an array of objects with consistent keys.
func FromYAML(r io.Reader) ([]*csvpp.ColumnHeader, [][]*csvpp.Field, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read input: %w", err)
	}

	var records []map[string]any
	if err := yaml.Unmarshal(data, &records); err != nil {
		return nil, nil, fmt.Errorf("failed to decode YAML: %w", err)
	}

	if len(records) == 0 {
		return nil, nil, nil
	}

	order, err := extractYAMLKeyOrder(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract YAML key order: %w", err)
	}

	headers := inferHeaders(records, order)
	fields := convertRecords(headers, records)

	return headers, fields, nil
}

// extractJSONKeyOrder extracts key order from the first record in a JSON array.
func extractJSONKeyOrder(data []byte) (*keyOrderInfo, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return &keyOrderInfo{nested: make(map[string]*keyOrderInfo)}, nil
	}
	return readJSONObjectOrder(json.NewDecoder(bytes.NewReader(raw[0])))
}

// readJSONObjectOrder reads one JSON object from a decoder and extracts ordered keys.
func readJSONObjectOrder(dec *json.Decoder) (*keyOrderInfo, error) {
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := t.(json.Delim); !ok || d != '{' {
		return nil, fmt.Errorf("expected '{', got %v", t)
	}

	info := &keyOrderInfo{nested: make(map[string]*keyOrderInfo)}

	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := t.(string)
		if !ok {
			return nil, fmt.Errorf("expected string key, got %T", t)
		}
		info.keys = append(info.keys, key)

		nested, err := readJSONValueOrder(dec)
		if err != nil {
			return nil, err
		}
		if nested != nil {
			info.nested[key] = nested
		}
	}

	// consume closing '}'
	_, err = dec.Token()
	return info, err
}

// readJSONValueOrder reads one JSON value, extracting key order if it's an object or array of objects.
func readJSONValueOrder(dec *json.Decoder) (*keyOrderInfo, error) {
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}

	d, ok := t.(json.Delim)
	if !ok {
		return nil, nil // scalar value
	}

	switch d {
	case '{':
		// Object value - extract component keys
		info := &keyOrderInfo{nested: make(map[string]*keyOrderInfo)}
		for dec.More() {
			t, err = dec.Token()
			if err != nil {
				return nil, err
			}
			key, ok := t.(string)
			if !ok {
				return nil, fmt.Errorf("expected string key, got %T", t)
			}
			info.keys = append(info.keys, key)
			if err := skipJSONValue(dec); err != nil {
				return nil, err
			}
		}
		_, err = dec.Token() // '}'
		return info, err

	case '[':
		// Array value - check if first element is an object
		if !dec.More() {
			_, err = dec.Token() // ']'
			return nil, err
		}

		t, err = dec.Token()
		if err != nil {
			return nil, err
		}

		if d2, ok := t.(json.Delim); ok && d2 == '{' {
			// Array of objects - extract keys from the first object
			info := &keyOrderInfo{nested: make(map[string]*keyOrderInfo)}
			for dec.More() {
				t, err = dec.Token()
				if err != nil {
					return nil, err
				}
				key, ok := t.(string)
				if !ok {
					return nil, fmt.Errorf("expected string key, got %T", t)
				}
				info.keys = append(info.keys, key)
				if err := skipJSONValue(dec); err != nil {
					return nil, err
				}
			}
			_, err = dec.Token() // closing '}' of first object
			if err != nil {
				return nil, err
			}
			// Skip remaining array elements
			for dec.More() {
				if err := skipJSONValue(dec); err != nil {
					return nil, err
				}
			}
			_, err = dec.Token() // ']'
			return info, err
		}

		// First element is not an object - skip the rest
		if d2, ok := t.(json.Delim); ok {
			if err := skipJSONDelimContent(dec, d2); err != nil {
				return nil, err
			}
		}
		for dec.More() {
			if err := skipJSONValue(dec); err != nil {
				return nil, err
			}
		}
		_, err = dec.Token() // ']'
		return nil, err
	}

	return nil, nil
}

// skipJSONValue reads and discards one complete JSON value.
func skipJSONValue(dec *json.Decoder) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}
	d, ok := t.(json.Delim)
	if !ok {
		return nil // scalar
	}
	return skipJSONDelimContent(dec, d)
}

// skipJSONDelimContent skips content after an opening delimiter ('{' or '[') was consumed.
func skipJSONDelimContent(dec *json.Decoder, open json.Delim) error {
	for dec.More() {
		if open == '{' {
			if _, err := dec.Token(); err != nil { // skip key
				return err
			}
		}
		if err := skipJSONValue(dec); err != nil {
			return err
		}
	}
	_, err := dec.Token() // closing delimiter
	return err
}

// extractYAMLKeyOrder extracts key order from the first record in a YAML sequence.
func extractYAMLKeyOrder(data []byte) (*keyOrderInfo, error) {
	var records []yaml.MapSlice
	if err := yaml.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return &keyOrderInfo{nested: make(map[string]*keyOrderInfo)}, nil
	}
	return buildYAMLKeyOrder(records[0]), nil
}

// buildYAMLKeyOrder builds keyOrderInfo from a yaml.MapSlice.
func buildYAMLKeyOrder(ms yaml.MapSlice) *keyOrderInfo {
	info := &keyOrderInfo{nested: make(map[string]*keyOrderInfo)}
	for _, item := range ms {
		key := fmt.Sprintf("%v", item.Key)
		info.keys = append(info.keys, key)

		switch v := item.Value.(type) {
		case yaml.MapSlice:
			info.nested[key] = buildYAMLKeyOrder(v)
		case []any:
			if len(v) > 0 {
				if ms2, ok := v[0].(yaml.MapSlice); ok {
					info.nested[key] = buildYAMLKeyOrder(ms2)
				}
			}
		}
	}
	return info
}

// inferHeaders infers CSVPP headers from data using the provided key order.
//
// Header inference rules:
//   - string → SimpleField
//   - []string → ArrayField
//   - map[string]any → StructuredField
//   - []map[string]any → ArrayStructuredField
func inferHeaders(data []map[string]any, order *keyOrderInfo) []*csvpp.ColumnHeader {
	if len(data) == 0 {
		return nil
	}

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

	// Build headers using preserved key order
	headers := make([]*csvpp.ColumnHeader, 0, len(order.keys))
	for _, key := range order.keys {
		components := keyComponents[key]
		if nestedOrder, ok := order.nested[key]; ok && len(components) > 0 {
			components = reorderComponents(components, nestedOrder)
		}

		header := &csvpp.ColumnHeader{
			Name:               key,
			Kind:               keyTypes[key],
			ArrayDelimiter:     csvpp.DefaultArrayDelimiter,
			ComponentDelimiter: csvpp.DefaultComponentDelimiter,
			Components:         components,
		}
		headers = append(headers, header)
	}

	return headers
}

// reorderComponents reorders component headers according to the key order info.
func reorderComponents(components []*csvpp.ColumnHeader, order *keyOrderInfo) []*csvpp.ColumnHeader {
	compMap := make(map[string]*csvpp.ColumnHeader, len(components))
	for _, c := range components {
		compMap[c.Name] = c
	}

	ordered := make([]*csvpp.ColumnHeader, 0, len(order.keys))
	for _, key := range order.keys {
		if c, ok := compMap[key]; ok {
			ordered = append(ordered, c)
		}
	}
	return ordered
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
// Note: The returned order may be non-deterministic (map iteration).
// Callers should use reorderComponents to fix the order.
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
