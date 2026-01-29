package csvpp_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/osamingo/go-csvpp"
)

// Benchmark data generators

func generateSimpleCSV(rows int) string {
	var sb strings.Builder
	sb.WriteString("id,name,age\n")
	for range rows {
		sb.WriteString("1,Alice,30\n")
	}
	return sb.String()
}

func generateArrayCSV(rows int) string {
	var sb strings.Builder
	sb.WriteString("id,name,tags[]\n")
	for range rows {
		sb.WriteString("1,Alice,go~rust~python~java~typescript\n")
	}
	return sb.String()
}

func generateStructuredCSV(rows int) string {
	var sb strings.Builder
	sb.WriteString("id,name,geo(lat^lon)\n")
	for range rows {
		sb.WriteString("1,Alice,34.0522^-118.2437\n")
	}
	return sb.String()
}

func generateArrayStructuredCSV(rows int) string {
	var sb strings.Builder
	sb.WriteString("id,name,address[](street^city^state^zip)\n")
	for range rows {
		sb.WriteString("1,Alice,123 Main St^Los Angeles^CA^90210~456 Oak Ave^New York^NY^10001\n")
	}
	return sb.String()
}

// Reader Benchmarks

func BenchmarkReader_Read_Simple(b *testing.B) {
	input := generateSimpleCSV(100)

	for b.Loop() {
		r := csvpp.NewReader(strings.NewReader(input))
		for {
			_, err := r.Read()
			if err != nil {
				break
			}
		}
	}
}

func BenchmarkReader_Read_Array(b *testing.B) {
	input := generateArrayCSV(100)

	for b.Loop() {
		r := csvpp.NewReader(strings.NewReader(input))
		for {
			_, err := r.Read()
			if err != nil {
				break
			}
		}
	}
}

func BenchmarkReader_Read_Structured(b *testing.B) {
	input := generateStructuredCSV(100)

	for b.Loop() {
		r := csvpp.NewReader(strings.NewReader(input))
		for {
			_, err := r.Read()
			if err != nil {
				break
			}
		}
	}
}

func BenchmarkReader_Read_ArrayStructured(b *testing.B) {
	input := generateArrayStructuredCSV(100)

	for b.Loop() {
		r := csvpp.NewReader(strings.NewReader(input))
		for {
			_, err := r.Read()
			if err != nil {
				break
			}
		}
	}
}

func BenchmarkReader_ReadAll_Simple(b *testing.B) {
	input := generateSimpleCSV(1000)

	for b.Loop() {
		r := csvpp.NewReader(strings.NewReader(input))
		if _, err := r.ReadAll(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReader_ReadAll_ArrayStructured(b *testing.B) {
	input := generateArrayStructuredCSV(1000)

	for b.Loop() {
		r := csvpp.NewReader(strings.NewReader(input))
		if _, err := r.ReadAll(); err != nil {
			b.Fatal(err)
		}
	}
}

// Writer Benchmarks

func BenchmarkWriter_Write_Simple(b *testing.B) {
	headers := []*csvpp.ColumnHeader{
		{Name: "id", Kind: csvpp.SimpleField},
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "age", Kind: csvpp.SimpleField},
	}
	record := []*csvpp.Field{
		{Value: "1"},
		{Value: "Alice"},
		{Value: "30"},
	}

	for b.Loop() {
		var buf bytes.Buffer
		w := csvpp.NewWriter(&buf)
		w.SetHeaders(headers)
		if err := w.WriteHeader(); err != nil {
			b.Fatal(err)
		}
		for range 100 {
			if err := w.Write(record); err != nil {
				b.Fatal(err)
			}
		}
		w.Flush()
	}
}

func BenchmarkWriter_Write_Array(b *testing.B) {
	headers := []*csvpp.ColumnHeader{
		{Name: "id", Kind: csvpp.SimpleField},
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "tags", Kind: csvpp.ArrayField, ArrayDelimiter: '~'},
	}
	record := []*csvpp.Field{
		{Value: "1"},
		{Value: "Alice"},
		{Values: []string{"go", "rust", "python", "java", "typescript"}},
	}

	for b.Loop() {
		var buf bytes.Buffer
		w := csvpp.NewWriter(&buf)
		w.SetHeaders(headers)
		if err := w.WriteHeader(); err != nil {
			b.Fatal(err)
		}
		for range 100 {
			if err := w.Write(record); err != nil {
				b.Fatal(err)
			}
		}
		w.Flush()
	}
}

func BenchmarkWriter_Write_ArrayStructured(b *testing.B) {
	headers := []*csvpp.ColumnHeader{
		{Name: "id", Kind: csvpp.SimpleField},
		{Name: "name", Kind: csvpp.SimpleField},
		{
			Name:               "address",
			Kind:               csvpp.ArrayStructuredField,
			ArrayDelimiter:     '~',
			ComponentDelimiter: '^',
			Components: []*csvpp.ColumnHeader{
				{Name: "street", Kind: csvpp.SimpleField},
				{Name: "city", Kind: csvpp.SimpleField},
				{Name: "state", Kind: csvpp.SimpleField},
				{Name: "zip", Kind: csvpp.SimpleField},
			},
		},
	}
	record := []*csvpp.Field{
		{Value: "1"},
		{Value: "Alice"},
		{Components: []*csvpp.Field{
			{Components: []*csvpp.Field{{Value: "123 Main St"}, {Value: "Los Angeles"}, {Value: "CA"}, {Value: "90210"}}},
			{Components: []*csvpp.Field{{Value: "456 Oak Ave"}, {Value: "New York"}, {Value: "NY"}, {Value: "10001"}}},
		}},
	}

	for b.Loop() {
		var buf bytes.Buffer
		w := csvpp.NewWriter(&buf)
		w.SetHeaders(headers)
		if err := w.WriteHeader(); err != nil {
			b.Fatal(err)
		}
		for range 100 {
			if err := w.Write(record); err != nil {
				b.Fatal(err)
			}
		}
		w.Flush()
	}
}

func BenchmarkWriter_WriteAll(b *testing.B) {
	headers := []*csvpp.ColumnHeader{
		{Name: "id", Kind: csvpp.SimpleField},
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "age", Kind: csvpp.SimpleField},
	}
	records := make([][]*csvpp.Field, 1000)
	for i := range records {
		records[i] = []*csvpp.Field{
			{Value: "1"},
			{Value: "Alice"},
			{Value: "30"},
		}
	}

	for b.Loop() {
		var buf bytes.Buffer
		w := csvpp.NewWriter(&buf)
		w.SetHeaders(headers)
		if err := w.WriteAll(records); err != nil {
			b.Fatal(err)
		}
	}
}

