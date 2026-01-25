package csvpp

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewWriter(t *testing.T) {
	t.Parallel()

	t.Run("success: creates writer with default comma", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		w := NewWriter(&buf)
		if w.Comma != ',' {
			t.Errorf("NewWriter().Comma = %q, want ','", w.Comma)
		}
	})
}

func TestWriter_WriteHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		headers []*ColumnHeader
		want    string
		wantErr bool
	}{
		{
			name: "success: simple headers",
			headers: []*ColumnHeader{
				{Name: "name", Kind: SimpleField},
				{Name: "age", Kind: SimpleField},
			},
			want: "name,age\n",
		},
		{
			name: "success: array header with default delimiter",
			headers: []*ColumnHeader{
				{Name: "name", Kind: SimpleField},
				{Name: "phone", Kind: ArrayField, ArrayDelimiter: DefaultArrayDelimiter},
			},
			want: "name,phone[]\n",
		},
		{
			name: "success: array header with custom delimiter",
			headers: []*ColumnHeader{
				{Name: "name", Kind: SimpleField},
				{Name: "phone", Kind: ArrayField, ArrayDelimiter: '|'},
			},
			want: "name,phone[|]\n",
		},
		{
			name: "success: structured header",
			headers: []*ColumnHeader{
				{Name: "name", Kind: SimpleField},
				{
					Name:               "geo",
					Kind:               StructuredField,
					ComponentDelimiter: DefaultComponentDelimiter,
					Components: []*ColumnHeader{
						{Name: "lat", Kind: SimpleField},
						{Name: "lon", Kind: SimpleField},
					},
				},
			},
			want: "name,geo(lat^lon)\n",
		},
		{
			name: "success: structured header with custom delimiter",
			headers: []*ColumnHeader{
				{Name: "name", Kind: SimpleField},
				{
					Name:               "geo",
					Kind:               StructuredField,
					ComponentDelimiter: ';',
					Components: []*ColumnHeader{
						{Name: "lat", Kind: SimpleField},
						{Name: "lon", Kind: SimpleField},
					},
				},
			},
			want: "name,geo;(lat;lon)\n",
		},
		{
			name: "success: array structured header",
			headers: []*ColumnHeader{
				{Name: "name", Kind: SimpleField},
				{
					Name:               "address",
					Kind:               ArrayStructuredField,
					ArrayDelimiter:     DefaultArrayDelimiter,
					ComponentDelimiter: DefaultComponentDelimiter,
					Components: []*ColumnHeader{
						{Name: "type", Kind: SimpleField},
						{Name: "street", Kind: SimpleField},
					},
				},
			},
			want: "name,address[](type^street)\n",
		},
		{
			name:    "error: no headers",
			headers: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			w := NewWriter(&buf)
			w.SetHeaders(tt.headers)

			err := w.WriteHeader()
			if (err != nil) != tt.wantErr {
				t.Errorf("Writer.WriteHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			w.Flush()
			got := buf.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Writer.WriteHeader() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWriter_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		headers []*ColumnHeader
		record  []*Field
		want    string
	}{
		{
			name: "success: simple fields",
			headers: []*ColumnHeader{
				{Name: "name", Kind: SimpleField},
				{Name: "age", Kind: SimpleField},
			},
			record: []*Field{
				{Value: "Alice"},
				{Value: "30"},
			},
			want: "Alice,30\n",
		},
		{
			name: "success: array field",
			headers: []*ColumnHeader{
				{Name: "name", Kind: SimpleField},
				{Name: "phone", Kind: ArrayField, ArrayDelimiter: DefaultArrayDelimiter},
			},
			record: []*Field{
				{Value: "Alice"},
				{Values: []string{"555-1234", "555-5678"}},
			},
			want: "Alice,555-1234~555-5678\n",
		},
		{
			name: "success: structured field",
			headers: []*ColumnHeader{
				{Name: "name", Kind: SimpleField},
				{
					Name:               "geo",
					Kind:               StructuredField,
					ComponentDelimiter: DefaultComponentDelimiter,
					Components: []*ColumnHeader{
						{Name: "lat", Kind: SimpleField},
						{Name: "lon", Kind: SimpleField},
					},
				},
			},
			record: []*Field{
				{Value: "Alice"},
				{Components: []*Field{
					{Value: "34.0522"},
					{Value: "-118.2437"},
				}},
			},
			want: "Alice,34.0522^-118.2437\n",
		},
		{
			name: "success: array structured field",
			headers: []*ColumnHeader{
				{Name: "name", Kind: SimpleField},
				{
					Name:               "address",
					Kind:               ArrayStructuredField,
					ArrayDelimiter:     DefaultArrayDelimiter,
					ComponentDelimiter: DefaultComponentDelimiter,
					Components: []*ColumnHeader{
						{Name: "type", Kind: SimpleField},
						{Name: "street", Kind: SimpleField},
					},
				},
			},
			record: []*Field{
				{Value: "Alice"},
				{Components: []*Field{
					{Components: []*Field{{Value: "home"}, {Value: "123 Main"}}},
					{Components: []*Field{{Value: "work"}, {Value: "456 Oak"}}},
				}},
			},
			want: "Alice,home^123 Main~work^456 Oak\n",
		},
		{
			name: "success: empty array field",
			headers: []*ColumnHeader{
				{Name: "name", Kind: SimpleField},
				{Name: "phone", Kind: ArrayField, ArrayDelimiter: DefaultArrayDelimiter},
			},
			record: []*Field{
				{Value: "Alice"},
				{Values: []string{}},
			},
			want: "Alice,\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			w := NewWriter(&buf)
			w.SetHeaders(tt.headers)

			err := w.Write(tt.record)
			if err != nil {
				t.Fatalf("Writer.Write() error = %v", err)
			}

			w.Flush()
			got := buf.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Writer.Write() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWriter_WriteAll(t *testing.T) {
	t.Parallel()

	headers := []*ColumnHeader{
		{Name: "name", Kind: SimpleField},
		{Name: "age", Kind: SimpleField},
	}
	records := [][]*Field{
		{{Value: "Alice"}, {Value: "30"}},
		{{Value: "Bob"}, {Value: "25"}},
	}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.SetHeaders(headers)

	err := w.WriteAll(records)
	if err != nil {
		t.Fatalf("Writer.WriteAll() error = %v", err)
	}

	want := "name,age\nAlice,30\nBob,25\n"
	got := buf.String()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Writer.WriteAll() mismatch (-want +got):\n%s", diff)
	}
}

func TestWriter_CustomComma(t *testing.T) {
	t.Parallel()

	headers := []*ColumnHeader{
		{Name: "name", Kind: SimpleField},
		{Name: "age", Kind: SimpleField},
	}
	record := []*Field{
		{Value: "Alice"},
		{Value: "30"},
	}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.Comma = ';'
	w.SetHeaders(headers)

	if err := w.WriteHeader(); err != nil {
		t.Fatalf("Writer.WriteHeader() error = %v", err)
	}
	if err := w.Write(record); err != nil {
		t.Fatalf("Writer.Write() error = %v", err)
	}
	w.Flush()

	want := "name;age\nAlice;30\n"
	got := buf.String()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Writer with custom comma mismatch (-want +got):\n%s", diff)
	}
}

func TestWriter_UseCRLF(t *testing.T) {
	t.Parallel()

	headers := []*ColumnHeader{
		{Name: "name", Kind: SimpleField},
	}
	record := []*Field{
		{Value: "Alice"},
	}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.UseCRLF = true
	w.SetHeaders(headers)

	if err := w.WriteHeader(); err != nil {
		t.Fatalf("Writer.WriteHeader() error = %v", err)
	}
	if err := w.Write(record); err != nil {
		t.Fatalf("Writer.Write() error = %v", err)
	}
	w.Flush()

	want := "name\r\nAlice\r\n"
	got := buf.String()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Writer with UseCRLF mismatch (-want +got):\n%s", diff)
	}
}

func TestFormatColumnHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header *ColumnHeader
		want   string
	}{
		{
			name:   "success: simple field",
			header: &ColumnHeader{Name: "name", Kind: SimpleField},
			want:   "name",
		},
		{
			name:   "success: array field with default delimiter",
			header: &ColumnHeader{Name: "phone", Kind: ArrayField, ArrayDelimiter: DefaultArrayDelimiter},
			want:   "phone[]",
		},
		{
			name:   "success: array field with custom delimiter",
			header: &ColumnHeader{Name: "phone", Kind: ArrayField, ArrayDelimiter: '|'},
			want:   "phone[|]",
		},
		{
			name: "success: structured field",
			header: &ColumnHeader{
				Name:               "geo",
				Kind:               StructuredField,
				ComponentDelimiter: DefaultComponentDelimiter,
				Components: []*ColumnHeader{
					{Name: "lat", Kind: SimpleField},
					{Name: "lon", Kind: SimpleField},
				},
			},
			want: "geo(lat^lon)",
		},
		{
			name: "success: array structured field",
			header: &ColumnHeader{
				Name:               "address",
				Kind:               ArrayStructuredField,
				ArrayDelimiter:     DefaultArrayDelimiter,
				ComponentDelimiter: DefaultComponentDelimiter,
				Components: []*ColumnHeader{
					{Name: "type", Kind: SimpleField},
					{Name: "street", Kind: SimpleField},
				},
			},
			want: "address[](type^street)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatColumnHeader(tt.header)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("formatColumnHeader() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWriter_EmptyStructuredField(t *testing.T) {
	t.Parallel()

	headers := []*ColumnHeader{
		{Name: "name", Kind: SimpleField},
		{
			Name:               "geo",
			Kind:               StructuredField,
			ComponentDelimiter: DefaultComponentDelimiter,
			Components: []*ColumnHeader{
				{Name: "lat", Kind: SimpleField},
				{Name: "lon", Kind: SimpleField},
			},
		},
	}
	record := []*Field{
		{Value: "Alice"},
		{Components: []*Field{}},
	}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.SetHeaders(headers)

	err := w.Write(record)
	if err != nil {
		t.Fatalf("Writer.Write() error = %v", err)
	}

	w.Flush()
	want := "Alice,\n"
	got := buf.String()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Writer.Write() mismatch (-want +got):\n%s", diff)
	}
}

func TestWriter_Flush(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	w := NewWriter(&buf)

	// Flush before writing anything should not panic
	w.Flush()

	if w.Error() != nil {
		t.Errorf("Writer.Error() = %v, want nil", w.Error())
	}
}

func TestWriter_WriteAllError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	w := NewWriter(&buf)
	// Don't set headers - should cause error

	records := [][]*Field{
		{{Value: "Alice"}, {Value: "30"}},
	}

	err := w.WriteAll(records)
	if err == nil {
		t.Error("Writer.WriteAll() expected error when no headers set")
	}
}

func TestWriter_NoHeaderForRecord(t *testing.T) {
	t.Parallel()

	headers := []*ColumnHeader{
		{Name: "name", Kind: SimpleField},
	}
	// Record has more fields than headers
	record := []*Field{
		{Value: "Alice"},
		{Value: "extra"},
	}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.SetHeaders(headers)

	err := w.Write(record)
	if err != nil {
		t.Fatalf("Writer.Write() error = %v", err)
	}

	w.Flush()
	// Extra field should be written as simple value
	want := "Alice,extra\n"
	got := buf.String()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Writer.Write() mismatch (-want +got):\n%s", diff)
	}
}

func TestFormatComponentList(t *testing.T) {
	t.Parallel()

	t.Run("success: empty component list", func(t *testing.T) {
		t.Parallel()

		got := formatComponentList([]*ColumnHeader{}, '^')
		if got != "" {
			t.Errorf("formatComponentList() = %q, want empty string", got)
		}
	})

	t.Run("success: multiple components", func(t *testing.T) {
		t.Parallel()

		components := []*ColumnHeader{
			{Name: "lat", Kind: SimpleField},
			{Name: "lon", Kind: SimpleField},
		}
		got := formatComponentList(components, '^')
		want := "lat^lon"
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("formatComponentList() mismatch (-want +got):\n%s", diff)
		}
	})
}
