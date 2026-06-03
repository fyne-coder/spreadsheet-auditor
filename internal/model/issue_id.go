package model

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// IssueID returns a stable export-selection key:
// rule_id|sheet|cell|message|suffix.
// The readable prefix stays analyst-friendly; the suffix hashes formula and
// details so same-location issues with different evidence do not collide.
func IssueID(issue Issue) string {
	if issue.IssueID != "" {
		return issue.IssueID
	}
	return IssueBaseKey(issue) + "|" + issueIdentitySuffix(issue)
}

// ApplyIssueIDs stores backend-generated IDs on issues for UI consumers.
func ApplyIssueIDs(issues []Issue) {
	for index := range issues {
		issues[index].IssueID = IssueBaseKey(issues[index]) + "|" + issueIdentitySuffix(issues[index])
	}
}

// IssueBaseKey returns the readable issue identity without the hash suffix.
func IssueBaseKey(issue Issue) string {
	return fmt.Sprintf(
		"%s|%s|%s|%s",
		issue.RuleID,
		issue.Evidence.Sheet,
		issue.Evidence.Cell,
		issue.Message,
	)
}

func issueIdentitySuffix(issue Issue) string {
	payload, err := canonicalIssueIdentityJSON(issue)
	if err != nil {
		sum := sha256.Sum256([]byte(IssueBaseKey(issue)))
		return hex.EncodeToString(sum[:4])
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:4])
}

func canonicalIssueIdentityJSON(issue Issue) ([]byte, error) {
	payload := map[string]any{
		"details": sortMapKeys(issue.Details),
		"formula": issue.Evidence.Formula,
	}
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(payload); err != nil {
		return nil, err
	}
	out := buffer.Bytes()
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return out, nil
}

func sortMapKeys(value map[string]any) map[string]any {
	if len(value) == 0 {
		return map[string]any{}
	}
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make(map[string]any, len(keys))
	for _, key := range keys {
		result[key] = sortAnyValue(value[key])
	}
	return result
}

func sortAnyValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return sortMapKeys(typed)
	case []any:
		copied := make([]any, len(typed))
		for index, item := range typed {
			copied[index] = sortAnyValue(item)
		}
		return copied
	default:
		return value
	}
}
