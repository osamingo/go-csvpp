package csvpputil

import (
	"bytes"
	"io"

	"github.com/goccy/go-yaml"
	"github.com/osamingo/go-csvpp"
)

// YAMLArrayWriterOption is a functional option for YAMLArrayWriter.
type YAMLArrayWriterOption func(*YAMLArrayWriter)

// WithYAMLCapacity pre-allocates the internal buffer for the expected number of records.
// This reduces memory allocations when the approximate record count is known in advance.
func WithYAMLCapacity(n int) YAMLArrayWriterOption {
	return func(w *YAMLArrayWriter) {
		if n > 0 {
			w.records = make([]yaml.MapSlice, 0, n)
		}
	}
}

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
func NewYAMLArrayWriter(w io.Writer, headers []*csvpp.ColumnHeader, opts ...YAMLArrayWriterOption) *YAMLArrayWriter {
	writer := &YAMLArrayWriter{
		w:       w,
		headers: headers,
	}
	for _, opt := range opts {
		opt(writer)
	}
	return writer
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
	if err := encodeYAMLRecords(&buf, headers, records); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WriteYAML writes CSV++ records as a YAML array to the provided writer.
// The output is a YAML array where each element is a record.
func WriteYAML(w io.Writer, headers []*csvpp.ColumnHeader, records [][]*csvpp.Field) error {
	return encodeYAMLRecords(w, headers, records)
}

// encodeYAMLRecords builds the complete MapSlice array with exact allocation
// and encodes it in one shot. This avoids the overhead of the YAMLArrayWriter's
// per-record append growth.
func encodeYAMLRecords(w io.Writer, headers []*csvpp.ColumnHeader, records [][]*csvpp.Field) error {
	ms := make([]yaml.MapSlice, len(records))
	for i, record := range records {
		ms[i] = fieldsToMapSlice(headers, record)
	}
	enc := yaml.NewEncoder(w)
	if err := enc.Encode(ms); err != nil {
		return err
	}
	return enc.Close()
}
