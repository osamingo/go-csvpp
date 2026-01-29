package csvpputil

import (
	"bytes"
	"encoding/json/v2"
	"io"

	"github.com/osamingo/go-csvpp"
)

// JSONEncoderOption is a functional option for JSONEncoder.
type JSONEncoderOption func(*JSONEncoder)

// WithDeterministic sets whether JSON output should have deterministic key ordering.
// When true, map keys are sorted alphabetically for consistent output.
func WithDeterministic(v bool) JSONEncoderOption {
	return func(e *JSONEncoder) {
		e.deterministic = v
	}
}

// JSONEncoder writes CSV++ records as JSON to an output stream.
// It outputs a JSON array where each element is an object representing a record.
type JSONEncoder struct {
	w             io.Writer
	headers       []*csvpp.ColumnHeader
	deterministic bool
	started       bool
	closed        bool
}

// NewJSONEncoder creates a new JSONEncoder that writes to w.
func NewJSONEncoder(w io.Writer, headers []*csvpp.ColumnHeader, opts ...JSONEncoderOption) *JSONEncoder {
	e := &JSONEncoder{
		w:       w,
		headers: headers,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Encode writes a single record as a JSON object.
// The first call writes the opening bracket of the JSON array.
func (e *JSONEncoder) Encode(record []*csvpp.Field) error {
	if e.closed {
		return io.ErrClosedPipe
	}

	if !e.started {
		if _, err := e.w.Write([]byte{'['}); err != nil {
			return err
		}
		e.started = true
	} else {
		if _, err := e.w.Write([]byte{','}); err != nil {
			return err
		}
	}

	m := RecordToMap(e.headers, record)
	return json.MarshalWrite(e.w, m, json.Deterministic(e.deterministic))
}

// Close finishes the JSON array by writing the closing bracket.
// It must be called after all records have been encoded.
func (e *JSONEncoder) Close() error {
	if e.closed {
		return nil
	}
	e.closed = true

	if !e.started {
		_, err := e.w.Write([]byte("[]"))
		return err
	}

	_, err := e.w.Write([]byte{']'})
	return err
}

// MarshalJSON converts CSV++ records to JSON bytes.
// The output is a JSON array of objects, where each object represents a record.
func MarshalJSON(headers []*csvpp.ColumnHeader, records [][]*csvpp.Field, opts ...JSONEncoderOption) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewJSONEncoder(&buf, headers, opts...)

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
