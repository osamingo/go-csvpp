package tui_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/google/go-cmp/cmp"

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

func TestParseFilterQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want tui.FilterQuery
	}{
		{
			name: "success: empty string",
			in:   "",
			want: tui.FilterQuery{},
		},
		{
			name: "success: whitespace only",
			in:   "   ",
			want: tui.FilterQuery{},
		},
		{
			name: "success: simple text searches all columns",
			in:   "alice",
			want: tui.FilterQuery{Column: "", Value: "alice"},
		},
		{
			name: "success: preserves lowercase conversion",
			in:   "ALICE",
			want: tui.FilterQuery{Column: "", Value: "alice"},
		},
		{
			name: "success: column specific search",
			in:   "name:alice",
			want: tui.FilterQuery{Column: "name", Value: "alice"},
		},
		{
			name: "success: column specific with spaces trimmed",
			in:   " name : alice ",
			want: tui.FilterQuery{Column: "name", Value: "alice"},
		},
		{
			name: "success: column name is lowercased",
			in:   "Name:alice",
			want: tui.FilterQuery{Column: "name", Value: "alice"},
		},
		{
			name: "success: empty column part searches all columns",
			in:   ":alice",
			want: tui.FilterQuery{Column: "", Value: "alice"},
		},
		{
			name: "success: column with empty value",
			in:   "name:",
			want: tui.FilterQuery{Column: "name", Value: ""},
		},
		{
			name: "success: value with colon inside",
			in:   "name:a:b",
			want: tui.FilterQuery{Column: "name", Value: "a:b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tui.ParseFilterQuery(tt.in)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseFilterQuery(%q) mismatch (-want +got):\n%s", tt.in, diff)
			}
		})
	}
}

func TestMatchesFilter(t *testing.T) {
	t.Parallel()

	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "age", Kind: csvpp.SimpleField},
		{Name: "city", Kind: csvpp.SimpleField},
	}

	// row[0] = selection marker, row[1..] = data columns
	row := table.Row{" ", "Alice", "30", "Tokyo"}

	tests := []struct {
		name  string
		query tui.FilterQuery
		row   table.Row
		want  bool
	}{
		{
			name:  "success: empty query matches all",
			query: tui.FilterQuery{Value: ""},
			row:   row,
			want:  true,
		},
		{
			name:  "success: all columns match by name",
			query: tui.FilterQuery{Value: "alice"},
			row:   row,
			want:  true,
		},
		{
			name:  "success: all columns match by city",
			query: tui.FilterQuery{Value: "tokyo"},
			row:   row,
			want:  true,
		},
		{
			name:  "success: case insensitive match on row value",
			query: tui.FilterQuery{Value: "alice"},
			row:   table.Row{" ", "ALICE", "30", "Tokyo"},
			want:  true,
		},
		{
			name:  "success: partial match",
			query: tui.FilterQuery{Value: "ali"},
			row:   row,
			want:  true,
		},
		{
			name:  "error: no match in any column",
			query: tui.FilterQuery{Value: "xyz"},
			row:   row,
			want:  false,
		},
		{
			name:  "success: column specific match",
			query: tui.FilterQuery{Column: "name", Value: "alice"},
			row:   row,
			want:  true,
		},
		{
			name:  "error: column specific no match",
			query: tui.FilterQuery{Column: "name", Value: "tokyo"},
			row:   row,
			want:  false,
		},
		{
			name:  "error: nonexistent column",
			query: tui.FilterQuery{Column: "email", Value: "alice"},
			row:   row,
			want:  false,
		},
		{
			name:  "success: column specific match by age",
			query: tui.FilterQuery{Column: "age", Value: "30"},
			row:   row,
			want:  true,
		},
		{
			name:  "error: selection marker not searched",
			query: tui.FilterQuery{Value: "✓"},
			row:   table.Row{"✓", "Alice", "30", "Tokyo"},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tui.MatchesFilter(tt.query, headers, tt.row)
			if got != tt.want {
				t.Errorf("MatchesFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
