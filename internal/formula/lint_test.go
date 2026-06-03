package formula

import "testing"

func TestVolatileFunctionsIgnoresStringLiterals(t *testing.T) {
	result := VolatileFunctions(`=IF(A1=1,"Call TODAY() before noon","")`)
	if len(result.Functions) != 0 {
		t.Fatalf("expected no volatile functions, got %#v", result.Functions)
	}
}

func TestVolatileFunctionsDetectsRealCalls(t *testing.T) {
	result := VolatileFunctions("=TODAY()+NOW()")
	if len(result.Functions) != 2 {
		t.Fatalf("expected two volatile functions, got %#v", result.Functions)
	}
}

func TestVolatileFunctionsDetectsRandarrayWithDynamicArrayFlag(t *testing.T) {
	result := VolatileFunctions("=RANDARRAY(5,1)")
	if len(result.Functions) != 1 || result.Functions[0] != "RANDARRAY" {
		t.Fatalf("expected RANDARRAY, got %#v", result.Functions)
	}
	if !result.DynamicArray {
		t.Fatal("expected dynamic_array flag")
	}
	if len(result.DynamicNames) != 1 || result.DynamicNames[0] != "RANDARRAY" {
		t.Fatalf("expected dynamic names [RANDARRAY], got %#v", result.DynamicNames)
	}
}

func TestWholeRangeReferencesDetectsWholeRowAndColumn(t *testing.T) {
	refs := WholeRangeReferences("=SUMPRODUCT(A:A,B:B)+SUM(1:1)")
	if len(refs) != 3 {
		t.Fatalf("expected three whole ranges, got %#v", refs)
	}
	kinds := map[string]struct{}{}
	for _, ref := range refs {
		kinds[ref.Kind] = struct{}{}
	}
	if _, ok := kinds["whole_column"]; !ok {
		t.Fatalf("expected whole_column, got %#v", refs)
	}
	if _, ok := kinds["whole_row"]; !ok {
		t.Fatalf("expected whole_row, got %#v", refs)
	}
}

func TestWholeRangeReferencesDetectsMaxSheetRange(t *testing.T) {
	refs := WholeRangeReferences("=SUM(A1:XFD1048576)")
	if len(refs) != 1 || refs[0].Kind != "max_sheet_range" {
		t.Fatalf("expected max_sheet_range, got %#v", refs)
	}
}

func TestWholeRangeReferencesDetectsTableWholeColumn(t *testing.T) {
	refs := WholeRangeReferences("=SUM(Table1[Amount])")
	if len(refs) != 1 || refs[0].Kind != "table_whole_column" {
		t.Fatalf("expected table_whole_column, got %#v", refs)
	}
}

func TestWholeRangeReferencesIgnoresStringLiterals(t *testing.T) {
	refs := WholeRangeReferences(`=IF(A1=1,"A:A is not a range","")`)
	if len(refs) != 0 {
		t.Fatalf("expected no whole ranges in string literal, got %#v", refs)
	}
}
