package csvpputil_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"

	"github.com/osamingo/go-csvpp"
	"github.com/osamingo/go-csvpp/csvpputil"
)

func TestYAMLArrayWriter_Write(t *testing.T) {
	t.Parallel()

	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "tags", Kind: csvpp.ArrayField},
	}

	t.Run("success: single record", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		w := csvpputil.NewYAMLArrayWriter(&buf, headers)

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
		if err := yaml.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("yaml.Unmarshal() error = %v", err)
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
		w := csvpputil.NewYAMLArrayWriter(&buf, headers)

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
		if err := yaml.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("yaml.Unmarshal() error = %v", err)
		}

		want := []map[string]any{
			{"name": "Alice", "tags": []any{"go"}},
			{"name": "Bob", "tags": []any{"rust", "python"}},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("output mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("error: write after close", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		w := csvpputil.NewYAMLArrayWriter(&buf, headers)

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
		w := csvpputil.NewYAMLArrayWriter(&buf, headers)

		if err := w.Close(); err != nil {
			t.Fatalf("first Close() error = %v", err)
		}

		if err := w.Close(); err != nil {
			t.Fatalf("second Close() error = %v", err)
		}
	})
}

func TestYAMLArrayWriter_WriteWithCapacity(t *testing.T) {
	t.Parallel()

	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "tags", Kind: csvpp.ArrayField},
	}

	t.Run("success: writer with capacity hint", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		w := csvpputil.NewYAMLArrayWriter(&buf, headers, csvpputil.WithYAMLCapacity(2))

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
		if err := yaml.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("yaml.Unmarshal() error = %v", err)
		}

		want := []map[string]any{
			{"name": "Alice", "tags": []any{"go"}},
			{"name": "Bob", "tags": []any{"rust", "python"}},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("output mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: zero capacity is safe", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		w := csvpputil.NewYAMLArrayWriter(&buf, headers, csvpputil.WithYAMLCapacity(0))

		if err := w.Write([]*csvpp.Field{{Value: "Alice"}, {Values: []string{"go"}}}); err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		if err := w.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		var got []map[string]any
		if err := yaml.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("yaml.Unmarshal() error = %v", err)
		}

		want := []map[string]any{
			{"name": "Alice", "tags": []any{"go"}},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("output mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestMarshalYAML(t *testing.T) {
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

			got, err := csvpputil.MarshalYAML(tt.headers, tt.records)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var gotParsed []map[string]any
			if err := yaml.Unmarshal(got, &gotParsed); err != nil {
				t.Fatalf("yaml.Unmarshal() error = %v", err)
			}

			if diff := cmp.Diff(tt.want, gotParsed); diff != "" {
				t.Errorf("MarshalYAML() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWriteYAML(t *testing.T) {
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

		err := csvpputil.WriteYAML(&buf, headers, records)
		if err != nil {
			t.Fatalf("WriteYAML() error = %v", err)
		}

		var got []map[string]any
		if err := yaml.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("yaml.Unmarshal() error = %v", err)
		}

		want := []map[string]any{
			{"name": "Alice", "tags": []any{"go"}},
			{"name": "Bob", "tags": []any{"rust"}},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("WriteYAML() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: empty records", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		err := csvpputil.WriteYAML(&buf, headers, nil)
		if err != nil {
			t.Fatalf("WriteYAML() error = %v", err)
		}

		var got []map[string]any
		if err := yaml.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("yaml.Unmarshal() error = %v", err)
		}

		if diff := cmp.Diff([]map[string]any{}, got); diff != "" {
			t.Errorf("WriteYAML() mismatch (-want +got):\n%s", diff)
		}
	})
}
