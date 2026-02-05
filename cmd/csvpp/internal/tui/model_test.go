package tui_test

import (
	"strings"
	"testing"

	"github.com/osamingo/go-csvpp"
	"github.com/osamingo/go-csvpp/cmd/csvpp/internal/tui"
)

func TestPlainView(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		headers      []*csvpp.ColumnHeader
		records      [][]*csvpp.Field
		wantContains []string
	}{
		{
			name: "success: simple fields",
			headers: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField},
				{Name: "age", Kind: csvpp.SimpleField},
			},
			records: [][]*csvpp.Field{
				{{Value: "Alice"}, {Value: "30"}},
				{{Value: "Bob"}, {Value: "25"}},
			},
			wantContains: []string{"name", "age", "Alice", "30", "Bob", "25"},
		},
		{
			name: "success: array field",
			headers: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField},
				{Name: "phones", Kind: csvpp.ArrayField},
			},
			records: [][]*csvpp.Field{
				{{Value: "Alice"}, {Values: []string{"111", "222"}}},
			},
			wantContains: []string{"name", "phones[]", "Alice", "111", "222"},
		},
		{
			name: "success: structured field",
			headers: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField},
				{
					Name: "geo",
					Kind: csvpp.StructuredField,
					Components: []*csvpp.ColumnHeader{
						{Name: "lat", Kind: csvpp.SimpleField},
						{Name: "lon", Kind: csvpp.SimpleField},
					},
				},
			},
			records: [][]*csvpp.Field{
				{
					{Value: "Alice"},
					{Components: []*csvpp.Field{{Value: "34.05"}, {Value: "-118.24"}}},
				},
			},
			wantContains: []string{"name", "geo(lat,lon)", "Alice", "lat:", "lon:", "34.05", "-118.24"},
		},
		{
			name: "success: array structured field",
			headers: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField},
				{
					Name: "addresses",
					Kind: csvpp.ArrayStructuredField,
					Components: []*csvpp.ColumnHeader{
						{Name: "street", Kind: csvpp.SimpleField},
						{Name: "city", Kind: csvpp.SimpleField},
					},
				},
			},
			records: [][]*csvpp.Field{
				{
					{Value: "Alice"},
					{Components: []*csvpp.Field{
						{Components: []*csvpp.Field{{Value: "123 Main"}, {Value: "LA"}}},
						{Components: []*csvpp.Field{{Value: "456 Oak"}, {Value: "NY"}}},
					}},
				},
			},
			wantContains: []string{"name", "addresses[](street,city)", "Alice", "123 Main", "LA", "456 Oak", "NY"},
		},
		{
			name:         "success: empty records",
			headers:      []*csvpp.ColumnHeader{{Name: "name", Kind: csvpp.SimpleField}},
			records:      [][]*csvpp.Field{},
			wantContains: []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tui.PlainView(tt.headers, tt.records)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q:\n%s", want, got)
				}
			}
		})
	}
}

func TestNewModel(t *testing.T) {
	t.Parallel()

	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "age", Kind: csvpp.SimpleField},
	}
	records := [][]*csvpp.Field{
		{{Value: "Alice"}, {Value: "30"}},
	}

	model := tui.NewModel(headers, records)

	// Model should be initialized without panic
	view := model.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}
