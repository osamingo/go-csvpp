package csvpp

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// Unmarshal decodes CSV++ data into a slice of structs.
// dst must be a pointer to a slice of structs.
func Unmarshal(r io.Reader, dst any) error {
	reader := NewReader(r)
	return UnmarshalReader(reader, dst)
}

// UnmarshalReader decodes from a Reader into a slice of structs.
func UnmarshalReader(r *Reader, dst any) error {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Pointer {
		return fmt.Errorf("csvpp: dst must be a pointer to slice")
	}

	sliceVal := dstVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("csvpp: dst must be a pointer to slice")
	}

	elemType := sliceVal.Type().Elem()
	if elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("csvpp: slice element must be a struct")
	}

	// Get headers
	headers, err := r.Headers()
	if err != nil {
		return err
	}

	// Create field mapping
	fieldMap := buildFieldMap(elemType, headers)

	// Read and decode all records
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Create new element
		elemVal := reflect.New(elemType).Elem()

		// Set field values
		if err := decodeRecord(record, elemVal, fieldMap); err != nil {
			return err
		}

		sliceVal.Set(reflect.Append(sliceVal, elemVal))
	}

	return nil
}

// Marshal encodes a slice of structs to CSV++ data.
func Marshal(w io.Writer, src any) error {
	writer := NewWriter(w)
	return MarshalWriter(writer, src)
}

// MarshalWriter encodes a slice of structs to a Writer.
func MarshalWriter(w *Writer, src any) error {
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() == reflect.Pointer {
		srcVal = srcVal.Elem()
	}
	if srcVal.Kind() != reflect.Slice {
		return fmt.Errorf("csvpp: src must be a slice")
	}

	if srcVal.Len() == 0 {
		return nil
	}

	elemType := srcVal.Type().Elem()
	if elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("csvpp: slice element must be a struct")
	}

	// Build headers
	headers := buildHeaders(elemType)
	w.SetHeaders(headers)

	// Write headers
	if err := w.WriteHeader(); err != nil {
		return err
	}

	// Encode each element
	for i := 0; i < srcVal.Len(); i++ {
		elemVal := srcVal.Index(i)
		if elemVal.Kind() == reflect.Pointer {
			elemVal = elemVal.Elem()
		}

		record := encodeRecord(elemVal, headers)
		if err := w.Write(record); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}

// fieldMapping holds the mapping information between fields and columns.
type fieldMapping struct {
	fieldIndex  int
	header      *ColumnHeader
	columnIndex int
}

// buildFieldMap creates a mapping between struct fields and headers.
func buildFieldMap(t reflect.Type, headers []*ColumnHeader) []fieldMapping {
	var mappings []fieldMapping

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("csvpp")
		if tag == "" || tag == "-" {
			continue
		}

		// Extract column name from tag (first part is the column name)
		tagName := extractTagName(tag)

		// Find corresponding column in headers
		for j, h := range headers {
			if h.Name == tagName {
				mappings = append(mappings, fieldMapping{
					fieldIndex:  i,
					header:      h,
					columnIndex: j,
				})
				break
			}
		}
	}

	return mappings
}

// extractTagName extracts the column name from a tag.
// Example: "phone[|]" → "phone", "address^(street^city)" → "address"
func extractTagName(tag string) string {
	// Extract up to "[" or "^" or "("
	for i, r := range tag {
		if r == '[' || r == '^' || r == '(' {
			return tag[:i]
		}
	}
	return tag
}

// buildHeaders builds headers from a struct.
func buildHeaders(t reflect.Type) []*ColumnHeader {
	var headers []*ColumnHeader

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("csvpp")
		if tag == "" || tag == "-" {
			continue
		}

		h, err := parseColumnHeader(tag)
		if err != nil {
			// Treat as simple field if error
			h = &ColumnHeader{
				Name: tag,
				Kind: SimpleField,
			}
		}
		headers = append(headers, h)
	}

	return headers
}

// decodeRecord decodes a record into a struct.
func decodeRecord(record []*Field, dst reflect.Value, mappings []fieldMapping) error {
	for _, m := range mappings {
		if m.columnIndex >= len(record) {
			continue
		}

		field := dst.Field(m.fieldIndex)
		if !field.CanSet() {
			continue
		}

		if err := decodeField(record[m.columnIndex], field, m.header); err != nil {
			return err
		}
	}

	return nil
}

// decodeField decodes a field value into a struct field.
func decodeField(f *Field, dst reflect.Value, header *ColumnHeader) error {
	switch header.Kind {
	case SimpleField:
		return decodeSimpleValue(f.Value, dst)
	case ArrayField:
		return decodeArrayValue(f.Values, dst)
	case StructuredField, ArrayStructuredField:
		return decodeStructuredValue(f.Components, dst, header)
	}
	return nil
}

// decodeSimpleValue decodes a simple value.
func decodeSimpleValue(value string, dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.String:
		dst.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value == "" {
			dst.SetInt(0)
		} else {
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			dst.SetInt(n)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value == "" {
			dst.SetUint(0)
		} else {
			n, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			dst.SetUint(n)
		}
	case reflect.Float32, reflect.Float64:
		if value == "" {
			dst.SetFloat(0)
		} else {
			n, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			dst.SetFloat(n)
		}
	case reflect.Bool:
		if value == "" {
			dst.SetBool(false)
		} else {
			b, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			dst.SetBool(b)
		}
	default:
		return fmt.Errorf("csvpp: unsupported type %s", dst.Type())
	}
	return nil
}

