package main_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// testBinary is the path to the compiled test binary.
var testBinary string

func TestMain(m *testing.M) {
	// Build the binary for testing
	tmpDir, err := os.MkdirTemp("", "csvpp-test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binaryName := "csvpp"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	testBinary = filepath.Join(tmpDir, binaryName)

	// Build with GOEXPERIMENT=jsonv2
	cmd := exec.Command("go", "build", "-o", testBinary, ".")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(), "GOEXPERIMENT=jsonv2")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build: %s\n", out)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func runCommand(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	cmd := exec.Command(testBinary, args...)
	cmd.Dir = "."

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestValidateCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "success: valid file",
			args:       []string{"validate", "testdata/validate/valid.csvpp"},
			wantErr:    false,
			wantOutput: "Valid CSV++ file with 2 record(s)\n",
		},
		{
			name:    "error: invalid header",
			args:    []string{"validate", "testdata/validate/invalid_header.csvpp"},
			wantErr: true,
		},
		{
			name:    "error: file not found",
			args:    []string{"validate", "testdata/validate/nonexistent.csvpp"},
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

			if stdout != tt.wantOutput {
				t.Errorf("output mismatch:\nwant: %q\ngot:  %q", tt.wantOutput, stdout)
			}
		})
	}
}
