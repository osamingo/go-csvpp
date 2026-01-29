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

func BenchmarkRecordToMap(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = csvpputil.RecordToMap(benchHeaders, benchRecord)
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

func BenchmarkJSONEncoder_Encode(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		enc := csvpputil.NewJSONEncoder(io.Discard, benchHeaders)
		for _, record := range benchRecords {
			if err := enc.Encode(record); err != nil {
				b.Fatal(err)
			}
		}
		if err := enc.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYAMLEncoder_Encode(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		enc := csvpputil.NewYAMLEncoder(io.Discard, benchHeaders)
		for _, record := range benchRecords {
			if err := enc.Encode(record); err != nil {
				b.Fatal(err)
			}
		}
		if err := enc.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONEncoder_SingleRecord(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var buf bytes.Buffer
		enc := csvpputil.NewJSONEncoder(&buf, benchHeaders)
		if err := enc.Encode(benchRecord); err != nil {
			b.Fatal(err)
		}
		if err := enc.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYAMLEncoder_SingleRecord(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var buf bytes.Buffer
		enc := csvpputil.NewYAMLEncoder(&buf, benchHeaders)
		if err := enc.Encode(benchRecord); err != nil {
			b.Fatal(err)
		}
		if err := enc.Close(); err != nil {
			b.Fatal(err)
		}
	}
}
