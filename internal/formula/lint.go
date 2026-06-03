package formula

import (
	"regexp"
	"sort"
	"strings"

	"github.com/xuri/efp"
)

var (
	wholeColumnRangeRE = regexp.MustCompile(`^[A-Z]{1,3}:[A-Z]{1,3}$`)
	wholeRowRangeRE    = regexp.MustCompile(`^\d+:\d+$`)
	tableWholeColumnRE = regexp.MustCompile(`^[A-Z_][A-Z0-9_.]*\[[A-Z0-9_][A-Z0-9_ ]*\]$`)
)

var volatileFunctions = map[string]struct{}{
	"CELL":        {},
	"INFO":        {},
	"INDIRECT":    {},
	"NOW":         {},
	"OFFSET":      {},
	"RAND":        {},
	"RANDBETWEEN": {},
	"TODAY":       {},
}

var dynamicArrayVolatileFunctions = map[string]struct{}{
	"RANDARRAY": {},
}

// VolatileFunctionResult captures token-aware volatile-function lint output.
type VolatileFunctionResult struct {
	Functions    []string
	DynamicArray bool
	DynamicNames []string
}

// VolatileFunctions inspects function-name tokens and ignores string literals.
func VolatileFunctions(formula string) VolatileFunctionResult {
	tokens := Tokenize(formula)
	seen := map[string]struct{}{}
	dynamicSeen := map[string]struct{}{}
	for _, token := range tokens {
		if token.Type != efp.TokenTypeFunction || token.Subtype != efp.TokenSubTypeStart {
			continue
		}
		name := normalizeFunctionName(token.Value)
		if _, ok := dynamicArrayVolatileFunctions[name]; ok {
			dynamicSeen[name] = struct{}{}
			seen[name] = struct{}{}
			continue
		}
		if _, ok := volatileFunctions[name]; ok {
			seen[name] = struct{}{}
		}
	}
	if len(seen) == 0 {
		return VolatileFunctionResult{}
	}
	return VolatileFunctionResult{
		Functions:    sortedKeys(seen),
		DynamicArray: len(dynamicSeen) > 0,
		DynamicNames: sortedKeys(dynamicSeen),
	}
}

// RangeReference describes a whole-sheet range pattern found in formula text.
type RangeReference struct {
	Kind      string
	Reference string
}

// WholeRangeReferences returns token-aware whole-column, whole-row, max-sheet,
// and structured-table whole-column references.
func WholeRangeReferences(formula string) []RangeReference {
	tokens := Tokenize(formula)
	seen := map[string]RangeReference{}
	for _, token := range tokens {
		if token.Type != efp.TokenTypeOperand || token.Subtype != efp.TokenSubTypeRange {
			continue
		}
		if kind, ok := classifyRangeReference(token.Value); ok {
			key := kind + "|" + strings.ToUpper(token.Value)
			seen[key] = RangeReference{Kind: kind, Reference: token.Value}
		}
	}
	if len(seen) == 0 {
		return nil
	}
	refs := make([]RangeReference, 0, len(seen))
	for _, ref := range seen {
		refs = append(refs, ref)
	}
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Kind != refs[j].Kind {
			return refs[i].Kind < refs[j].Kind
		}
		return strings.ToUpper(refs[i].Reference) < strings.ToUpper(refs[j].Reference)
	})
	return refs
}

func classifyRangeReference(raw string) (string, bool) {
	ref := strings.ToUpper(strings.TrimSpace(raw))
	if idx := strings.LastIndex(ref, "!"); idx >= 0 {
		ref = ref[idx+1:]
	}
	ref = strings.ReplaceAll(ref, "$", "")

	if wholeColumnRangeRE.MatchString(ref) {
		return "whole_column", true
	}
	if wholeRowRangeRE.MatchString(ref) {
		return "whole_row", true
	}
	if isMaxSheetRange(ref) {
		return "max_sheet_range", true
	}
	if tableWholeColumnRE.MatchString(ref) {
		return "table_whole_column", true
	}
	return "", false
}

func isMaxSheetRange(ref string) bool {
	if strings.Contains(ref, "XFD1048576") {
		return true
	}
	if strings.HasSuffix(ref, ":XFD") || strings.HasSuffix(ref, ":1048576") {
		return true
	}
	if strings.HasPrefix(ref, "XFD:") || strings.HasPrefix(ref, "1048576:") {
		return true
	}
	return false
}

func normalizeFunctionName(name string) string {
	upper := strings.ToUpper(strings.TrimSpace(name))
	if idx := strings.LastIndex(upper, "."); idx >= 0 {
		upper = upper[idx+1:]
	}
	return upper
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for value := range values {
		keys = append(keys, value)
	}
	sort.Strings(keys)
	return keys
}
