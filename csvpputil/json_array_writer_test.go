package csvpputil_test

import (
	"bytes"
	"encoding/json/v2"
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/osamingo/go-csvpp"
	"github.com/osamingo/go-csvpp/csvpputil"
)

func TestJSONArrayWriter_Write(t *testing.T) {
	t.Parallel()

	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "tags", Kind: csvpp.ArrayField},
	}

	t.Run("success: single record", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		w := csvpputil.NewJSONArrayWriter(&buf, headers)

		err := w.Write([]*csvpp.Field{
			{Value: "Alice"},
			{Values: []string{"go", "rust"}},
		})
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		if err := w.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		var got []map[string]any
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		want := []map[string]any{
			{"name": "Alice", "tags": []any{"go", "rust"}},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("output mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: multiple records", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		w := csvpputil.NewJSONArrayWriter(&buf, headers)

		records := [][]*csvpp.Field{
			{{Value: "Alice"}, {Values: []string{"go"}}},
			{{Value: "Bob"}, {Values: []string{"rust", "python"}}},
		}

		for _, record := range records {
			if err := w.Write(record); err != nil {
				t.Fatalf("Write() error = %v", err)
			}
		}

		if err := w.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		var got []map[string]any
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		want := []map[string]any{
			{"name": "Alice", "tags": []any{"go"}},
			{"name": "Bob", "tags": []any{"rust", "python"}},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("output mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: empty records", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		w := csvpputil.NewJSONArrayWriter(&buf, headers)

		if err := w.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		if diff := cmp.Diff("[]\n", buf.String()); diff != "" {
			t.Errorf("output mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("error: write after close", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		w := csvpputil.NewJSONArrayWriter(&buf, headers)

		if err := w.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		err := w.Write([]*csvpp.Field{{Value: "Alice"}})
		if !errors.Is(err, io.ErrClosedPipe) {
			t.Errorf("Write() error = %v, want %v", err, io.ErrClosedPipe)
		}
	})

	t.Run("success: double close is safe", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		w := csvpputil.NewJSONArrayWriter(&buf, headers)

		if err := w.Close(); err != nil {
			t.Fatalf("first Close() error = %v", err)
		}

		if err := w.Close(); err != nil {
			t.Fatalf("second Close() error = %v", err)
		}
	})
}

func TestMarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		records [][]*csvpp.Field
		headers []*csvpp.ColumnHeader
		want    []map[string]any
		wantErr bool
	}{
		{
			name:    "success: nil records",
			records: nil,
			headers: []*csvpp.ColumnHeader{{Name: "name", Kind: csvpp.SimpleField}},
			want:    []map[string]any{},
		},
		{
			name:    "success: empty records",
			records: [][]*csvpp.Field{},
			headers: []*csvpp.ColumnHeader{{Name: "name", Kind: csvpp.SimpleField}},
			want:    []map[string]any{},
		},
		{
			name: "success: single record",
			records: [][]*csvpp.Field{
				{{Value: "Alice"}},
			},
			headers: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField},
			},
			want: []map[string]any{
				{"name": "Alice"},
			},
		},
		{
			name: "success: multiple records",
			records: [][]*csvpp.Field{
				{{Value: "Alice"}},
				{{Value: "Bob"}},
			},
			headers: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField},
			},
			want: []map[string]any{
				{"name": "Alice"},
				{"name": "Bob"},
			},
		},
		{
			name: "success: complex record",
			records: [][]*csvpp.Field{
				{
					{Value: "Alice"},
					{Values: []string{"go", "rust"}},
					{
						Components: []*csvpp.Field{
							{Value: "35.6"},
							{Value: "139.7"},
						},
					},
				},
			},
			headers: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField},
				{Name: "tags", Kind: csvpp.ArrayField},
				{
					Name: "geo",
					Kind: csvpp.StructuredField,
					Components: []*csvpp.ColumnHeader{
						{Name: "lat", Kind: csvpp.SimpleField},
						{Name: "lon", Kind: csvpp.SimpleField},
					},
				},
			},
			want: []map[string]any{
				{
					"name": "Alice",
					"tags": []any{"go", "rust"},
					"geo":  map[string]any{"lat": "35.6", "lon": "139.7"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := csvpputil.MarshalJSON(tt.headers, tt.records)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var gotParsed []map[string]any
			if err := json.Unmarshal(got, &gotParsed); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if diff := cmp.Diff(tt.want, gotParsed); diff != "" {
				t.Errorf("MarshalJSON() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "tags", Kind: csvpp.ArrayField},
	}

	t.Run("success: write records", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		records := [][]*csvpp.Field{
			{{Value: "Alice"}, {Values: []string{"go"}}},
			{{Value: "Bob"}, {Values: []string{"rust"}}},
		}

		err := csvpputil.WriteJSON(&buf, headers, records)
		if err != nil {
			t.Fatalf("WriteJSON() error = %v", err)
		}

		var got []map[string]any
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		want := []map[string]any{
			{"name": "Alice", "tags": []any{"go"}},
			{"name": "Bob", "tags": []any{"rust"}},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("WriteJSON() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: empty records", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		err := csvpputil.WriteJSON(&buf, headers, nil)
		if err != nil {
			t.Fatalf("WriteJSON() error = %v", err)
		}

		if diff := cmp.Diff("[]\n", buf.String()); diff != "" {
			t.Errorf("WriteJSON() mismatch (-want +got):\n%s", diff)
		}
	})
}
