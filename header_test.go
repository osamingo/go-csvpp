package csvpp_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/osamingo/go-csvpp"
)

func TestParseColumnHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    *csvpp.ColumnHeader
		wantErr bool
	}{
		{
			name:  "success: simple field",
			input: "name",
			want: &csvpp.ColumnHeader{
				Name: "name",
				Kind: csvpp.SimpleField,
			},
		},
		{
			name:  "success: simple field with underscore",
			input: "first_name",
			want: &csvpp.ColumnHeader{
				Name: "first_name",
				Kind: csvpp.SimpleField,
			},
		},
		{
			name:  "success: simple field with hyphen",
			input: "first-name",
			want: &csvpp.ColumnHeader{
				Name: "first-name",
				Kind: csvpp.SimpleField,
			},
		},
		{
			name:  "success: simple field with digits",
			input: "field123",
			want: &csvpp.ColumnHeader{
				Name: "field123",
				Kind: csvpp.SimpleField,
			},
		},
		{
			name:  "success: array field with default delimiter",
			input: "phone[]",
			want: &csvpp.ColumnHeader{
				Name:           "phone",
				Kind:           csvpp.ArrayField,
				ArrayDelimiter: csvpp.DefaultArrayDelimiter,
			},
		},
		{
			name:  "success: array field with custom delimiter",
			input: "phone[|]",
			want: &csvpp.ColumnHeader{
				Name:           "phone",
				Kind:           csvpp.ArrayField,
				ArrayDelimiter: '|',
			},
		},
		{
			name:  "success: structured field with default delimiter",
			input: "geo(lat^lon)",
			want: &csvpp.ColumnHeader{
				Name:               "geo",
				Kind:               csvpp.StructuredField,
				ComponentDelimiter: csvpp.DefaultComponentDelimiter,
				Components: []*csvpp.ColumnHeader{
					{Name: "lat", Kind: csvpp.SimpleField},
					{Name: "lon", Kind: csvpp.SimpleField},
				},
			},
		},
		{
			name:  "success: structured field with custom delimiter",
			input: "geo;(lat;lon)",
			want: &csvpp.ColumnHeader{
				Name:               "geo",
				Kind:               csvpp.StructuredField,
				ComponentDelimiter: ';',
				Components: []*csvpp.ColumnHeader{
					{Name: "lat", Kind: csvpp.SimpleField},
					{Name: "lon", Kind: csvpp.SimpleField},
				},
			},
		},
		{
			name:  "success: array structured field",
			input: "address[](type^street)",
			want: &csvpp.ColumnHeader{
				Name:               "address",
				Kind:               csvpp.ArrayStructuredField,
				ArrayDelimiter:     csvpp.DefaultArrayDelimiter,
				ComponentDelimiter: csvpp.DefaultComponentDelimiter,
				Components: []*csvpp.ColumnHeader{
					{Name: "type", Kind: csvpp.SimpleField},
					{Name: "street", Kind: csvpp.SimpleField},
				},
			},
		},
		{
			name:  "success: array structured field with custom delimiters",
			input: "address[|];(type;street)",
			want: &csvpp.ColumnHeader{
				Name:               "address",
				Kind:               csvpp.ArrayStructuredField,
				ArrayDelimiter:     '|',
				ComponentDelimiter: ';',
				Components: []*csvpp.ColumnHeader{
					{Name: "type", Kind: csvpp.SimpleField},
					{Name: "street", Kind: csvpp.SimpleField},
				},
			},
		},
		{
			name:    "error: empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "error: missing closing bracket",
			input:   "phone[",
			wantErr: true,
		},
		{
			name:    "error: missing closing parenthesis",
			input:   "geo(lat^lon",
			wantErr: true,
		},
		{
			name:    "error: empty component list",
			input:   "geo()",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := csvpp.ParseColumnHeader(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseColumnHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("parseColumnHeader() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantName string
		wantRest string
		wantErr  bool
	}{
		{
			name:     "success: simple name",
			input:    "name",
			wantName: "name",
			wantRest: "",
		},
		{
			name:     "success: name with rest",
			input:    "phone[]",
			wantName: "phone",
			wantRest: "[]",
		},
		{
			name:     "success: name with parenthesis",
			input:    "geo(lat^lon)",
			wantName: "geo",
			wantRest: "(lat^lon)",
		},
		{
			name:     "success: name with underscore",
			input:    "first_name",
			wantName: "first_name",
			wantRest: "",
		},
		{
			name:     "success: name with hyphen",
			input:    "first-name",
			wantName: "first-name",
			wantRest: "",
		},
		{
			name:     "success: name with digits",
			input:    "field123[]",
			wantName: "field123",
			wantRest: "[]",
		},
		{
			name:    "error: empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "error: starts with special character",
			input:   "[name]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotName, gotRest, err := csvpp.ParseName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.wantName, gotName); diff != "" {
				t.Errorf("parseName() name mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantRest, gotRest); diff != "" {
				t.Errorf("parseName() rest mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseArrayDelimiter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantDelim rune
		wantRest  string
		wantErr   bool
	}{
		{
			name:      "success: default delimiter",
			input:     "[]",
			wantDelim: csvpp.DefaultArrayDelimiter,
			wantRest:  "",
		},
		{
			name:      "success: custom delimiter",
			input:     "[|]",
			wantDelim: '|',
			wantRest:  "",
		},
		{
			name:      "success: with rest",
			input:     "[](lat^lon)",
			wantDelim: csvpp.DefaultArrayDelimiter,
			wantRest:  "(lat^lon)",
		},
		{
			name:      "success: no bracket",
			input:     "(lat^lon)",
			wantDelim: 0,
			wantRest:  "(lat^lon)",
		},
		{
			name:    "error: missing closing bracket",
			input:   "[|",
			wantErr: true,
		},
		{
			name:    "error: multiple characters as delimiter",
			input:   "[||]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotDelim, gotRest, err := csvpp.ParseArrayDelimiter(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseArrayDelimiter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if gotDelim != tt.wantDelim {
				t.Errorf("parseArrayDelimiter() delim = %v, want %v", gotDelim, tt.wantDelim)
			}
			if diff := cmp.Diff(tt.wantRest, gotRest); diff != "" {
				t.Errorf("parseArrayDelimiter() rest mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIsFieldChar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input rune
		want  bool
	}{
		{name: "success: lowercase letter", input: 'a', want: true},
		{name: "success: uppercase letter", input: 'Z', want: true},
		{name: "success: digit", input: '5', want: true},
		{name: "success: underscore", input: '_', want: true},
		{name: "success: hyphen", input: '-', want: true},
		{name: "success: invalid bracket", input: '[', want: false},
		{name: "success: invalid parenthesis", input: '(', want: false},
		{name: "success: invalid caret", input: '^', want: false},
		{name: "success: invalid space", input: ' ', want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := csvpp.IsFieldChar(tt.input)
			if got != tt.want {
				t.Errorf("isFieldChar(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSplitByDelimiter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		delim rune
		want  []string
	}{
		{
			name:  "success: simple split",
			input: "a^b^c",
			delim: '^',
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "success: single element",
			input: "a",
			delim: '^',
			want:  []string{"a"},
		},
		{
			name:  "success: empty string",
			input: "",
			delim: '^',
			want:  nil,
		},
		{
			name:  "success: ignore delimiter inside parentheses",
			input: "a^nested(b^c)^d",
			delim: '^',
			want:  []string{"a", "nested(b^c)", "d"},
		},
		{
			name:  "success: nested parentheses",
			input: "a^outer(inner(x^y)^z)^b",
			delim: '^',
			want:  []string{"a", "outer(inner(x^y)^z)", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := csvpp.SplitByDelimiter(tt.input, tt.delim)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("splitByDelimiter() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseHeaderRecordWithMaxDepth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fields   []string
		maxDepth int
		want     []*csvpp.ColumnHeader
		wantErr  bool
	}{
		{
			name:     "success: simple fields",
			fields:   []string{"name", "age"},
			maxDepth: 10,
			want: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField},
				{Name: "age", Kind: csvpp.SimpleField},
			},
		},
		{
			name:     "success: mixed fields",
			fields:   []string{"name", "phone[]", "geo(lat^lon)"},
			maxDepth: 10,
			want: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField},
				{Name: "phone", Kind: csvpp.ArrayField, ArrayDelimiter: csvpp.DefaultArrayDelimiter},
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
			name:     "error: empty fields",
			fields:   []string{},
			maxDepth: 10,
			wantErr:  true,
		},
		{
			name:     "error: invalid header",
			fields:   []string{"name", ""},
			maxDepth: 10,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := csvpp.ParseHeaderRecordWithMaxDepth(tt.fields, tt.maxDepth)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHeaderRecordWithMaxDepth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("parseHeaderRecordWithMaxDepth() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseColumnHeaderWithDepth_NestingLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		maxDepth int
		wantErr  bool
	}{
		{
			name:     "success: within depth limit",
			input:    "a(b(c))",
			maxDepth: 3,
			wantErr:  false,
		},
		{
			name:     "error: exceeds depth limit",
			input:    "a(b(c(d)))",
			maxDepth: 2,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := csvpp.ParseColumnHeaderWithDepth(tt.input, 0, tt.maxDepth)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseColumnHeaderWithDepth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
