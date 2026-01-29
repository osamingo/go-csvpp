package csvpp_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/osamingo/go-csvpp"
)

func TestFieldKind_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind csvpp.FieldKind
		want string
	}{
		{
			name: "success: SimpleField",
			kind: csvpp.SimpleField,
			want: "SimpleField",
		},
		{
			name: "success: ArrayField",
			kind: csvpp.ArrayField,
			want: "ArrayField",
		},
		{
			name: "success: StructuredField",
			kind: csvpp.StructuredField,
			want: "StructuredField",
		},
		{
			name: "success: ArrayStructuredField",
			kind: csvpp.ArrayStructuredField,
			want: "ArrayStructuredField",
		},
		{
			name: "success: unknown FieldKind",
			kind: csvpp.FieldKind(999),
			want: "FieldKind(999)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.kind.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("FieldKind.String() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  *csvpp.ParseError
		want string
	}{
		{
			name: "success: line only",
			err: &csvpp.ParseError{
				Line: 1,
				Err:  errors.New("test error"),
			},
			want: "csvpp: line 1: test error",
		},
		{
			name: "success: line and column",
			err: &csvpp.ParseError{
				Line:   2,
				Column: 3,
				Err:    errors.New("test error"),
			},
			want: "csvpp: line 2, column 3: test error",
		},
		{
			name: "success: line, column, and field",
			err: &csvpp.ParseError{
				Line:   4,
				Column: 5,
				Field:  "name",
				Err:    errors.New("test error"),
			},
			want: `csvpp: line 4, column 5 (field "name"): test error`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.err.Error()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseError.Error() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseError_Unwrap(t *testing.T) {
	t.Parallel()

	originalErr := errors.New("original error")
	parseErr := &csvpp.ParseError{
		Line: 1,
		Err:  originalErr,
	}

	t.Run("success: unwrap returns original error", func(t *testing.T) {
		t.Parallel()

		got := parseErr.Unwrap()
		if got != originalErr {
			t.Errorf("ParseError.Unwrap() = %v, want %v", got, originalErr)
		}
	})

	t.Run("success: errors.Is works with wrapped error", func(t *testing.T) {
		t.Parallel()

		if !errors.Is(parseErr, originalErr) {
			t.Errorf("errors.Is(parseErr, originalErr) = false, want true")
		}
	})
}

func TestHasFormulaPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "success: empty string",
			input: "",
			want:  false,
		},
		{
			name:  "success: normal string",
			input: "hello",
			want:  false,
		},
		{
			name:  "success: starts with equals",
			input: "=SUM(A1:A10)",
			want:  true,
		},
		{
			name:  "success: starts with plus",
			input: "+1234",
			want:  true,
		},
		{
			name:  "success: starts with minus",
			input: "-1234",
			want:  true,
		},
		{
			name:  "success: starts with at",
			input: "@import",
			want:  true,
		},
		{
			name:  "success: contains but not starts with formula char",
			input: "a=b",
			want:  false,
		},
		{
			name:  "success: number string",
			input: "12345",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := csvpp.HasFormulaPrefix(tt.input)
			if got != tt.want {
				t.Errorf("HasFormulaPrefix(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
