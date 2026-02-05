package main_test

import (
	"strings"
	"testing"
)

func TestViewCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		wantContains []string
	}{
		{
			name:         "success: view simple file (piped)",
			args:         []string{"view", "testdata/validate/valid.csvpp"},
			wantErr:      false,
			wantContains: []string{"name", "age", "city", "Alice", "Bob", "Tokyo", "Osaka"},
		},
		{
			name:    "error: file not found",
			args:    []string{"view", "nonexistent.csvpp"},
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

			for _, want := range tt.wantContains {
				if !strings.Contains(stdout, want) {
					t.Errorf("output missing %q:\n%s", want, stdout)
				}
			}
		})
	}
}
