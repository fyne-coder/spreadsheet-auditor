package priority

import (
	"testing"

	"spreadsheet-auditor/internal/model"
)

func TestAssignIsolatedTodayIsLowerThanRiskyVolatileCluster(t *testing.T) {
	today := model.BuildIssue(
		"VOLATILE_FUNCTION",
		"volatile",
		"Model",
		"B2",
		"=TODAY()",
		map[string]any{"functions": []string{"TODAY"}},
	)
	todayPriority, todayFactors := Assign(today, "visible")

	risky := model.BuildIssue(
		"VOLATILE_FUNCTION",
		"volatile",
		"Model",
		"C2",
		"=OFFSET(A1,1,1)",
		map[string]any{"functions": []string{"OFFSET"}},
	)
	riskyPriority, riskyFactors := Assign(risky, "visible")

	if todayPriority != model.PriorityLow {
		t.Fatalf("expected TODAY-only issue to be low priority, got %q", todayPriority)
	}
	if riskyPriority == todayPriority {
		t.Fatalf("expected distinct priority for risky volatile, got both %q", todayPriority)
	}
	if riskyPriority != model.PriorityHigh {
		t.Fatalf("expected OFFSET issue to be high priority, got %q", riskyPriority)
	}
	if !hasFactorCode(todayFactors, "isolated_low_risk_volatile") {
		t.Fatalf("expected isolated low-risk factor, got %#v", todayFactors)
	}
	if !hasFactorCode(riskyFactors, "risky_volatile_functions") {
		t.Fatalf("expected risky volatile factor, got %#v", riskyFactors)
	}
}

func TestAssignHiddenStructuralIssueIsCritical(t *testing.T) {
	issue := model.BuildIssue(
		"BROKEN_REF_FORMULA",
		"broken",
		"HiddenCalc",
		"A1",
		"=#REF!",
		nil,
	)
	priority, factors := Assign(issue, "veryHidden")
	if priority != model.PriorityCritical {
		t.Fatalf("expected critical priority, got %q", priority)
	}
	if !hasFactorCode(factors, "very_hidden_sheet") {
		t.Fatalf("expected very hidden sheet factor, got %#v", factors)
	}
}

func TestApplyIsDeterministic(t *testing.T) {
	report := &model.AuditReport{
		Sheets: []model.SheetSummary{{
			Name: "Outputs", State: "visible", UsedRange: "A1:A1",
		}},
		Issues: []model.Issue{
			model.BuildIssue(
				"EXTERNAL_WORKBOOK_REFERENCE",
				"external",
				"Outputs",
				"B2",
				"=[book.xlsx]Sheet1!A1",
				map[string]any{"references": []string{"[book.xlsx]"}},
			),
		},
	}
	Apply(report)
	first := report.Issues[0].Priority
	Apply(report)
	if report.Issues[0].Priority != first {
		t.Fatalf("priority assignment changed on re-apply: %q vs %q", first, report.Issues[0].Priority)
	}
}

func hasFactorCode(factors []model.ImpactFactor, code string) bool {
	for _, factor := range factors {
		if factor.Code == code {
			return true
		}
	}
	return false
}
