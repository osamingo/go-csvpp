package csvpputil

import (
	"bytes"
	"encoding/json/jsontext"
	"io"

	"github.com/osamingo/go-csvpp"
)

// JSONArrayWriterOption is a functional option for JSONArrayWriter.
type JSONArrayWriterOption func(*JSONArrayWriter)

// WithDeterministic is kept for API compatibility but has no effect.
// Output order is always determined by the headers order.
//
// Deprecated: This option is no longer needed as output order follows headers order.
func WithDeterministic(_ bool) JSONArrayWriterOption { //nolint:unused // kept for API compatibility
	return func(_ *JSONArrayWriter) {} //nolint:unused // kept for API compatibility
}

// JSONArrayWriter writes CSV++ records as a JSON array using streaming output.
// It uses jsontext.Encoder internally for efficient token-level writing.
type JSONArrayWriter struct {
	enc     *jsontext.Encoder
	headers []*csvpp.ColumnHeader
	started bool
	closed  bool
}

// NewJSONArrayWriter creates a new JSONArrayWriter that writes to w.
func NewJSONArrayWriter(w io.Writer, headers []*csvpp.ColumnHeader, opts ...JSONArrayWriterOption) *JSONArrayWriter {
	writer := &JSONArrayWriter{
		enc:     jsontext.NewEncoder(w),
		headers: headers,
	}
	for _, opt := range opts {
		opt(writer)
	}
	return writer
}

// Write writes a single record as a JSON object in the array.
// The first call writes '[', subsequent calls add ',' before each object.
func (w *JSONArrayWriter) Write(record []*csvpp.Field) error {
	if w.closed {
		return io.ErrClosedPipe
	}

	if !w.started {
		if err := w.enc.WriteToken(jsontext.BeginArray); err != nil {
			return err
		}
		w.started = true
	}

	return w.writeObject(w.headers, record)
}

// writeObject writes fields as a JSON object.
func (w *JSONArrayWriter) writeObject(headers []*csvpp.ColumnHeader, fields []*csvpp.Field) error {
	if err := w.enc.WriteToken(jsontext.BeginObject); err != nil {
		return err
	}

	n := min(len(headers), len(fields))
	for i := range n {
		header := headers[i]
		field := fields[i]

		// Write key
		if err := w.enc.WriteToken(jsontext.String(header.Name)); err != nil {
			return err
		}

		// Write value
		if err := w.writeValue(header, field); err != nil {
			return err
		}
	}

	return w.enc.WriteToken(jsontext.EndObject)
}

// writeValue writes a single field value.
func (w *JSONArrayWriter) writeValue(header *csvpp.ColumnHeader, field *csvpp.Field) error {
	if header == nil || field == nil {
		return w.enc.WriteToken(jsontext.Null)
	}

	switch header.Kind {
	case csvpp.SimpleField:
		return w.enc.WriteToken(jsontext.String(field.Value))

	case csvpp.ArrayField:
		if err := w.enc.WriteToken(jsontext.BeginArray); err != nil {
			return err
		}
		for _, v := range field.Values {
			if err := w.enc.WriteToken(jsontext.String(v)); err != nil {
				return err
			}
		}
		return w.enc.WriteToken(jsontext.EndArray)

	case csvpp.StructuredField:
		return w.writeObject(header.Components, field.Components)

	case csvpp.ArrayStructuredField:
		if err := w.enc.WriteToken(jsontext.BeginArray); err != nil {
			return err
		}
		for _, comp := range field.Components {
			if comp != nil {
				if err := w.writeObject(header.Components, comp.Components); err != nil {
					return err
				}
			}
		}
		return w.enc.WriteToken(jsontext.EndArray)

	default:
		return w.enc.WriteToken(jsontext.String(field.Value))
	}
}

// Close finishes the JSON array by writing ']'.
// jsontext.Encoder auto-flushes, so no explicit Flush needed.
func (w *JSONArrayWriter) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true

	if !w.started {
		if err := w.enc.WriteToken(jsontext.BeginArray); err != nil {
			return err
		}
	}

	return w.enc.WriteToken(jsontext.EndArray)
}

// MarshalJSON converts CSV++ records to JSON bytes.
// The output is a JSON array of objects, where each object represents a record.
func MarshalJSON(headers []*csvpp.ColumnHeader, records [][]*csvpp.Field, opts ...JSONArrayWriterOption) ([]byte, error) {
	var buf bytes.Buffer
	w := NewJSONArrayWriter(&buf, headers, opts...)

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

// WriteJSON writes CSV++ records as a JSON array to the provided writer.
// The output is a JSON array of objects, where each object represents a record.
func WriteJSON(w io.Writer, headers []*csvpp.ColumnHeader, records [][]*csvpp.Field, opts ...JSONArrayWriterOption) error {
	writer := NewJSONArrayWriter(w, headers, opts...)

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return writer.Close()
}
