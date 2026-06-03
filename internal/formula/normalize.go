package formula

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/efp"
	"github.com/xuri/excelize/v2"
)

var cellReferenceRE = regexp.MustCompile(`(?i)^(?P<col_abs>\$)?(?P<col>[A-Z]{1,3})(?P<row_abs>\$)?(?P<row>\d+)$`)

// NormalizeFormula returns a position-relative pattern for comparing copied formulas.
func NormalizeFormula(formula, anchorCell string) (string, bool) {
	colName, row, err := excelize.SplitCellName(anchorCell)
	if err != nil {
		return "", false
	}
	anchorCol, err := excelize.ColumnNameToNumber(colName)
	if err != nil {
		return "", false
	}

	tokens := Tokenize(formula)
	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token.Type == efp.TokenTypeOperand && token.Subtype == efp.TokenSubTypeRange {
			parts = append(parts, normalizeRangeOperand(token.Value, anchorCol, row))
		} else {
			parts = append(parts, token.Value)
		}
	}
	return strings.Join(parts, ""), true
}

func normalizeRangeOperand(value string, anchorCol, anchorRow int) string {
	sheetPrefix := ""
	reference := value
	if idx := strings.LastIndex(value, "!"); idx >= 0 {
		sheetPrefix = value[:idx+1]
		reference = value[idx+1:]
	}

	if strings.Contains(reference, ":") {
		start, end, _ := strings.Cut(reference, ":")
		return sheetPrefix +
			normalizeCellReference(start, anchorCol, anchorRow) +
			":" +
			normalizeCellReference(end, anchorCol, anchorRow)
	}
	return sheetPrefix + normalizeCellReference(reference, anchorCol, anchorRow)
}

func normalizeCellReference(reference string, anchorCol, anchorRow int) string {
	match := cellReferenceRE.FindStringSubmatch(reference)
	if match == nil {
		return reference
	}

	colAbs := match[1] != ""
	rowAbs := match[3] != ""
	colName := strings.ToUpper(match[2])
	colIndex, err := excelize.ColumnNameToNumber(colName)
	if err != nil {
		return reference
	}
	rowIndex, err := strconv.Atoi(match[4])
	if err != nil {
		return reference
	}

	var columnPart string
	if colAbs {
		columnPart = "$" + colName
	} else {
		columnPart = strconv.Itoa(colIndex - anchorCol)
	}

	var rowPart string
	if rowAbs {
		rowPart = "$" + match[4]
	} else {
		rowPart = strconv.Itoa(rowIndex - anchorRow)
	}
	return "R" + rowPart + "C" + columnPart
}
