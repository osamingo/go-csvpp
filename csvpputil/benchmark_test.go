package csvpputil_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/osamingo/go-csvpp"
	"github.com/osamingo/go-csvpp/csvpputil"
)

var (
	benchHeaders = []*csvpp.ColumnHeader{
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
	}

	benchRecord = []*csvpp.Field{
		{Value: "Alice"},
		{Values: []string{"go", "rust", "python"}},
		{
			Components: []*csvpp.Field{
				{Value: "35.6762"},
				{Value: "139.6503"},
			},
		},
	}

	benchRecords = func() [][]*csvpp.Field {
		records := make([][]*csvpp.Field, 1000)
		for i := range records {
			records[i] = benchRecord
		}
		return records
	}()
)

func BenchmarkFieldsToMap(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = csvpputil.FieldsToMap(benchHeaders, benchRecord)
	}
}

func BenchmarkMarshalJSON(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if _, err := csvpputil.MarshalJSON(benchHeaders, benchRecords); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalYAML(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if _, err := csvpputil.MarshalYAML(benchHeaders, benchRecords); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONArrayWriter_Write(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		w := csvpputil.NewJSONArrayWriter(io.Discard, benchHeaders)
		for _, record := range benchRecords {
			if err := w.Write(record); err != nil {
				b.Fatal(err)
			}
		}
		if err := w.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYAMLArrayWriter_Write(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		w := csvpputil.NewYAMLArrayWriter(io.Discard, benchHeaders)
		for _, record := range benchRecords {
			if err := w.Write(record); err != nil {
				b.Fatal(err)
			}
		}
		if err := w.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYAMLArrayWriter_WriteWithCapacity(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		w := csvpputil.NewYAMLArrayWriter(io.Discard, benchHeaders, csvpputil.WithYAMLCapacity(len(benchRecords)))
		for _, record := range benchRecords {
			if err := w.Write(record); err != nil {
				b.Fatal(err)
			}
		}
		if err := w.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONArrayWriter_SingleRecord(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var buf bytes.Buffer
		w := csvpputil.NewJSONArrayWriter(&buf, benchHeaders)
		if err := w.Write(benchRecord); err != nil {
			b.Fatal(err)
		}
		if err := w.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYAMLArrayWriter_SingleRecord(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var buf bytes.Buffer
		w := csvpputil.NewYAMLArrayWriter(&buf, benchHeaders)
		if err := w.Write(benchRecord); err != nil {
			b.Fatal(err)
		}
		if err := w.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteJSON(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if err := csvpputil.WriteJSON(io.Discard, benchHeaders, benchRecords); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteYAML(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if err := csvpputil.WriteYAML(io.Discard, benchHeaders, benchRecords); err != nil {
			b.Fatal(err)
		}
	}
}
