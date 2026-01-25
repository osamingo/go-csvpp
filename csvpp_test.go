package csvpp

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFieldKind_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind FieldKind
		want string
	}{
		{
			name: "success: SimpleField",
			kind: SimpleField,
			want: "SimpleField",
		},
		{
			name: "success: ArrayField",
			kind: ArrayField,
			want: "ArrayField",
		},
		{
			name: "success: StructuredField",
			kind: StructuredField,
			want: "StructuredField",
		},
		{
			name: "success: ArrayStructuredField",
			kind: ArrayStructuredField,
			want: "ArrayStructuredField",
		},
		{
			name: "success: unknown FieldKind",
			kind: FieldKind(999),
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
		err  *ParseError
		want string
	}{
		{
			name: "success: line only",
			err: &ParseError{
				Line: 1,
				Err:  errors.New("test error"),
			},
			want: "csvpp: line 1: test error",
		},
		{
			name: "success: line and column",
			err: &ParseError{
				Line:   2,
				Column: 3,
				Err:    errors.New("test error"),
			},
			want: "csvpp: line 2, column 3: test error",
		},
		{
			name: "success: line, column, and field",
			err: &ParseError{
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
	parseErr := &ParseError{
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
