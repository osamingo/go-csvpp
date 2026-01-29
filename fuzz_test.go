package csvpp_test

import (
	"strings"
	"testing"

	"github.com/osamingo/go-csvpp"
)

// FuzzParseColumnHeader tests that parseColumnHeader does not panic on arbitrary input.
// It verifies robustness against malformed or adversarial header strings.
func FuzzParseColumnHeader(f *testing.F) {
	// Add seed corpus with various valid and invalid header formats
	seeds := []string{
		// Valid simple fields
		"name",
		"field_name",
		"field-name",
		"field123",
		// Valid array fields
		"tags[]",
		"phone[|]",
		"items[~]",
		// Valid structured fields
		"geo(lat^lon)",
		"address(street^city^zip)",
		"data;(a;b;c)",
		// Valid array structured fields
		"address[](street^city)",
		"items[|](name^value)",
		// Nested structures
		"data(outer(inner1^inner2)^simple)",
		"complex;(a(b^c);d[~];e)",
		// Empty and edge cases
		"",
		"a",
		"[]",
		"()",
		"[]()",
		// Potentially problematic input
		"[[[]]]",
		"((()))",
		"[~](^)",
		"name[",
		"name(",
		"name]",
		"name)",
		"name[](",
		// Unicode (should be rejected as header names are ASCII-only)
		"日本語",
		"имя",
		// Special characters
		"@#$%",
		"name@field",
		"field.name",
		// Long input
		strings.Repeat("a", 1000),
		strings.Repeat("[]", 100),
		strings.Repeat("()", 100),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// The function should not panic regardless of input
		// Errors are expected for invalid input
		if _, err := csvpp.ParseColumnHeader(input); err != nil {
			return
		}
	})
}

// FuzzReader tests that Reader does not panic on arbitrary CSV++ input.
func FuzzReader(f *testing.F) {
	// Add seed corpus with various CSV++ formats
	seeds := []string{
		// Simple CSV
		"name,age\nAlice,30\n",
		"a,b,c\n1,2,3\n",
		// Array fields
		"name,tags[]\nAlice,go~rust~python\n",
		"id,values[|]\n1,a|b|c\n",
		// Structured fields
		"name,geo(lat^lon)\nAlice,34.0522^-118.2437\n",
		// Array structured fields
		"name,address[](street^city)\nAlice,123 Main^LA~456 Oak^NY\n",
		// Empty values
		"a,b\n,\n",
		"name,tags[]\n,\n",
		// Edge cases
		"",
		"\n",
		"\n\n\n",
		"name\n",
		"name,age\n",
		// Quoted fields (RFC 4180)
		"name,note\n\"Alice\",\"Hello, World\"\n",
		"a,b\n\"line1\nline2\",value\n",
		// Unicode data (header names are ASCII, values can be Unicode)
		"name,city\n田中太郎,東京\n",
		"name,city\nМария,Москва\n",
		// Malformed input
		"name[,age\nAlice,30\n",
		"name(,age\nAlice,30\n",
		"name,age][\nAlice,30\n",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// The reader should not panic regardless of input
		r := csvpp.NewReader(strings.NewReader(input))

		// Try to read headers
		_, err := r.Headers()
		if err != nil {
			return // Expected for invalid input
		}

		// Try to read all records
		// Errors are expected for invalid input
		if _, err := r.ReadAll(); err != nil {
			return
		}
	})
}

// FuzzSplitByRune tests that splitByRune does not panic on arbitrary input.
func FuzzSplitByRune(f *testing.F) {
	// Add seed corpus
	seeds := []struct {
		input string
		sep   rune
	}{
		{"a~b~c", '~'},
		{"", '~'},
		{"~~~", '~'},
		{"no-separator", '~'},
		{strings.Repeat("a~", 1000), '~'},
		{"日本語~中文~한국어", '~'},
		{"a^b^c", '^'},
		{"a;b;c", ';'},
		{"a:b:c", ':'},
	}

	for _, seed := range seeds {
		f.Add(seed.input, seed.sep)
	}

	f.Fuzz(func(t *testing.T, input string, sep rune) {
		// Should not panic
		result := csvpp.SplitByRune(input, sep)

		// Basic invariants
		if input == "" && len(result) != 0 {
			t.Errorf("splitByRune(%q, %q) returned non-empty for empty input", input, sep)
		}
	})
}

// FuzzNestingDepth tests that deeply nested structures are handled correctly.
func FuzzNestingDepth(f *testing.F) {
	// Generate increasingly nested structures
	f.Add("a(b)")
	f.Add("a(b(c))")
	f.Add("a(b(c(d)))")
	f.Add("a(b(c(d(e))))")
	f.Add("a(b(c(d(e(f)))))")
	f.Add("a(b(c(d(e(f(g))))))")
	f.Add("a(b(c(d(e(f(g(h)))))))")
	f.Add("a(b(c(d(e(f(g(h(i))))))))")
	f.Add("a(b(c(d(e(f(g(h(i(j)))))))))")
	f.Add("a(b(c(d(e(f(g(h(i(j(k))))))))))")
	f.Add("a(b(c(d(e(f(g(h(i(j(k(l)))))))))))")

	f.Fuzz(func(t *testing.T, input string) {
		// Should not panic, should respect nesting limit
		// Errors are expected for invalid input
		if _, err := csvpp.ParseColumnHeader(input); err != nil {
			return
		}
	})
}
