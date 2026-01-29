package csvpp_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/osamingo/go-csvpp"
)

type SimpleRecord struct {
	Name string `csvpp:"name"`
	Age  int    `csvpp:"age"`
}

type ArrayRecord struct {
	Name   string   `csvpp:"name"`
	Phones []string `csvpp:"phone[]"`
}

type GeoLocation struct {
	Lat float64
	Lon float64
}

type StructuredRecord struct {
	Name string      `csvpp:"name"`
	Geo  GeoLocation `csvpp:"geo(lat^lon)"`
}

type Address struct {
	Type   string
	Street string
}

type ArrayStructuredRecord struct {
	Name      string    `csvpp:"name"`
	Addresses []Address `csvpp:"address[](type^street)"`
}

type IgnoredFieldRecord struct {
	Name    string `csvpp:"name"`
	Age     int    `csvpp:"age"`
	ignored string
	Skipped string `csvpp:"-"`
}

type AllTypesRecord struct {
	Name   string  `csvpp:"name"`
	Age    int     `csvpp:"age"`
	Score  uint    `csvpp:"score"`
	Height float64 `csvpp:"height"`
	Active bool    `csvpp:"active"`
}

type PointerStructRecord struct {
	Name string       `csvpp:"name"`
	Geo  *GeoLocation `csvpp:"geo(lat^lon)"`
}

