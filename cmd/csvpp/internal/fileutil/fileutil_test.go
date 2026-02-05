package fileutil_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/osamingo/go-csvpp/cmd/csvpp/internal/fileutil"
)

func TestOpenInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		setup    func(t *testing.T) string
		wantErr  bool
	}{
		{
			name:     "success: empty filename returns stdin wrapper",
			filename: "",
			wantErr:  false,
		},
		{
			name:     "success: existing file",
			filename: "", // Will be set by setup
			setup: func(t *testing.T) string {
				t.Helper()
				f, err := os.CreateTemp(t.TempDir(), "test-*.txt")
				if err != nil {
					t.Fatal(err)
				}
				f.WriteString("test content")
				f.Close()
				return f.Name()
			},
			wantErr: false,
		},
		{
			name:     "error: nonexistent file",
			filename: "/nonexistent/path/file.txt",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			filename := tt.filename
			if tt.setup != nil {
				filename = tt.setup(t)
			}

			rc, err := fileutil.OpenInput(filename)

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

			if rc == nil {
				t.Error("expected non-nil ReadCloser")
				return
			}

			if err := rc.Close(); err != nil {
				t.Errorf("Close() error: %v", err)
			}
		})
	}
}

func TestOpenInputFromArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		setup   func(t *testing.T) string
		wantErr bool
	}{
		{
			name:    "success: empty args returns stdin wrapper",
			args:    []string{},
			wantErr: false,
		},
		{
			name: "success: file from args",
			args: nil, // Will be set by setup
			setup: func(t *testing.T) string {
				t.Helper()
				f, err := os.CreateTemp(t.TempDir(), "test-*.txt")
				if err != nil {
					t.Fatal(err)
				}
				f.Close()
				return f.Name()
			},
			wantErr: false,
		},
		{
			name:    "error: nonexistent file",
			args:    []string{"/nonexistent/file.txt"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			args := tt.args
			if tt.setup != nil {
				args = []string{tt.setup(t)}
			}

			rc, err := fileutil.OpenInputFromArgs(args)

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

			if rc == nil {
				t.Error("expected non-nil ReadCloser")
				return
			}

			if err := rc.Close(); err != nil {
				t.Errorf("Close() error: %v", err)
			}
		})
	}
}

func TestOpenOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "success: empty filename returns fallback wrapper",
			filename: "",
			wantErr:  false,
		},
		{
			name:     "success: new file",
			filename: "", // Will be set in test
			wantErr:  false,
		},
		{
			name:     "error: invalid path",
			filename: "/nonexistent/dir/file.txt",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			filename := tt.filename
			if tt.name == "success: new file" {
				filename = filepath.Join(t.TempDir(), "output.txt")
			}

			var fallback bytes.Buffer
			wc, err := fileutil.OpenOutput(filename, &fallback)

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

			if wc == nil {
				t.Error("expected non-nil WriteCloser")
				return
			}

			// Write something
			_, err = wc.Write([]byte("test"))
			if err != nil {
				t.Errorf("Write() error: %v", err)
			}

			if err := wc.Close(); err != nil {
				t.Errorf("Close() error: %v", err)
			}
		})
	}
}
