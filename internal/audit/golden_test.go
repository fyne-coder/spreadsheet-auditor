package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const normalizedWorkbookPath = "<workbook>"

var goldenFixtures = []string{
	"empty_workbook",
	"single_value_cell",
	"consistent_copied_formulas",
	"lexical_cell_ordering",
	"formula_anomaly",
	"sheet_qualified_reference",
	"static_xlsm",
	"escaping_workbook_text",
	"combined_risky",
	"hidden_and_very_hidden",
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func workbookPath(t *testing.T, name string) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(repoRoot(t), "tests", "fixtures", "workbooks", name+".*"))
	if err != nil {
		t.Fatalf("glob workbook fixture: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one workbook fixture for %s, got %v", name, matches)
	}
	return matches[0]
}

func goldenPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "tests", "fixtures", "golden", name+".json")
}

func canonicalReport(t *testing.T, workbook string) []byte {
	t.Helper()
	report, err := AuditWorkbook(workbook)
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	report.WorkbookPath = normalizedWorkbookPath
	payload, err := report.CanonicalJSON()
	if err != nil {
		t.Fatalf("canonical json: %v", err)
	}
	return payload
}

func TestGoldenParityFixtures(t *testing.T) {
	for _, name := range goldenFixtures {
		t.Run(name, func(t *testing.T) {
			actual := canonicalReport(t, workbookPath(t, name))
			path := goldenPath(t, name)
			if os.Getenv("UPDATE_GOLDENS") == "1" {
				if err := os.WriteFile(path, actual, 0o644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
			}
			expected, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			if string(actual) != string(expected) {
				t.Fatalf("golden drift for %s (go %d bytes, golden %d bytes); first diff context:\n%s",
					name, len(actual), len(expected), diffContext(expected, actual))
			}
		})
	}
}

func TestLexicalIssueOrderingMatchesPython(t *testing.T) {
	report := loadActualReport(t, "lexical_cell_ordering")
	cells := make([]string, 0, len(report.Issues))
	for _, issue := range report.Issues {
		cells = append(cells, issue.Evidence.Cell)
	}
	b10 := indexOf(cells, "B10")
	b2 := indexOf(cells, "B2")
	if b10 < 0 || b2 < 0 || b10 >= b2 {
		t.Fatalf("expected B10 before B2 in lexical ordering, got %v", cells)
	}
}

type goldenReport struct {
	Issues []struct {
		Evidence struct {
			Cell string `json:"cell"`
		} `json:"evidence"`
	} `json:"issues"`
}

func loadActualReport(t *testing.T, name string) goldenReport {
	t.Helper()
	raw := canonicalReport(t, workbookPath(t, name))
	var report goldenReport
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatalf("decode actual: %v", err)
	}
	return report
}

func diffContext(expected, actual []byte) string {
	limit := len(expected)
	if len(actual) < limit {
		limit = len(actual)
	}
	for index := 0; index < limit; index++ {
		if expected[index] == actual[index] {
			continue
		}
		start := index - 40
		if start < 0 {
			start = 0
		}
		end := index + 40
		if end > len(expected) {
			end = len(expected)
		}
		endActual := index + 40
		if endActual > len(actual) {
			endActual = len(actual)
		}
		return "expected:\n" + string(expected[start:end]) + "\nactual:\n" + string(actual[start:endActual])
	}
	if len(expected) != len(actual) {
		return "payload lengths differ with matching prefix"
	}
	return "payloads differ but no byte mismatch found in shared prefix"
}

func indexOf(values []string, target string) int {
	for index, value := range values {
		if value == target {
			return index
		}
	}
	return -1
}
