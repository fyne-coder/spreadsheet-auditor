package formula

import (
	"sort"

	"github.com/xuri/efp"
)

// Token mirrors the narrow tokenizer fields used by the Python analyzer.
type Token struct {
	Type    string
	Subtype string
	Value   string
}

// Tokenize parses a formula with the maintained efp tokenizer.
func Tokenize(formula string) []Token {
	parser := efp.ExcelParser()
	raw := parser.Parse(formula)
	tokens := make([]Token, 0, len(raw))
	for _, token := range raw {
		tokens = append(tokens, Token{
			Type:    token.TType,
			Subtype: token.TSubType,
			Value:   token.TValue,
		})
	}
	return tokens
}

// NumberOperands returns sorted unique numeric literal values.
func NumberOperands(tokens []Token) []string {
	seen := map[string]struct{}{}
	for _, token := range tokens {
		if token.Type == efp.TokenTypeOperand && token.Subtype == efp.TokenSubTypeNumber {
			seen[token.Value] = struct{}{}
		}
	}
	if len(seen) == 0 {
		return nil
	}
	values := make([]string, 0, len(seen))
	for value := range seen {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}
