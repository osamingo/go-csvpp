package csvpputil

import (
	"bytes"
	"io"

	"github.com/goccy/go-yaml"
	"github.com/osamingo/go-csvpp"
)

// YAMLArrayWriter writes CSV++ records as a YAML array.
// Due to YAML's structure (go-yaml doesn't support streaming array elements),
// records are buffered until Close.
type YAMLArrayWriter struct {
	w       io.Writer
	headers []*csvpp.ColumnHeader
	records []yaml.MapSlice
	closed  bool
}

// NewYAMLArrayWriter creates a new YAMLArrayWriter that writes to w.
func NewYAMLArrayWriter(w io.Writer, headers []*csvpp.ColumnHeader) *YAMLArrayWriter {
	return &YAMLArrayWriter{
		w:       w,
		headers: headers,
	}
}

// Write adds a single record to the buffer.
// The actual writing happens on Close.
func (w *YAMLArrayWriter) Write(record []*csvpp.Field) error {
	if w.closed {
		return io.ErrClosedPipe
	}

	m := fieldsToMapSlice(w.headers, record)
	w.records = append(w.records, m)
	return nil
}

// Close writes all buffered records as a YAML array and closes the writer.
// go-yaml requires the complete array for proper YAML array output.
func (w *YAMLArrayWriter) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true

	enc := yaml.NewEncoder(w.w)
	if err := enc.Encode(w.records); err != nil {
		return err
	}
	return enc.Close()
}

// MarshalYAML converts CSV++ records to YAML bytes.
// The output is a YAML array where each element is a record.
func MarshalYAML(headers []*csvpp.ColumnHeader, records [][]*csvpp.Field) ([]byte, error) {
	var buf bytes.Buffer
	w := NewYAMLArrayWriter(&buf, headers)

	for _, record := range records {
		if err := w.Write(record); err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// WriteYAML writes CSV++ records as a YAML array to the provided writer.
// The output is a YAML array where each element is a record.
func WriteYAML(w io.Writer, headers []*csvpp.ColumnHeader, records [][]*csvpp.Field) error {
	writer := NewYAMLArrayWriter(w, headers)

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return writer.Close()
}
