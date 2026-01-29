package csvpputil

import (
	"bytes"
	"io"
	"testing"

	"github.com/osamingo/go-csvpp"
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
		_ = RecordToMap(benchHeaders, benchRecord)
	}
}

func BenchmarkMarshalJSON(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = MarshalJSON(benchHeaders, benchRecords)
	}
}

func BenchmarkMarshalYAML(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = MarshalYAML(benchHeaders, benchRecords)
	}
}

func BenchmarkJSONEncoder_Encode(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		enc := NewJSONEncoder(io.Discard, benchHeaders)
		for _, record := range benchRecords {
			_ = enc.Encode(record)
		}
		_ = enc.Close()
	}
}

func BenchmarkYAMLEncoder_Encode(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		enc := NewYAMLEncoder(io.Discard, benchHeaders)
		for _, record := range benchRecords {
			_ = enc.Encode(record)
		}
		_ = enc.Close()
	}
}

func BenchmarkJSONEncoder_SingleRecord(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var buf bytes.Buffer
		enc := NewJSONEncoder(&buf, benchHeaders)
		_ = enc.Encode(benchRecord)
		_ = enc.Close()
	}
}

func BenchmarkYAMLEncoder_SingleRecord(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var buf bytes.Buffer
		enc := NewYAMLEncoder(&buf, benchHeaders)
		_ = enc.Encode(benchRecord)
		_ = enc.Close()
	}
}
