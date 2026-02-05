package csvpputil_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/osamingo/go-csvpp"
	"github.com/osamingo/go-csvpp/csvpputil"
)

func TestGolden_ConvertJSON(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("testdata/convert/*.csvpp")
	if err != nil {
		t.Fatalf("failed to glob convert tests: %v", err)
	}

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".csvpp")

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runGoldenConvertJSONTest(t, file)
		})
	}
}

func TestGolden_ConvertYAML(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("testdata/convert/*.csvpp")
	if err != nil {
		t.Fatalf("failed to glob convert tests: %v", err)
	}

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".csvpp")

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runGoldenConvertYAMLTest(t, file)
		})
	}
}

func runGoldenConvertJSONTest(t *testing.T, csvppFile string) {
	t.Helper()

	// Read .csvpp file
	input, err := os.Open(csvppFile)
	if err != nil {
		t.Fatalf("failed to open csvpp file: %v", err)
	}
	defer input.Close()

	// Parse CSV++
	r := csvpp.NewReader(input)
	headers, err := r.Headers()
	if err != nil {
		t.Fatalf("failed to parse headers: %v", err)
	}

	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to read records: %v", err)
	}

	// Convert to JSON
	got, err := csvpputil.MarshalJSON(headers, records)
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	// Read expected JSON
	jsonFile := strings.TrimSuffix(csvppFile, ".csvpp") + ".json"
	want, err := os.ReadFile(jsonFile)
	if err != nil {
		t.Fatalf("failed to read expected JSON: %v", err)
	}

	// Compare (trim trailing newline for comparison)
	gotJSON := strings.TrimSuffix(string(got), "\n")
	wantJSON := strings.TrimSuffix(string(want), "\n")

	if diff := cmp.Diff(wantJSON, gotJSON); diff != "" {
		t.Errorf("JSON mismatch (-want +got):\n%s", diff)
	}
}

func runGoldenConvertYAMLTest(t *testing.T, csvppFile string) {
	t.Helper()

	// Read .csvpp file
	input, err := os.Open(csvppFile)
	if err != nil {
		t.Fatalf("failed to open csvpp file: %v", err)
	}
	defer input.Close()

	// Parse CSV++
	r := csvpp.NewReader(input)
	headers, err := r.Headers()
	if err != nil {
		t.Fatalf("failed to parse headers: %v", err)
	}

	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to read records: %v", err)
	}

	// Convert to YAML
	got, err := csvpputil.MarshalYAML(headers, records)
	if err != nil {
		t.Fatalf("MarshalYAML() error: %v", err)
	}

	// Read expected YAML
	yamlFile := strings.TrimSuffix(csvppFile, ".csvpp") + ".yaml"
	want, err := os.ReadFile(yamlFile)
	if err != nil {
		t.Fatalf("failed to read expected YAML: %v", err)
	}

	if diff := cmp.Diff(string(want), string(got)); diff != "" {
		t.Errorf("YAML mismatch (-want +got):\n%s", diff)
	}
}
