package csvpputil

import (
	"bytes"
	"io"

	"github.com/goccy/go-yaml"
	"github.com/osamingo/go-csvpp"
)

// YAMLEncoder writes CSV++ records as YAML to an output stream.
// Records are collected and written as a single YAML array on Close.
type YAMLEncoder struct {
	w       io.Writer
	headers []*csvpp.ColumnHeader
	records []map[string]any
	closed  bool
}

// NewYAMLEncoder creates a new YAMLEncoder that writes to w.
func NewYAMLEncoder(w io.Writer, headers []*csvpp.ColumnHeader) *YAMLEncoder {
	return &YAMLEncoder{
		w:       w,
		headers: headers,
	}
}

// Encode adds a record to the encoder.
// The record will be written when Close is called.
func (e *YAMLEncoder) Encode(record []*csvpp.Field) error {
	if e.closed {
		return io.ErrClosedPipe
	}

	m := RecordToMap(e.headers, record)
	e.records = append(e.records, m)
	return nil
}

// Close writes all records as a YAML array and finishes the stream.
func (e *YAMLEncoder) Close() error {
	if e.closed {
		return nil
	}
	e.closed = true

	enc := yaml.NewEncoder(e.w)
	if err := enc.Encode(e.records); err != nil {
		return err
	}
	return enc.Close()
}

// MarshalYAML converts CSV++ records to YAML bytes.
// The output is a YAML array where each element is a record.
func MarshalYAML(headers []*csvpp.ColumnHeader, records [][]*csvpp.Field) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewYAMLEncoder(&buf, headers)

	for _, record := range records {
		if err := enc.Encode(record); err != nil {
			return nil, err
		}
	}

	if err := enc.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
