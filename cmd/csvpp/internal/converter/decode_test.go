package converter_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/osamingo/go-csvpp"
	"github.com/osamingo/go-csvpp/cmd/csvpp/internal/converter"
)

func TestFromJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		wantHeaders []*csvpp.ColumnHeader
		wantRecords [][]*csvpp.Field
		wantErr     bool
	}{
		{
			name:  "success: simple fields with preserved order",
			input: `[{"name":"Alice","age":"30"},{"name":"Bob","age":"25"}]`,
			wantHeaders: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
				{Name: "age", Kind: csvpp.SimpleField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
			},
			wantRecords: [][]*csvpp.Field{
				{{Value: "Alice"}, {Value: "30"}},
				{{Value: "Bob"}, {Value: "25"}},
			},
		},
		{
			name:  "success: array field with preserved order",
			input: `[{"name":"Alice","phones":["111","222"]},{"name":"Bob","phones":["333"]}]`,
			wantHeaders: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
				{Name: "phones", Kind: csvpp.ArrayField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
			},
			wantRecords: [][]*csvpp.Field{
				{{Value: "Alice"}, {Values: []string{"111", "222"}}},
				{{Value: "Bob"}, {Values: []string{"333"}}},
			},
		},
		{
			name:        "success: empty array",
			input:       `[]`,
			wantHeaders: nil,
			wantRecords: nil,
		},
		{
			name:    "error: invalid json",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			headers, records, err := converter.FromJSON(strings.NewReader(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if diff := cmp.Diff(tt.wantHeaders, headers); diff != "" {
				t.Errorf("headers mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.wantRecords, records); diff != "" {
				t.Errorf("records mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFromYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		wantHeaders []*csvpp.ColumnHeader
		wantRecords [][]*csvpp.Field
		wantErr     bool
	}{
		{
			name: "success: simple fields with preserved order",
			input: `- name: Alice
  age: "30"
- name: Bob
  age: "25"
`,
			wantHeaders: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
				{Name: "age", Kind: csvpp.SimpleField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
			},
			wantRecords: [][]*csvpp.Field{
				{{Value: "Alice"}, {Value: "30"}},
				{{Value: "Bob"}, {Value: "25"}},
			},
		},
		{
			name:        "success: empty array",
			input:       `[]`,
			wantHeaders: nil,
			wantRecords: nil,
		},
		{
			name:    "error: invalid yaml",
			input:   `{: invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			headers, records, err := converter.FromYAML(strings.NewReader(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if diff := cmp.Diff(tt.wantHeaders, headers); diff != "" {
				t.Errorf("headers mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.wantRecords, records); diff != "" {
				t.Errorf("records mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFromJSONStructuredField(t *testing.T) {
	t.Parallel()

	input := `[{"name":"Alice","geo":{"lat":"34.05","lon":"-118.24"}}]`

	headers, records, err := converter.FromJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantHeaders := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
		{
			Name: "geo", Kind: csvpp.StructuredField, ArrayDelimiter: '~', ComponentDelimiter: '^',
			Components: []*csvpp.ColumnHeader{
				{Name: "lat", Kind: csvpp.SimpleField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
				{Name: "lon", Kind: csvpp.SimpleField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
			},
		},
	}

	if diff := cmp.Diff(wantHeaders, headers); diff != "" {
		t.Errorf("headers mismatch (-want +got):\n%s", diff)
	}

	wantRecords := [][]*csvpp.Field{
		{
			{Value: "Alice"},
			{Components: []*csvpp.Field{{Value: "34.05"}, {Value: "-118.24"}}},
		},
	}

	if diff := cmp.Diff(wantRecords, records); diff != "" {
		t.Errorf("records mismatch (-want +got):\n%s", diff)
	}
}

func TestFromJSONArrayStructuredField(t *testing.T) {
	t.Parallel()

	input := `[{"name":"Alice","addresses":[{"street":"123 Main","city":"LA"},{"street":"456 Oak","city":"NY"}]}]`

	headers, records, err := converter.FromJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantHeaders := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
		{
			Name: "addresses", Kind: csvpp.ArrayStructuredField, ArrayDelimiter: '~', ComponentDelimiter: '^',
			Components: []*csvpp.ColumnHeader{
				{Name: "street", Kind: csvpp.SimpleField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
				{Name: "city", Kind: csvpp.SimpleField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
			},
		},
	}

	if diff := cmp.Diff(wantHeaders, headers); diff != "" {
		t.Errorf("headers mismatch (-want +got):\n%s", diff)
	}

	wantRecords := [][]*csvpp.Field{
		{
			{Value: "Alice"},
			{Components: []*csvpp.Field{
				{Components: []*csvpp.Field{{Value: "123 Main"}, {Value: "LA"}}},
				{Components: []*csvpp.Field{{Value: "456 Oak"}, {Value: "NY"}}},
			}},
		},
	}

	if diff := cmp.Diff(wantRecords, records); diff != "" {
		t.Errorf("records mismatch (-want +got):\n%s", diff)
	}
}

func TestToString(t *testing.T) {
	t.Parallel()

	// This tests the toString function indirectly through FromJSON
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "success: integer value",
			input: `[{"value":42}]`,
			want:  "42",
		},
		{
			name:  "success: float value",
			input: `[{"value":3.14}]`,
			want:  "3.14",
		},
		{
			name:  "success: boolean value",
			input: `[{"value":true}]`,
			want:  "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			headers, records, err := converter.FromJSON(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(records) != 1 || len(records[0]) != 1 {
				t.Fatalf("unexpected records structure")
			}

			// Find the field value
			var got string
			for i, h := range headers {
				if h.Name == "value" {
					got = records[0][i].Value
					break
				}
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("value mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