type NestedArrayRecord struct {
	Name   string `csvpp:"name"`
	Scores []int  `csvpp:"scores[]"`
}

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	t.Run("success: simple record", func(t *testing.T) {
		t.Parallel()

		input := "name,age\nAlice,30\nBob,25\n"
		var records []SimpleRecord

		err := csvpp.Unmarshal(strings.NewReader(input), &records)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		want := []SimpleRecord{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
		}
		if diff := cmp.Diff(want, records); diff != "" {
			t.Errorf("Unmarshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: array record", func(t *testing.T) {
		t.Parallel()

		input := "name,phone[]\nAlice,111~222\nBob,333\n"
		var records []ArrayRecord

		err := csvpp.Unmarshal(strings.NewReader(input), &records)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		want := []ArrayRecord{
			{Name: "Alice", Phones: []string{"111", "222"}},
			{Name: "Bob", Phones: []string{"333"}},
		}
		if diff := cmp.Diff(want, records); diff != "" {
			t.Errorf("Unmarshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: structured record", func(t *testing.T) {
		t.Parallel()

		input := "name,geo(lat^lon)\nAlice,34.0522^-118.2437\n"
		var records []StructuredRecord

		err := csvpp.Unmarshal(strings.NewReader(input), &records)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		want := []StructuredRecord{
			{Name: "Alice", Geo: GeoLocation{Lat: 34.0522, Lon: -118.2437}},
		}
		if diff := cmp.Diff(want, records); diff != "" {
			t.Errorf("Unmarshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: array structured record", func(t *testing.T) {
		t.Parallel()

		input := "name,address[](type^street)\nAlice,home^123 Main~work^456 Oak\n"
		var records []ArrayStructuredRecord

		err := csvpp.Unmarshal(strings.NewReader(input), &records)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		want := []ArrayStructuredRecord{
			{
				Name: "Alice",
				Addresses: []Address{
					{Type: "home", Street: "123 Main"},
					{Type: "work", Street: "456 Oak"},
				},
			},
		}
		if diff := cmp.Diff(want, records); diff != "" {
			t.Errorf("Unmarshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: ignored fields", func(t *testing.T) {
		t.Parallel()

		input := "name,age\nAlice,30\n"
		var records []IgnoredFieldRecord

		err := csvpp.Unmarshal(strings.NewReader(input), &records)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		want := []IgnoredFieldRecord{
			{Name: "Alice", Age: 30},
		}
		if diff := cmp.Diff(want, records, cmpopts.IgnoreUnexported(IgnoredFieldRecord{})); diff != "" {
			t.Errorf("Unmarshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: empty input (header only)", func(t *testing.T) {
		t.Parallel()

		input := "name,age\n"
		var records []SimpleRecord

		err := csvpp.Unmarshal(strings.NewReader(input), &records)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if len(records) != 0 {
			t.Errorf("Unmarshal() len = %d, want 0", len(records))
		}
	})

	t.Run("error: dst is not a pointer", func(t *testing.T) {
		t.Parallel()

		input := "name,age\nAlice,30\n"
		var records []SimpleRecord

		err := csvpp.Unmarshal(strings.NewReader(input), records)
		if err == nil {
			t.Error("Unmarshal() expected error for non-pointer dst")
		}
	})

	t.Run("error: dst is not a pointer to slice", func(t *testing.T) {
		t.Parallel()

		input := "name,age\nAlice,30\n"
		var record SimpleRecord

		err := csvpp.Unmarshal(strings.NewReader(input), &record)
		if err == nil {
			t.Error("Unmarshal() expected error for non-slice dst")
		}
	})

	t.Run("error: slice element is not a struct", func(t *testing.T) {
		t.Parallel()

		input := "name,age\nAlice,30\n"
		var records []string

		err := csvpp.Unmarshal(strings.NewReader(input), &records)
		if err == nil {
			t.Error("Unmarshal() expected error for non-struct element")
		}
	})

	t.Run("success: all types record", func(t *testing.T) {
		t.Parallel()

		input := "name,age,score,height,active\nAlice,30,100,1.65,true\nBob,25,85,1.80,false\n"
		var records []AllTypesRecord

		err := csvpp.Unmarshal(strings.NewReader(input), &records)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		want := []AllTypesRecord{
			{Name: "Alice", Age: 30, Score: 100, Height: 1.65, Active: true},
			{Name: "Bob", Age: 25, Score: 85, Height: 1.80, Active: false},
		}
		if diff := cmp.Diff(want, records); diff != "" {
			t.Errorf("Unmarshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: all types with empty values", func(t *testing.T) {
		t.Parallel()

		input := "name,age,score,height,active\nAlice,,,,"
		var records []AllTypesRecord

		err := csvpp.Unmarshal(strings.NewReader(input), &records)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		want := []AllTypesRecord{
			{Name: "Alice", Age: 0, Score: 0, Height: 0, Active: false},
		}
		if diff := cmp.Diff(want, records); diff != "" {
			t.Errorf("Unmarshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: pointer struct record", func(t *testing.T) {
		t.Parallel()

		input := "name,geo(lat^lon)\nAlice,34.0522^-118.2437\n"
		var records []PointerStructRecord

		err := csvpp.Unmarshal(strings.NewReader(input), &records)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		want := []PointerStructRecord{
			{Name: "Alice", Geo: &GeoLocation{Lat: 34.0522, Lon: -118.2437}},
		}
		if diff := cmp.Diff(want, records); diff != "" {
			t.Errorf("Unmarshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: nested array of int", func(t *testing.T) {
		t.Parallel()

		input := "name,scores[]\nAlice,90~85~95\n"
		var records []NestedArrayRecord

		err := csvpp.Unmarshal(strings.NewReader(input), &records)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		want := []NestedArrayRecord{
			{Name: "Alice", Scores: []int{90, 85, 95}},
		}
		if diff := cmp.Diff(want, records); diff != "" {
			t.Errorf("Unmarshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("error: invalid int value", func(t *testing.T) {
		t.Parallel()

		input := "name,age\nAlice,invalid\n"
		var records []SimpleRecord

		err := csvpp.Unmarshal(strings.NewReader(input), &records)
		if err == nil {
			t.Error("Unmarshal() expected error for invalid int")
		}
	})

	t.Run("error: invalid uint value", func(t *testing.T) {
		t.Parallel()

		input := "name,age,score,height,active\nAlice,30,invalid,1.65,true\n"
		var records []AllTypesRecord

		err := csvpp.Unmarshal(strings.NewReader(input), &records)
		if err == nil {
			t.Error("Unmarshal() expected error for invalid uint")
		}
	})

	t.Run("error: invalid float value", func(t *testing.T) {
		t.Parallel()

		input := "name,age,score,height,active\nAlice,30,100,invalid,true\n"
		var records []AllTypesRecord

		err := csvpp.Unmarshal(strings.NewReader(input), &records)
		if err == nil {
			t.Error("Unmarshal() expected error for invalid float")
		}
	})

	t.Run("error: invalid bool value", func(t *testing.T) {
		t.Parallel()

		input := "name,age,score,height,active\nAlice,30,100,1.65,invalid\n"
		var records []AllTypesRecord

		err := csvpp.Unmarshal(strings.NewReader(input), &records)
		if err == nil {
			t.Error("Unmarshal() expected error for invalid bool")
		}
	})
}

func TestMarshal(t *testing.T) {
	t.Parallel()

	t.Run("success: simple record", func(t *testing.T) {
		t.Parallel()

		records := []SimpleRecord{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
		}

		var buf bytes.Buffer
		err := csvpp.Marshal(&buf, records)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		want := "name,age\nAlice,30\nBob,25\n"
		got := buf.String()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Marshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: array record", func(t *testing.T) {
		t.Parallel()

		records := []ArrayRecord{
			{Name: "Alice", Phones: []string{"111", "222"}},
			{Name: "Bob", Phones: []string{"333"}},
		}

		var buf bytes.Buffer
		err := csvpp.Marshal(&buf, records)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		want := "name,phone[]\nAlice,111~222\nBob,333\n"
		got := buf.String()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Marshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: structured record", func(t *testing.T) {
		t.Parallel()

		records := []StructuredRecord{
			{Name: "Alice", Geo: GeoLocation{Lat: 34.0522, Lon: -118.2437}},
		}

		var buf bytes.Buffer
		err := csvpp.Marshal(&buf, records)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		want := "name,geo(lat^lon)\nAlice,34.0522^-118.2437\n"
		got := buf.String()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Marshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: array structured record", func(t *testing.T) {
		t.Parallel()

		records := []ArrayStructuredRecord{
			{
				Name: "Alice",
				Addresses: []Address{
					{Type: "home", Street: "123 Main"},
					{Type: "work", Street: "456 Oak"},
				},
			},
		}

		var buf bytes.Buffer
		err := csvpp.Marshal(&buf, records)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		want := "name,address[](type^street)\nAlice,home^123 Main~work^456 Oak\n"
		got := buf.String()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Marshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: empty slice", func(t *testing.T) {
		t.Parallel()

		var records []SimpleRecord

		var buf bytes.Buffer
		err := csvpp.Marshal(&buf, records)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		got := buf.String()
		if got != "" {
			t.Errorf("Marshal() = %q, want empty string", got)
		}
	})

	t.Run("success: pointer to slice", func(t *testing.T) {
		t.Parallel()

		records := []SimpleRecord{
			{Name: "Alice", Age: 30},
		}

		var buf bytes.Buffer
		err := csvpp.Marshal(&buf, &records)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		want := "name,age\nAlice,30\n"
		got := buf.String()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Marshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("error: src is not a slice", func(t *testing.T) {
		t.Parallel()

		record := SimpleRecord{Name: "Alice", Age: 30}

		var buf bytes.Buffer
		err := csvpp.Marshal(&buf, record)
		if err == nil {
			t.Error("Marshal() expected error for non-slice src")
		}
	})

	t.Run("error: slice element is not a struct", func(t *testing.T) {
		t.Parallel()

		records := []string{"Alice", "Bob"}

		var buf bytes.Buffer
		err := csvpp.Marshal(&buf, records)
		if err == nil {
			t.Error("Marshal() expected error for non-struct element")
		}
	})

	t.Run("success: all types record", func(t *testing.T) {
		t.Parallel()

		records := []AllTypesRecord{
			{Name: "Alice", Age: 30, Score: 100, Height: 1.65, Active: true},
			{Name: "Bob", Age: 25, Score: 85, Height: 1.80, Active: false},
		}

		var buf bytes.Buffer
		err := csvpp.Marshal(&buf, records)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		want := "name,age,score,height,active\nAlice,30,100,1.65,true\nBob,25,85,1.8,false\n"
		got := buf.String()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Marshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: pointer struct record", func(t *testing.T) {
		t.Parallel()

		records := []PointerStructRecord{
			{Name: "Alice", Geo: &GeoLocation{Lat: 34.0522, Lon: -118.2437}},
		}

		var buf bytes.Buffer
		err := csvpp.Marshal(&buf, records)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		want := "name,geo(lat^lon)\nAlice,34.0522^-118.2437\n"
		got := buf.String()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Marshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: pointer struct record with nil", func(t *testing.T) {
		t.Parallel()

		records := []PointerStructRecord{
			{Name: "Alice", Geo: nil},
		}

		var buf bytes.Buffer
		err := csvpp.Marshal(&buf, records)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		want := "name,geo(lat^lon)\nAlice,\n"
		got := buf.String()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Marshal() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: nested array of int", func(t *testing.T) {
		t.Parallel()

		records := []NestedArrayRecord{
			{Name: "Alice", Scores: []int{90, 85, 95}},
		}

		var buf bytes.Buffer
		err := csvpp.Marshal(&buf, records)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		want := "name,scores[]\nAlice,90~85~95\n"
		got := buf.String()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Marshal() mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestMarshalUnmarshal_RoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("success: simple record round trip", func(t *testing.T) {
		t.Parallel()

		original := []SimpleRecord{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
		}

		var buf bytes.Buffer
		if err := csvpp.Marshal(&buf, original); err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		var decoded []SimpleRecord
		if err := csvpp.Unmarshal(&buf, &decoded); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if diff := cmp.Diff(original, decoded); diff != "" {
			t.Errorf("round trip mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: array record round trip", func(t *testing.T) {
		t.Parallel()

		original := []ArrayRecord{
			{Name: "Alice", Phones: []string{"111", "222"}},
			{Name: "Bob", Phones: []string{"333"}},
		}

		var buf bytes.Buffer
		if err := csvpp.Marshal(&buf, original); err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		var decoded []ArrayRecord
		if err := csvpp.Unmarshal(&buf, &decoded); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if diff := cmp.Diff(original, decoded); diff != "" {
			t.Errorf("round trip mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: structured record round trip", func(t *testing.T) {
		t.Parallel()

		original := []StructuredRecord{
			{Name: "Alice", Geo: GeoLocation{Lat: 34.0522, Lon: -118.2437}},
		}

		var buf bytes.Buffer
		if err := csvpp.Marshal(&buf, original); err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		var decoded []StructuredRecord
		if err := csvpp.Unmarshal(&buf, &decoded); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if diff := cmp.Diff(original, decoded); diff != "" {
			t.Errorf("round trip mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success: array structured record round trip", func(t *testing.T) {
		t.Parallel()

		original := []ArrayStructuredRecord{
			{
				Name: "Alice",
				Addresses: []Address{
					{Type: "home", Street: "123 Main"},
					{Type: "work", Street: "456 Oak"},
				},
			},
		}

		var buf bytes.Buffer
		if err := csvpp.Marshal(&buf, original); err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		var decoded []ArrayStructuredRecord
		if err := csvpp.Unmarshal(&buf, &decoded); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if diff := cmp.Diff(original, decoded); diff != "" {
			t.Errorf("round trip mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestExtractTagName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "success: simple name",
			input: "name",
			want:  "name",
		},
		{
			name:  "success: array field",
			input: "phone[]",
			want:  "phone",
		},
		{
			name:  "success: array field with delimiter",
			input: "phone[|]",
			want:  "phone",
		},
		{
			name:  "success: structured field",
			input: "geo^(lat^lon)",
			want:  "geo",
		},
		{
			name:  "success: structured field with parenthesis",
			input: "geo(lat^lon)",
			want:  "geo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := csvpp.ExtractTagName(tt.input)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("extractTagName() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
