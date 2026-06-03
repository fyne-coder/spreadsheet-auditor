package model

type RuleDefinition struct {
	RuleID      string
	Title       string
	Severity    string
	Category    string
	Rationale   string
	Remediation string
}

var Rules = map[string]RuleDefinition{
	"BROKEN_REF_VALUE": {
		RuleID:      "BROKEN_REF_VALUE",
		Title:       "Broken reference value",
		Severity:    "high",
		Category:    "formula_integrity",
		Rationale:   "A cell displays #REF!, which usually means a deleted row, column, sheet, or workbook link left a dangling reference.",
		Remediation: "Open the cell and trace the formula to the missing reference. Restore the source range or replace the reference with a valid sheet/range.",
	},
	"BROKEN_REF_FORMULA": {
		RuleID:      "BROKEN_REF_FORMULA",
		Title:       "Broken reference in formula",
		Severity:    "high",
		Category:    "formula_integrity",
		Rationale:   "The stored formula text still contains #REF!, so the workbook may calculate incorrectly even when neighboring cells look fine.",
		Remediation: "Repair or remove the broken reference inside the formula, then recalculate downstream dependents after the source range is restored.",
	},
	"EXCEL_ERROR_VALUE": {
		RuleID:      "EXCEL_ERROR_VALUE",
		Title:       "Excel error value",
		Severity:    "high",
		Category:    "formula_integrity",
		Rationale:   "A cell displays an Excel error sentinel, which usually means the formula or inputs are invalid for the current workbook state.",
		Remediation: "Open the cell, inspect the formula and its inputs, and fix the underlying reference, type, or calculation problem.",
	},
	"EXCEL_ERROR_FORMULA": {
		RuleID:      "EXCEL_ERROR_FORMULA",
		Title:       "Excel error sentinel in formula",
		Severity:    "high",
		Category:    "formula_integrity",
		Rationale:   "The stored formula text contains an Excel error sentinel, so the workbook may calculate incorrectly even when neighboring cells look fine.",
		Remediation: "Repair or remove the error-producing expression inside the formula, then recalculate downstream dependents after the source inputs are corrected.",
	},
	"EXTERNAL_WORKBOOK_REFERENCE": {
		RuleID:      "EXTERNAL_WORKBOOK_REFERENCE",
		Title:       "External workbook reference",
		Severity:    "medium",
		Category:    "lineage",
		Rationale:   "The formula points at another workbook file, which makes results depend on external file paths, versions, and availability.",
		Remediation: "Document the external dependency, confirm the linked file path, or paste values into a controlled input range when the model must be self-contained.",
	},
	"FORMULA_PARSE_ERROR": {
		RuleID:      "FORMULA_PARSE_ERROR",
		Title:       "Formula parse error",
		Severity:    "medium",
		Category:    "formula_integrity",
		Rationale:   "The static tokenizer could not read the formula grammar, so additional lint rules may be skipped for this cell.",
		Remediation: "Open the formula in Excel to confirm it is valid, then simplify unusual syntax or named-range constructs and rerun the audit.",
	},
	"HARDCODED_NUMERIC_CONSTANT": {
		RuleID:      "HARDCODED_NUMERIC_CONSTANT",
		Title:       "Hardcoded numeric constant",
		Severity:    "medium",
		Category:    "formula_integrity",
		Rationale:   "Numeric literals embedded in formulas are harder to audit, reuse, and update than values stored in labeled input cells.",
		Remediation: "Move constants into clearly labeled input cells or named ranges and reference those cells from the formula.",
	},
	"VOLATILE_FUNCTION": {
		RuleID:      "VOLATILE_FUNCTION",
		Title:       "Volatile function",
		Severity:    "medium",
		Category:    "performance",
		Rationale:   "Volatile functions recalculate on every workbook change, which can slow large models and make timing-dependent results harder to reason about.",
		Remediation: "Replace volatile functions where possible, isolate them on a dedicated calculation sheet, or document why live recalculation is required.",
	},
	"WHOLE_COLUMN_RANGE": {
		RuleID:      "WHOLE_COLUMN_RANGE",
		Title:       "Whole-column range reference",
		Severity:    "medium",
		Category:    "performance",
		Rationale:   "Entire-column references expand to very large ranges and can slow recalculation or hide accidental blanks in large worksheets.",
		Remediation: "Restrict the range to the populated rows/columns, or convert the logic to a structured table with explicit bounds.",
	},
	"FORMULA_PATTERN_ANOMALY": {
		RuleID:      "FORMULA_PATTERN_ANOMALY",
		Title:       "Formula pattern anomaly",
		Severity:    "medium",
		Category:    "formula_integrity",
		Rationale:   "A formula in a copied row/column run does not match the normalized pattern used by its immediate neighbors, which often signals a paste error or one-off adjustment.",
		Remediation: "Compare the outlier cell with its neighboring formulas, confirm whether the difference is intentional, and align the formula with the copied pattern when it is not.",
	},
}
