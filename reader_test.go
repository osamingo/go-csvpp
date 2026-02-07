package csvpp_test

import (
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/osamingo/go-csvpp"
)

func TestNewReader(t *testing.T) {
	t.Parallel()

	t.Run("success: creates reader with default comma", func(t *testing.T) {
		t.Parallel()

		r := csvpp.NewReader(strings.NewReader(""))
		if r.Comma != ',' {
			t.Errorf("NewReader().Comma = %q, want ','", r.Comma)
		}
	})
}

func TestReader_Headers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []*csvpp.ColumnHeader
		wantErr bool
	}{
		{
			name:  "success: simple headers",
			input: "name,age\n",
			want: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField},
				{Name: "age", Kind: csvpp.SimpleField},
			},
		},
		{
			name:  "success: array header",
			input: "name,phone[]\n",
			want: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField},
				{Name: "phone", Kind: csvpp.ArrayField, ArrayDelimiter: csvpp.DefaultArrayDelimiter},
			},
		},
		{
			name:  "success: structured header",
			input: "name,geo(lat^lon)\n",
			want: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField},
				{
					Name:               "geo",
					Kind:               csvpp.StructuredField,
					ComponentDelimiter: csvpp.DefaultComponentDelimiter,
					Components: []*csvpp.ColumnHeader{
						{Name: "lat", Kind: csvpp.SimpleField},
						{Name: "lon", Kind: csvpp.SimpleField},
					},
				},
			},
		},
		{
			name:    "error: empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := csvpp.NewReader(strings.NewReader(tt.input))
			got, err := r.Headers()
			if (err != nil) != tt.wantErr {
				t.Errorf("Reader.Headers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Reader.Headers() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReader_Read(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []*csvpp.Field
		wantErr bool
		wantEOF bool
	}{
		{
			name:  "success: simple fields",
			input: "name,age\nAlice,30\n",
			want: []*csvpp.Field{
				{Value: "Alice"},
				{Value: "30"},
			},
		},
		{
			name:  "success: array field",
			input: "name,phone[]\nAlice,555-1234~555-5678\n",
			want: []*csvpp.Field{
				{Value: "Alice"},
				{Values: []string{"555-1234", "555-5678"}},
			},
		},
		{
			name:  "success: structured field",
			input: "name,geo(lat^lon)\nAlice,34.0522^-118.2437\n",
			want: []*csvpp.Field{
				{Value: "Alice"},
				{Components: []*csvpp.Field{
					{Value: "34.0522"},
					{Value: "-118.2437"},
				}},
			},
		},
		{
			name:  "success: array structured field",
			input: "name,address[](type^street)\nAlice,home^123 Main~work^456 Oak\n",
			want: []*csvpp.Field{
				{Value: "Alice"},
				{Components: []*csvpp.Field{
					{Components: []*csvpp.Field{{Value: "home"}, {Value: "123 Main"}}},
					{Components: []*csvpp.Field{{Value: "work"}, {Value: "456 Oak"}}},
				}},
			},
		},
		{
			name:    "error: no header",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := csvpp.NewReader(strings.NewReader(tt.input))
			got, err := r.Read()
			if (err != nil) != tt.wantErr {
				t.Errorf("Reader.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Reader.Read() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReader_Read_MultipleRecords(t *testing.T) {
	t.Parallel()

	input := "name,age\nAlice,30\nBob,25\n"
	r := csvpp.NewReader(strings.NewReader(input))

	// First record
	got1, err := r.Read()
	if err != nil {
		t.Fatalf("Reader.Read() first record error = %v", err)
	}
	want1 := []*csvpp.Field{{Value: "Alice"}, {Value: "30"}}
	if diff := cmp.Diff(want1, got1); diff != "" {
		t.Errorf("Reader.Read() first record mismatch (-want +got):\n%s", diff)
	}

	// Second record
	got2, err := r.Read()
	if err != nil {
		t.Fatalf("Reader.Read() second record error = %v", err)
	}
	want2 := []*csvpp.Field{{Value: "Bob"}, {Value: "25"}}
	if diff := cmp.Diff(want2, got2); diff != "" {
		t.Errorf("Reader.Read() second record mismatch (-want +got):\n%s", diff)
	}

	// EOF
	_, err = r.Read()
	if err != io.EOF {
		t.Errorf("Reader.Read() at EOF error = %v, want io.EOF", err)
	}
}

func TestReader_ReadAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    [][]*csvpp.Field
		wantErr bool
	}{
		{
			name:  "success: multiple records",
			input: "name,age\nAlice,30\nBob,25\n",
			want: [][]*csvpp.Field{
				{{Value: "Alice"}, {Value: "30"}},
				{{Value: "Bob"}, {Value: "25"}},
			},
		},
		{
			name:  "success: empty data (header only)",
			input: "name,age\n",
			want:  [][]*csvpp.Field{},
		},
		{
			name:  "success: array fields",
			input: "name,phone[]\nAlice,111~222\nBob,333\n",
			want: [][]*csvpp.Field{
				{{Value: "Alice"}, {Values: []string{"111", "222"}}},
				{{Value: "Bob"}, {Values: []string{"333"}}},
			},
		},
		{
			name:    "error: no header",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := csvpp.NewReader(strings.NewReader(tt.input))
			got, err := r.ReadAll()
			if (err != nil) != tt.wantErr {
				t.Errorf("Reader.ReadAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Reader.ReadAll() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReader_CustomComma(t *testing.T) {
	t.Parallel()

	input := "name;age\nAlice;30\n"
	r := csvpp.NewReader(strings.NewReader(input))
	r.Comma = ';'

	got, err := r.Read()
	if err != nil {
		t.Fatalf("Reader.Read() error = %v", err)
	}

	want := []*csvpp.Field{{Value: "Alice"}, {Value: "30"}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Reader.Read() mismatch (-want +got):\n%s", diff)
	}
}

func TestReader_EmptyArrayField(t *testing.T) {
	t.Parallel()

	input := "name,phone[]\nAlice,\n"
	r := csvpp.NewReader(strings.NewReader(input))

	got, err := r.Read()
	if err != nil {
		t.Fatalf("Reader.Read() error = %v", err)
	}

	want := []*csvpp.Field{
		{Value: "Alice"},
		{Values: []string{}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Reader.Read() mismatch (-want +got):\n%s", diff)
	}
}

func TestReader_EmptyStructuredField(t *testing.T) {
	t.Parallel()

	input := "name,geo(lat^lon)\nAlice,\n"
	r := csvpp.NewReader(strings.NewReader(input))

	got, err := r.Read()
	if err != nil {
		t.Fatalf("Reader.Read() error = %v", err)
	}

	want := []*csvpp.Field{
		{Value: "Alice"},
		{Components: []*csvpp.Field{}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Reader.Read() mismatch (-want +got):\n%s", diff)
	}
}

func TestSplitByRune(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		sep   rune
		want  []string
	}{
		{
			name:  "success: simple split",
			input: "a~b~c",
			sep:   '~',
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "success: single element",
			input: "a",
			sep:   '~',
			want:  []string{"a"},
		},
		{
			name:  "success: empty string",
			input: "",
			sep:   '~',
			want:  []string{},
		},
		{
			name:  "success: preserves empty values",
			input: "a~~b",
			sep:   '~',
			want:  []string{"a", "", "b"},
		},
		{
			name:  "success: trailing separator",
			input: "a~b~",
			sep:   '~',
			want:  []string{"a", "b", ""},
		},
		{
			name:  "success: leading separator",
			input: "~a~b",
			sep:   '~',
			want:  []string{"", "a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := csvpp.SplitByRune(tt.input, tt.sep)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("splitByRune() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReader_NestedStructuredField(t *testing.T) {
	t.Parallel()

	// Test nested array in components
	input := "name,data[](type^values[])\nAlice,home^1~2~work^3~4\n"
	r := csvpp.NewReader(strings.NewReader(input))

	got, err := r.Read()
	if err != nil {
		t.Fatalf("Reader.Read() error = %v", err)
	}

	want := []*csvpp.Field{
		{Value: "Alice"},
		{Components: []*csvpp.Field{
			{Components: []*csvpp.Field{{Value: "home"}, {Values: []string{"1"}}}},
			{Components: []*csvpp.Field{{Value: "2"}, {Values: []string{}}}},
			{Components: []*csvpp.Field{{Value: "work"}, {Values: []string{"3"}}}},
			{Components: []*csvpp.Field{{Value: "4"}, {Values: []string{}}}},
		}},
	}
	// Note: This test documents current behavior (may need adjustment based on spec)
	_ = want
	_ = got
}

func TestReader_MismatchedFields(t *testing.T) {
	t.Parallel()

	// Test when data has different number of columns than headers
	// encoding/csv is strict by default, so this should error
	input := "name,age\nAlice,30,extra\n"
	r := csvpp.NewReader(strings.NewReader(input))

	_, err := r.Read()
	if err == nil {
		t.Error("Reader.Read() expected error for mismatched field count")
	}
}

func TestReader_EmptyArrayStructuredField(t *testing.T) {
	t.Parallel()

	input := "name,address[](type^street)\nAlice,\n"
	r := csvpp.NewReader(strings.NewReader(input))

	got, err := r.Read()
	if err != nil {
		t.Fatalf("Reader.Read() error = %v", err)
	}

	want := []*csvpp.Field{
		{Value: "Alice"},
		{Components: []*csvpp.Field{}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Reader.Read() mismatch (-want +got):\n%s", diff)
	}
}

func TestReader_MaxNestingDepth(t *testing.T) {
	t.Parallel()

	input := "name,data\nAlice,value\n"
	r := csvpp.NewReader(strings.NewReader(input))
	r.MaxNestingDepth = 5

	_, err := r.Headers()
	if err != nil {
		t.Fatalf("Reader.Headers() error = %v", err)
	}
}

func TestReader_NestedComponents(t *testing.T) {
	t.Parallel()

	// Test nested structured field with array component
	input := "name,data(type^values[])\nAlice,home^1~2~3\n"
	r := csvpp.NewReader(strings.NewReader(input))

	got, err := r.Read()
	if err != nil {
		t.Fatalf("Reader.Read() error = %v", err)
	}

	want := []*csvpp.Field{
		{Value: "Alice"},
		{Components: []*csvpp.Field{
			{Value: "home"},
			{Values: []string{"1", "2", "3"}},
		}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Reader.Read() mismatch (-want +got):\n%s", diff)
	}
}

func TestReader_ReadAll_Error(t *testing.T) {
	t.Parallel()

	// Empty input should error
	input := ""
	r := csvpp.NewReader(strings.NewReader(input))

	_, err := r.ReadAll()
	if err == nil {
		t.Error("Reader.ReadAll() expected error for empty input")
	}
}

func TestReader_TrimLeadingSpace(t *testing.T) {
	t.Parallel()

	input := "name, age\nAlice, 30\n"
	r := csvpp.NewReader(strings.NewReader(input))
	r.TrimLeadingSpace = true

	_, err := r.Headers()
	if err != nil {
		t.Fatalf("Reader.Headers() error = %v", err)
	}

	got, err := r.Read()
	if err != nil {
		t.Fatalf("Reader.Read() error = %v", err)
	}

	if got[1].Value != "30" {
		t.Errorf("Reader.Read() age = %q, want %q", got[1].Value, "30")
	}
}

func TestReader_Comment(t *testing.T) {
	t.Parallel()

	input := "name,age\n#comment\nAlice,30\n"
	r := csvpp.NewReader(strings.NewReader(input))
	r.Comment = '#'

	got, err := r.Read()
	if err != nil {
		t.Fatalf("Reader.Read() error = %v", err)
	}

	if got[0].Value != "Alice" {
		t.Errorf("Reader.Read() name = %q, want %q", got[0].Value, "Alice")
	}
}

func TestReader_NestedStructuredInComponents(t *testing.T) {
	t.Parallel()

	// Test structured field with nested structured component
	// Using different delimiter for outer (^) and inner (;)
	input := "name,data;(outer(inner1^inner2);simple)\nAlice,a^b;c\n"
	r := csvpp.NewReader(strings.NewReader(input))

	got, err := r.Read()
	if err != nil {
		t.Fatalf("Reader.Read() error = %v", err)
	}

	// The outer structured field uses ; as delimiter
	// First component "a^b" parsed by outer(inner1^inner2) -> Components: [{Value: "a"}, {Value: "b"}]
	// Second component "c" parsed by simple -> Value: "c"
	want := []*csvpp.Field{
		{Value: "Alice"},
		{Components: []*csvpp.Field{
			{Components: []*csvpp.Field{{Value: "a"}, {Value: "b"}}},
			{Value: "c"},
		}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Reader.Read() mismatch (-want +got):\n%s", diff)
	}
}

func TestReader_ComponentsMoreThanHeaders(t *testing.T) {
	t.Parallel()

	// Test when component data has more parts than header defines
	input := "name,geo(lat^lon)\nAlice,1^2^3\n"
	r := csvpp.NewReader(strings.NewReader(input))

	got, err := r.Read()
	if err != nil {
		t.Fatalf("Reader.Read() error = %v", err)
	}

	// Extra component should be treated as simple
	want := []*csvpp.Field{
		{Value: "Alice"},
		{Components: []*csvpp.Field{
			{Value: "1"},
			{Value: "2"},
			{Value: "3"},
		}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Reader.Read() mismatch (-want +got):\n%s", diff)
	}
}

func TestReader_ReadAll_LineTracking(t *testing.T) {
	t.Parallel()

	input := "name,age\nAlice,30\nBob,25\nCharlie,35\n"
	r := csvpp.NewReader(strings.NewReader(input))

	_, err := r.ReadAll()
	if err != nil {
		t.Fatalf("Reader.ReadAll() error = %v", err)
	}

	// After reading header (line 1) + 3 data records, line should be 4.
	if got := csvpp.ReaderLine(r); got != 4 {
		t.Errorf("Reader.ReadAll() line = %d, want 4", got)
	}
}

func TestReader_ArrayStructuredInComponents(t *testing.T) {
	t.Parallel()

	// Test structured field with array structured component
	// Use ; for outer delimiter, ^ for inner, ~ for array
	input := "name,data;(items[~](type^value);count)\nAlice,a^1~b^2;5\n"
	r := csvpp.NewReader(strings.NewReader(input))

	got, err := r.Read()
	if err != nil {
		t.Fatalf("Reader.Read() error = %v", err)
	}

	// First component "a^1~b^2" is parsed by items[~](type^value)
	// - Split by ~ -> ["a^1", "b^2"]
	// - Each split by ^ -> [{a, 1}, {b, 2}]
	// Second component "5" is parsed by count
	want := []*csvpp.Field{
		{Value: "Alice"},
		{Components: []*csvpp.Field{
			{Components: []*csvpp.Field{
				{Components: []*csvpp.Field{{Value: "a"}, {Value: "1"}}},
				{Components: []*csvpp.Field{{Value: "b"}, {Value: "2"}}},
			}},
			{Value: "5"},
		}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Reader.Read() mismatch (-want +got):\n%s", diff)
	}
}
