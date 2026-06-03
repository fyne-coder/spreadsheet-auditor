package promptpack

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

const redactedPlaceholder = "[REDACTED]"

var (
	unixPathPattern   = regexp.MustCompile(`/(?:Users|home|var|tmp|private|Volumes|mnt|opt|data|srv|etc|share)(?:/[^\s'")\]]+)+`)
	winPathPattern    = regexp.MustCompile(`(?i)[A-Z]:\\(?:[^\s'")\]]+\\)*[^\s'")\]]+`)
	uncPathPattern    = regexp.MustCompile(`\\\\[^\s'")\]]+`)
	uriPathPattern    = regexp.MustCompile(`(?i)(?:(?:file|https?|smb|ftp|s3|gs|azure)://|mailto:)[^\s'")\]]+`)
	connStringPattern = regexp.MustCompile(`(?i)(?:Server|Data Source|Initial Catalog|User ID|Password|Trusted_Connection)\s*=\s*[^;'"]+`)
	secretPattern     = regexp.MustCompile(`(?i)(?:api[_-]?key|secret|token|password|bearer)\s*[:=]\s*\S+`)
	awsKeyPattern     = regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)
	githubPATPattern  = regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9_]{20,}\b`)
	emailPattern      = regexp.MustCompile(`\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`)
)

func redactString(value string) string {
	if value == "" {
		return value
	}
	result := value
	result = replacePromptDelimiters(result)
	result = replaceRedactedMatches(unixPathPattern, result)
	result = replaceRedactedMatches(winPathPattern, result)
	result = replaceRedactedMatches(uncPathPattern, result)
	result = replaceRedactedMatches(uriPathPattern, result)
	result = replaceRedactedMatches(connStringPattern, result)
	result = replaceRedactedMatches(secretPattern, result)
	result = replaceRedactedMatches(awsKeyPattern, result)
	result = replaceRedactedMatches(githubPATPattern, result)
	result = replaceRedactedMatches(emailPattern, result)
	return result
}

func redactWorkbookName(value string) string {
	if value == "" {
		return value
	}
	sum := sha256.Sum256([]byte(value))
	return "[WORKBOOK:" + hex.EncodeToString(sum[:4]) + "]"
}

func replacePromptDelimiters(value string) string {
	result := strings.ReplaceAll(value, evidenceBeginDelimiter, redactedToken(evidenceBeginDelimiter))
	result = strings.ReplaceAll(result, evidenceEndDelimiter, redactedToken(evidenceEndDelimiter))
	return result
}

func replaceRedactedMatches(pattern *regexp.Regexp, value string) string {
	return pattern.ReplaceAllStringFunc(value, redactedToken)
}

func redactedToken(value string) string {
	sum := sha256.Sum256([]byte(value))
	return redactedPlaceholder + ":" + hex.EncodeToString(sum[:4])
}

func redactAny(value any) any {
	switch typed := value.(type) {
	case string:
		return redactString(typed)
	case []string:
		out := make([]string, len(typed))
		for index, item := range typed {
			out[index] = redactString(item)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for index, item := range typed {
			out[index] = redactAny(item)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[key] = redactAny(item)
		}
		return out
	default:
		return value
	}
}

func containsAbsolutePath(text string) bool {
	return unixPathPattern.MatchString(text) ||
		winPathPattern.MatchString(text) ||
		uncPathPattern.MatchString(text) ||
		uriPathPattern.MatchString(text) ||
		strings.Contains(text, "/Users/") ||
		strings.Contains(text, ":\\")
}
