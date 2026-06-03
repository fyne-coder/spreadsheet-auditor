package formula

import "testing"

func TestNormalizeFormulaMatchesPythonPatterns(t *testing.T) {
	tests := []struct {
		formula  string
		anchor   string
		expected string
	}{
		{"=A4*B4+100", "C4", "R0C-2*R0C-1+100"},
		{"=A2*B2", "C2", "R0C-2*R0C-1"},
		{"=A1*1.05", "B1", "R0C-1*1.05"},
	}

	for _, test := range tests {
		pattern, ok := NormalizeFormula(test.formula, test.anchor)
		if !ok {
			t.Fatalf("normalize failed for %s at %s", test.formula, test.anchor)
		}
		if pattern != test.expected {
			t.Fatalf("normalize(%q, %q) = %q, want %q", test.formula, test.anchor, pattern, test.expected)
		}
	}
}
