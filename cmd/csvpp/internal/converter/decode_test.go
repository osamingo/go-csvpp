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
			name:  "success: simple fields",
			input: `[{"name":"Alice","age":"30"},{"name":"Bob","age":"25"}]`,
			wantHeaders: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
				{Name: "age", Kind: csvpp.SimpleField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
			},
			wantRecords: [][]*csvpp.Field{
				{{Value: "Alice"}, {Value: "30"}},
				{{Value: "Bob"}, {Value: "25"}},
			},
			wantErr: false,
		},
		{
			name:  "success: array field",
			input: `[{"name":"Alice","phones":["111","222"]},{"name":"Bob","phones":["333"]}]`,
			wantHeaders: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
				{Name: "phones", Kind: csvpp.ArrayField, ArrayDelimiter: '~', ComponentDelimiter: '^'},
			},
			wantRecords: [][]*csvpp.Field{
				{{Value: "Alice"}, {Values: []string{"111", "222"}}},
				{{Value: "Bob"}, {Values: []string{"333"}}},
			},
			wantErr: false,
		},
		{
			name:        "success: empty array",
			input:       `[]`,
			wantHeaders: nil,
			wantRecords: nil,
			wantErr:     false,
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

			// Compare headers (ignore order since JSON doesn't preserve it)
			if len(headers) != len(tt.wantHeaders) {
				t.Errorf("headers count mismatch: want %d, got %d", len(tt.wantHeaders), len(headers))
			}

			// Compare records count
			if len(records) != len(tt.wantRecords) {
				t.Errorf("records count mismatch: want %d, got %d", len(tt.wantRecords), len(records))
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
			name: "success: simple fields",
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
			wantErr: false,
		},
		{
			name:        "success: empty array",
			input:       `[]`,
			wantHeaders: nil,
			wantRecords: nil,
			wantErr:     false,
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

			// Compare headers count
			if len(headers) != len(tt.wantHeaders) {
				t.Errorf("headers count mismatch: want %d, got %d", len(tt.wantHeaders), len(headers))
			}

			// Compare headers by name (order not guaranteed due to map iteration)
			if len(tt.wantHeaders) > 0 {
				headerMap := make(map[string]*csvpp.ColumnHeader)
				for _, h := range headers {
					headerMap[h.Name] = h
				}
				for _, want := range tt.wantHeaders {
					got, ok := headerMap[want.Name]
					if !ok {
						t.Errorf("header %q not found", want.Name)
						continue
					}
					if got.Kind != want.Kind {
						t.Errorf("header %q kind mismatch: want %v, got %v", want.Name, want.Kind, got.Kind)
					}
				}
			}

			// Compare records count
			if len(records) != len(tt.wantRecords) {
				t.Errorf("records count mismatch: want %d, got %d", len(tt.wantRecords), len(records))
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

	// Find geo header
	var geoHeader *csvpp.ColumnHeader
	for _, h := range headers {
		if h.Name == "geo" {
			geoHeader = h
			break
		}
	}

	if geoHeader == nil {
		t.Fatal("geo header not found")
	}

	if geoHeader.Kind != csvpp.StructuredField {
		t.Errorf("geo kind mismatch: want StructuredField, got %v", geoHeader.Kind)
	}

	if len(geoHeader.Components) != 2 {
		t.Errorf("geo components count mismatch: want 2, got %d", len(geoHeader.Components))
	}

	if len(records) != 1 {
		t.Errorf("records count mismatch: want 1, got %d", len(records))
	}
}

func TestFromJSONArrayStructuredField(t *testing.T) {
	t.Parallel()

	input := `[{"name":"Alice","addresses":[{"street":"123 Main","city":"LA"},{"street":"456 Oak","city":"NY"}]}]`

	headers, records, err := converter.FromJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find addresses header
	var addrHeader *csvpp.ColumnHeader
	for _, h := range headers {
		if h.Name == "addresses" {
			addrHeader = h
			break
		}
	}

	if addrHeader == nil {
		t.Fatal("addresses header not found")
	}

	if addrHeader.Kind != csvpp.ArrayStructuredField {
		t.Errorf("addresses kind mismatch: want ArrayStructuredField, got %v", addrHeader.Kind)
	}

	if len(addrHeader.Components) != 2 {
		t.Errorf("addresses components count mismatch: want 2, got %d", len(addrHeader.Components))
	}

	if len(records) != 1 {
		t.Errorf("records count mismatch: want 1, got %d", len(records))
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