// Header Parsing Benchmarks

func BenchmarkParseColumnHeader_Simple(b *testing.B) {
	for b.Loop() {
		if _, err := csvpp.ParseColumnHeader("name"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseColumnHeader_Array(b *testing.B) {
	for b.Loop() {
		if _, err := csvpp.ParseColumnHeader("tags[]"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseColumnHeader_Structured(b *testing.B) {
	for b.Loop() {
		if _, err := csvpp.ParseColumnHeader("geo(lat^lon)"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseColumnHeader_ArrayStructured(b *testing.B) {
	for b.Loop() {
		if _, err := csvpp.ParseColumnHeader("address[](street^city^state^zip)"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseColumnHeader_Nested(b *testing.B) {
	for b.Loop() {
		if _, err := csvpp.ParseColumnHeader("data;(outer(inner1^inner2);simple)"); err != nil {
			b.Fatal(err)
		}
	}
}

// Marshal/Unmarshal Benchmarks

type BenchmarkPerson struct {
	Name   string   `csvpp:"name"`
	Age    int      `csvpp:"age"`
	Phones []string `csvpp:"phone[]"`
}

func BenchmarkUnmarshal(b *testing.B) {
	input := `name,age,phone[]
Alice,30,555-1234~555-5678
Bob,25,555-9999
Charlie,35,555-1111~555-2222~555-3333
`
	inputBytes := []byte(input)

	for b.Loop() {
		var people []BenchmarkPerson
		if err := csvpp.Unmarshal(bytes.NewReader(inputBytes), &people); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal(b *testing.B) {
	people := []BenchmarkPerson{
		{Name: "Alice", Age: 30, Phones: []string{"555-1234", "555-5678"}},
		{Name: "Bob", Age: 25, Phones: []string{"555-9999"}},
		{Name: "Charlie", Age: 35, Phones: []string{"555-1111", "555-2222", "555-3333"}},
	}

	for b.Loop() {
		var buf bytes.Buffer
		if err := csvpp.Marshal(&buf, people); err != nil {
			b.Fatal(err)
		}
	}
}

// splitByRune Benchmark

func BenchmarkSplitByRune(b *testing.B) {
	input := "a~b~c~d~e~f~g~h~i~j"

	for b.Loop() {
		_ = csvpp.SplitByRune(input, '~')
	}
}

func BenchmarkSplitByRune_Long(b *testing.B) {
	var sb strings.Builder
	for i := range 100 {
		if i > 0 {
			sb.WriteRune('~')
		}
		sb.WriteString("value")
	}
	input := sb.String()

	for b.Loop() {
		_ = csvpp.SplitByRune(input, '~')
	}
}
