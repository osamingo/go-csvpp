package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConvertCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "success: csvpp to json",
			args:       []string{"convert", "-i", "testdata/convert/simple.csvpp", "--to", "json"},
			wantErr:    false,
			wantOutput: "[{\"name\":\"Alice\",\"age\":\"30\"},{\"name\":\"Bob\",\"age\":\"25\"}]\n",
		},
		{
			name:    "success: csvpp to yaml",
			args:    []string{"convert", "-i", "testdata/convert/simple.csvpp", "--to", "yaml"},
			wantErr: false,
			wantOutput: `- name: Alice
  age: "30"
- name: Bob
  age: "25"
`,
		},
		{
			name:    "error: missing output format",
			args:    []string{"convert", "-i", "testdata/convert/simple.csvpp"},
			wantErr: true,
		},
		{
			name:    "error: file not found",
			args:    []string{"convert", "-i", "nonexistent.csvpp", "--to", "json"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stdout, _, err := runCommand(t, tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if diff := cmp.Diff(tt.wantOutput, stdout); diff != "" {
				t.Errorf("output mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConvertRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		inputFile    string
		format       string
		fromFormat   string
		wantContains []string
	}{
		{
			name:         "success: json roundtrip",
			inputFile:    "testdata/convert/simple.csvpp",
			format:       "json",
			fromFormat:   "json",
			wantContains: []string{"name", "age", "Alice", "Bob"},
		},
		{
			name:         "success: yaml roundtrip",
			inputFile:    "testdata/convert/simple.csvpp",
			format:       "yaml",
			fromFormat:   "yaml",
			wantContains: []string{"name", "age", "Alice", "Bob"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Step 1: Convert CSVPP to JSON/YAML
			tmpDir := t.TempDir()
			intermediateFile := filepath.Join(tmpDir, "intermediate."+tt.format)

			_, _, err := runCommand(t, "convert", "-i", tt.inputFile, "-o", intermediateFile)
			if err != nil {
				t.Fatalf("step 1 (to %s) failed: %v", tt.format, err)
			}

			// Step 2: Convert back to CSVPP
			outputFile := filepath.Join(tmpDir, "output.csvpp")
			_, _, err = runCommand(t, "convert", "-i", intermediateFile, "-o", outputFile)
			if err != nil {
				t.Fatalf("step 2 (to csvpp) failed: %v", err)
			}

			// Read output and verify
			content, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("failed to read output: %v", err)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(string(content), want) {
					t.Errorf("output missing %q:\n%s", want, content)
				}
			}
		})
	}
}

func TestConvertWithFromFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		inputFile  string
		fromFormat string
		toFormat   string
		wantErr    bool
	}{
		{
			name:       "success: json to csvpp with from flag",
			inputFile:  "testdata/convert/simple.json",
			fromFormat: "json",
			toFormat:   "csvpp",
			wantErr:    false,
		},
		{
			name:       "success: yaml to csvpp with from flag",
			inputFile:  "testdata/convert/simple.yaml",
			fromFormat: "yaml",
			toFormat:   "csvpp",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			outputFile := filepath.Join(tmpDir, "output.csvpp")

			_, _, err := runCommand(t, "convert", "-i", tt.inputFile, "--from", tt.fromFormat, "-o", outputFile)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify output file exists and has content
			info, err := os.Stat(outputFile)
			if err != nil {
				t.Errorf("output file not created: %v", err)
				return
			}
			if info.Size() == 0 {
				t.Error("output file is empty")
			}
		})
	}
}
