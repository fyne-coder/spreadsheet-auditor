package reviewpack

import (
	"bytes"
	"encoding/json"
	"sort"
)

// CanonicalDetailsJSON encodes issue details with sorted keys, no HTML escaping,
// and no trailing newline.
func CanonicalDetailsJSON(details map[string]any) (string, error) {
	if len(details) == 0 {
		return "{}", nil
	}
	normalized := sortMapKeys(details)
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(normalized); err != nil {
		return "", err
	}
	payload := buffer.Bytes()
	if len(payload) > 0 && payload[len(payload)-1] == '\n' {
		payload = payload[:len(payload)-1]
	}
	return string(payload), nil
}

func sortMapKeys(value map[string]any) map[string]any {
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
