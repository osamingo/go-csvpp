package csvpputil_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/osamingo/go-csvpp"
	"github.com/osamingo/go-csvpp/csvpputil"
)

func TestRecordToMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		record  []*csvpp.Field
		headers []*csvpp.ColumnHeader
		want    map[string]any
	}{
		{
			name:    "success: nil record",
			record:  nil,
			headers: []*csvpp.ColumnHeader{{Name: "name", Kind: csvpp.SimpleField}},
			want:    nil,
		},
		{
			name:    "success: nil headers",
			record:  []*csvpp.Field{{Value: "Alice"}},
			headers: nil,
			want:    nil,
		},
		{
			name:   "success: simple field",
			record: []*csvpp.Field{{Value: "Alice"}},
			headers: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField},
			},
			want: map[string]any{"name": "Alice"},
		},
		{
			name:   "success: array field",
			record: []*csvpp.Field{{Values: []string{"go", "rust"}}},
			headers: []*csvpp.ColumnHeader{
				{Name: "tags", Kind: csvpp.ArrayField},
			},
			want: map[string]any{"tags": []string{"go", "rust"}},
		},
		{
			name: "success: structured field",
			record: []*csvpp.Field{
				{
					Components: []*csvpp.Field{
						{Value: "35.6"},
						{Value: "139.7"},
					},
				},
			},
			headers: []*csvpp.ColumnHeader{
				{
					Name: "geo",
					Kind: csvpp.StructuredField,
					Components: []*csvpp.ColumnHeader{
						{Name: "lat", Kind: csvpp.SimpleField},
						{Name: "lon", Kind: csvpp.SimpleField},
					},
				},
			},
			want: map[string]any{
				"geo": map[string]any{"lat": "35.6", "lon": "139.7"},
			},
		},
		{
			name: "success: array structured field",
			record: []*csvpp.Field{
				{
					Components: []*csvpp.Field{
						{
							Components: []*csvpp.Field{
								{Value: "Tokyo"},
								{Value: "Japan"},
							},
						},
						{
							Components: []*csvpp.Field{
								{Value: "Osaka"},
								{Value: "Japan"},
							},
						},
					},
				},
			},
			headers: []*csvpp.ColumnHeader{
				{
					Name: "addresses",
					Kind: csvpp.ArrayStructuredField,
					Components: []*csvpp.ColumnHeader{
						{Name: "city", Kind: csvpp.SimpleField},
						{Name: "country", Kind: csvpp.SimpleField},
					},
				},
			},
			want: map[string]any{
				"addresses": []map[string]any{
					{"city": "Tokyo", "country": "Japan"},
					{"city": "Osaka", "country": "Japan"},
				},
			},
		},
		{
			name: "success: mixed fields",
			record: []*csvpp.Field{
				{Value: "Alice"},
				{Values: []string{"go", "rust"}},
				{
					Components: []*csvpp.Field{
						{Value: "35.6"},
						{Value: "139.7"},
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
			want: map[string]any{
				"name": "Alice",
				"tags": []string{"go", "rust"},
				"geo":  map[string]any{"lat": "35.6", "lon": "139.7"},
			},
		},
		{
			name:   "success: more records than headers",
			record: []*csvpp.Field{{Value: "Alice"}, {Value: "extra"}},
			headers: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField},
			},
			want: map[string]any{"name": "Alice"},
		},
		{
			name:   "success: more headers than records",
			record: []*csvpp.Field{{Value: "Alice"}},
			headers: []*csvpp.ColumnHeader{
				{Name: "name", Kind: csvpp.SimpleField},
				{Name: "age", Kind: csvpp.SimpleField},
			},
			want: map[string]any{"name": "Alice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := csvpputil.RecordToMap(tt.headers, tt.record)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("RecordToMap() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
