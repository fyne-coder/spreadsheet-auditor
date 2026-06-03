package audit

import (
	"strings"

	"github.com/xuri/excelize/v2"
	"spreadsheet-auditor/internal/formula"
	"spreadsheet-auditor/internal/model"
)

func sheetStates(file *excelize.File, sheetNames []string) (map[string]string, error) {
	states := make(map[string]string, len(sheetNames))
	for _, name := range sheetNames {
		state, err := sheetState(file, name)
		if err != nil {
			return nil, err
		}
		states[name] = state
	}
	return states, nil
}

func sheetState(file *excelize.File, sheetName string) (string, error) {
	visible, err := file.GetSheetVisible(sheetName)
	if err != nil {
		return "", err
	}
	if visible {
		return "visible", nil
	}

	if file.WorkBook == nil || file.WorkBook.Sheets.Sheet == nil {
		return "hidden", nil
	}
	for _, sheet := range file.WorkBook.Sheets.Sheet {
		if sheet.Name != sheetName {
			continue
		}
		switch strings.TrimSpace(sheet.State) {
		case "veryHidden":
			return "veryHidden", nil
		default:
			return "hidden", nil
		}
	}
	return "hidden", nil
}

func scanSheet(file *excelize.File, sheetName string) (int, string, []model.Issue, error) {
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return 0, "", nil, err
	}

	formulaCount := 0
	maxRow := 0
	maxCol := 0
	hasContent := false
	var issues []model.Issue
	var formulaRecords []formula.CellRecord

	for rowIndex, columns := range rows {
		row := rowIndex + 1
		for colIndex, displayed := range columns {
			col := colIndex + 1
			coordinate, err := excelize.CoordinatesToCellName(col, row)
			if err != nil {
				return 0, "", nil, err
			}

			formulaText, isFormula, err := cellFormula(file, sheetName, coordinate, displayed)
			if err != nil {
				return 0, "", nil, err
			}
			if isFormula {
				formulaCount++
				hasContent = true
				if row > maxRow {
					maxRow = row
				}
				if col > maxCol {
					maxCol = col
				}
				issues = append(issues, lintFormula(sheetName, coordinate, formulaText)...)
				if pattern, ok := formula.NormalizeFormula(formulaText, coordinate); ok {
					formulaRecords = append(formulaRecords, formula.CellRecord{
						Coordinate: coordinate,
						Row:        row,
						Column:     col,
						Formula:    formulaText,
						Pattern:    pattern,
					})
				}
				continue
			}

			if displayed != "" {
				hasContent = true
				if row > maxRow {
					maxRow = row
				}
				if col > maxCol {
					maxCol = col
				}
			}

			if displayed == "#REF!" {
				issues = append(issues, model.BuildIssue(
					"BROKEN_REF_VALUE",
					"Cell contains a broken #REF! value.",
					sheetName,
					coordinate,
					"",
					nil,
				))
			} else if code, ok := displayedExcelError(displayed); ok {
				issues = append(issues, excelErrorValueIssue(sheetName, coordinate, code))
			}
		}
	}

	issues = append(issues, formula.FindPatternAnomalies(sheetName, formulaRecords)...)

	usedRange, err := openpyxlUsedRange(hasContent, maxCol, maxRow)
	if err != nil {
		return 0, "", nil, err
	}
	return formulaCount, usedRange, issues, nil
}

func openpyxlUsedRange(hasContent bool, maxCol, maxRow int) (string, error) {
	if !hasContent {
		return "A1:A1", nil
	}
	end, err := excelize.CoordinatesToCellName(maxCol, maxRow)
	if err != nil {
		return "", err
	}
	return "A1:" + end, nil
}

func cellFormula(file *excelize.File, sheetName, coordinate, displayed string) (string, bool, error) {
	formula, err := file.GetCellFormula(sheetName, coordinate)
	if err != nil {
		return "", false, err
	}
	if formula != "" {
		if !strings.HasPrefix(formula, "=") {
			formula = "=" + formula
		}
		return formula, true, nil
	}

	if strings.HasPrefix(displayed, "=") {
		return displayed, true, nil
	}

	cellType, err := file.GetCellType(sheetName, coordinate)
	if err != nil {
		return "", false, err
	}
	if cellType == excelize.CellTypeFormula && displayed != "" {
		text := displayed
		if !strings.HasPrefix(text, "=") {
			text = "=" + text
		}
		return text, true, nil
	}

	return "", false, nil
}
