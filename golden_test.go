package csvpp_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/osamingo/go-csvpp"
)

// goldenReadTest represents a read test case from JSON.
type goldenReadTest struct {
	Description   string          `json:"description"`
	Input         string          `json:"input"`
	Headers       []goldenHeader  `json:"headers"`
	Records       [][]goldenField `json:"records"`
	ExpectError   bool            `json:"expectError"`
	ErrorContains string          `json:"errorContains"`
}

// goldenHeader represents an expected header.
type goldenHeader struct {
	Name               string         `json:"name"`
	Kind               string         `json:"kind"`
	ArrayDelimiter     string         `json:"arrayDelimiter,omitempty"`
	ComponentDelimiter string         `json:"componentDelimiter,omitempty"`
	Components         []goldenHeader `json:"components,omitempty"`
}

// goldenField represents an expected field value.
type goldenField struct {
	Value      string        `json:"value,omitempty"`
	Values     []string      `json:"values,omitempty"`
	Components []goldenField `json:"components,omitempty"`
}

func TestGolden_Read(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("testdata/read/*.json")
	if err != nil {
		t.Fatalf("failed to glob read tests: %v", err)
	}

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".json")

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runGoldenReadTest(t, file)
		})
	}
}

func TestGolden_Errors(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("testdata/errors/*.json")
	if err != nil {
		t.Fatalf("failed to glob error tests: %v", err)
	}

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".json")

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runGoldenErrorTest(t, file)
		})
	}
}

func TestGolden_Roundtrip(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("testdata/roundtrip/*.csvpp")
	if err != nil {
		t.Fatalf("failed to glob roundtrip tests: %v", err)
	}

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".csvpp")

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runGoldenRoundtripTest(t, file)
		})
	}
}

func runGoldenReadTest(t *testing.T, file string) {
	t.Helper()

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	var tc goldenReadTest
	if err := json.Unmarshal(data, &tc); err != nil {
		t.Fatalf("failed to parse test file: %v", err)
	}

	r := csvpp.NewReader(strings.NewReader(tc.Input))

	// Parse headers
	headers, err := r.Headers()
	if err != nil {
		t.Fatalf("failed to parse headers: %v", err)
	}

	// Verify headers
	if len(headers) != len(tc.Headers) {
		t.Errorf("header count mismatch: got %d, want %d", len(headers), len(tc.Headers))
	} else {
		for i, h := range headers {
			verifyHeader(t, i, h, tc.Headers[i])
		}
	}

	// Parse records
	var records [][]*csvpp.Field
	for {
		record, err := r.Read()
		if err != nil {
			if err.Error() == "EOF" || strings.Contains(err.Error(), "EOF") {
				break
			}
			t.Fatalf("failed to read record: %v", err)
		}
		records = append(records, record)
	}

	// Verify records
	if len(records) != len(tc.Records) {
		t.Errorf("record count mismatch: got %d, want %d", len(records), len(tc.Records))
	} else {
		for i, record := range records {
			verifyRecord(t, i, record, tc.Records[i])
		}
	}
}

func runGoldenErrorTest(t *testing.T, file string) {
	t.Helper()

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	var tc goldenReadTest
	if err := json.Unmarshal(data, &tc); err != nil {
		t.Fatalf("failed to parse test file: %v", err)
	}

	r := csvpp.NewReader(strings.NewReader(tc.Input))

	_, err = r.Headers()
	if err == nil {
		// Try reading records
		_, err = r.ReadAll()
	}

	if !tc.ExpectError {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		return
	}

	if err == nil {
		t.Errorf("expected error containing %q, got nil", tc.ErrorContains)
		return
	}

	if tc.ErrorContains != "" && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.ErrorContains)) {
		t.Errorf("error %q does not contain %q", err.Error(), tc.ErrorContains)
	}
}