// decodeArrayValue decodes an array value.
func decodeArrayValue(values []string, dst reflect.Value) error {
	if dst.Kind() != reflect.Slice {
		return fmt.Errorf("csvpp: expected slice for array field, got %s", dst.Kind())
	}

	slice := reflect.MakeSlice(dst.Type(), len(values), len(values))

	for i, v := range values {
		elem := slice.Index(i)
		if err := decodeSimpleValue(v, elem); err != nil {
			return err
		}
	}

	dst.Set(slice)
	return nil
}

// decodeStructuredValue decodes a structured value.
func decodeStructuredValue(components []*Field, dst reflect.Value, header *ColumnHeader) error {
	// Array structured case
	if header.Kind == ArrayStructuredField {
		if dst.Kind() != reflect.Slice {
			return fmt.Errorf("csvpp: expected slice for array structured field")
		}

		elemType := dst.Type().Elem()
		slice := reflect.MakeSlice(dst.Type(), len(components), len(components))

		for i, comp := range components {
			elem := slice.Index(i)
			if elem.Kind() == reflect.Pointer {
				elem.Set(reflect.New(elemType.Elem()))
				elem = elem.Elem()
			}
			if err := decodeStructComponents(comp.Components, elem, header.Components); err != nil {
				return err
			}
		}

		dst.Set(slice)
		return nil
	}

	// Simple structured case
	return decodeStructComponents(components, dst, header.Components)
}

// decodeStructComponents decodes components into a struct.
func decodeStructComponents(components []*Field, dst reflect.Value, headers []*ColumnHeader) error {
	if dst.Kind() == reflect.Pointer {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		dst = dst.Elem()
	}

	if dst.Kind() != reflect.Struct {
		return fmt.Errorf("csvpp: expected struct for structured field, got %s", dst.Kind())
	}

	for i := 0; i < dst.NumField() && i < len(components) && i < len(headers); i++ {
		field := dst.Field(i)
		if !field.CanSet() {
			continue
		}
		if err := decodeField(components[i], field, headers[i]); err != nil {
			return err
		}
	}

	return nil
}

// encodeRecord encodes a struct to a record.
func encodeRecord(src reflect.Value, headers []*ColumnHeader) []*Field {
	fields := make([]*Field, 0, len(headers))

	fieldIdx := 0
	for i := 0; i < src.NumField(); i++ {
		structField := src.Type().Field(i)
		tag := structField.Tag.Get("csvpp")
		if tag == "" || tag == "-" {
			continue
		}

		if fieldIdx >= len(headers) {
			break
		}

		field := src.Field(i)
		f := encodeField(field, headers[fieldIdx])
		fields = append(fields, f)
		fieldIdx++
	}

	return fields
}

// encodeField encodes a struct field to a field value.
func encodeField(src reflect.Value, header *ColumnHeader) *Field {
	switch header.Kind {
	case SimpleField:
		return &Field{Value: encodeSimpleValue(src)}
	case ArrayField:
		return &Field{Values: encodeArrayValue(src)}
	case StructuredField:
		return &Field{Components: encodeStructuredValue(src, header)}
	case ArrayStructuredField:
		return &Field{Components: encodeArrayStructuredValue(src, header)}
	}
	return &Field{Value: encodeSimpleValue(src)}
}

// encodeSimpleValue encodes a simple value.
func encodeSimpleValue(src reflect.Value) string {
	switch src.Kind() {
	case reflect.String:
		return src.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(src.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(src.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(src.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(src.Bool())
	default:
		return fmt.Sprintf("%v", src.Interface())
	}
}

// encodeArrayValue encodes an array value.
func encodeArrayValue(src reflect.Value) []string {
	if src.Kind() != reflect.Slice {
		return []string{encodeSimpleValue(src)}
	}

	values := make([]string, src.Len())
	for i := 0; i < src.Len(); i++ {
		values[i] = encodeSimpleValue(src.Index(i))
	}
	return values
}

// encodeStructuredValue encodes a structured value.
func encodeStructuredValue(src reflect.Value, header *ColumnHeader) []*Field {
	if src.Kind() == reflect.Pointer {
		if src.IsNil() {
			return []*Field{}
		}
		src = src.Elem()
	}

	if src.Kind() != reflect.Struct {
		return []*Field{{Value: encodeSimpleValue(src)}}
	}

	components := make([]*Field, 0, len(header.Components))
	for i := 0; i < src.NumField() && i < len(header.Components); i++ {
		field := src.Field(i)
		comp := encodeField(field, header.Components[i])
		components = append(components, comp)
	}
	return components
}

// encodeArrayStructuredValue encodes an array structured value.
func encodeArrayStructuredValue(src reflect.Value, header *ColumnHeader) []*Field {
	if src.Kind() != reflect.Slice {
		return encodeStructuredValue(src, header)
	}

	components := make([]*Field, src.Len())
	for i := 0; i < src.Len(); i++ {
		elem := src.Index(i)
		comps := encodeStructuredValue(elem, header)
		components[i] = &Field{Components: comps}
	}
	return components
}