func runGoldenRoundtripTest(t *testing.T, file string) {
	t.Helper()

	input, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	// Read
	r := csvpp.NewReader(bytes.NewReader(input))
	headers, err := r.Headers()
	if err != nil {
		t.Fatalf("failed to parse headers: %v", err)
	}

	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to read records: %v", err)
	}

	// Write
	var buf bytes.Buffer
	w := csvpp.NewWriter(&buf)
	w.SetHeaders(headers)

	if err := w.WriteHeader(); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}

	for _, record := range records {
		if err := w.Write(record); err != nil {
			t.Fatalf("failed to write record: %v", err)
		}
	}
	w.Flush()

	// Compare
	got := buf.String()
	want := string(input)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("roundtrip mismatch (-want +got):\n%s", diff)
	}
}

func verifyHeader(t *testing.T, index int, got *csvpp.ColumnHeader, want goldenHeader) {
	t.Helper()

	if got.Name != want.Name {
		t.Errorf("header[%d].Name = %q, want %q", index, got.Name, want.Name)
	}

	wantKind := parseKind(want.Kind)
	if got.Kind != wantKind {
		t.Errorf("header[%d].Kind = %v, want %v", index, got.Kind, wantKind)
	}

	if want.ArrayDelimiter != "" {
		wantDelim := []rune(want.ArrayDelimiter)[0]
		if got.ArrayDelimiter != wantDelim {
			t.Errorf("header[%d].ArrayDelimiter = %q, want %q", index, got.ArrayDelimiter, wantDelim)
		}
	}

	if want.ComponentDelimiter != "" {
		wantDelim := []rune(want.ComponentDelimiter)[0]
		if got.ComponentDelimiter != wantDelim {
			t.Errorf("header[%d].ComponentDelimiter = %q, want %q", index, got.ComponentDelimiter, wantDelim)
		}
	}

	if len(want.Components) > 0 {
		if len(got.Components) != len(want.Components) {
			t.Errorf("header[%d].Components count = %d, want %d", index, len(got.Components), len(want.Components))
		} else {
			for i, comp := range got.Components {
				verifyHeader(t, i, comp, want.Components[i])
			}
		}
	}
}

func verifyRecord(t *testing.T, recordIndex int, got []*csvpp.Field, want []goldenField) {
	t.Helper()

	if len(got) != len(want) {
		t.Errorf("record[%d] field count = %d, want %d", recordIndex, len(got), len(want))
		return
	}

	for i, field := range got {
		verifyField(t, recordIndex, i, field, want[i])
	}
}

func verifyField(t *testing.T, recordIndex, fieldIndex int, got *csvpp.Field, want goldenField) {
	t.Helper()

	prefix := func() string {
		return "record[" + string(rune('0'+recordIndex)) + "].field[" + string(rune('0'+fieldIndex)) + "]"
	}

	if want.Value != "" || (len(want.Values) == 0 && len(want.Components) == 0) {
		if got.Value != want.Value {
			t.Errorf("%s.Value = %q, want %q", prefix(), got.Value, want.Value)
		}
	}

	if len(want.Values) > 0 {
		if diff := cmp.Diff(want.Values, got.Values); diff != "" {
			t.Errorf("%s.Values mismatch (-want +got):\n%s", prefix(), diff)
		}
	}

	if len(want.Components) > 0 {
		if len(got.Components) != len(want.Components) {
			t.Errorf("%s.Components count = %d, want %d", prefix(), len(got.Components), len(want.Components))
		} else {
			for i, comp := range got.Components {
				verifyField(t, recordIndex, i, comp, want.Components[i])
			}
		}
	}
}

func parseKind(s string) csvpp.FieldKind {
	switch s {
	case "simple":
		return csvpp.SimpleField
	case "array":
		return csvpp.ArrayField
	case "structured":
		return csvpp.StructuredField
	case "arrayStructured":
		return csvpp.ArrayStructuredField
	default:
		return csvpp.SimpleField
	}
}
