package model

import (
	"bytes"
	"encoding/json"
)

const (
	PromptBundleVersionV1   = "1"
	PromptContractVersionV1 = "1"
)

// PromptBundleOptions controls redaction, exclusions, and optional workbook slices
// before preview, copy, or export. Workbook slices stay off unless explicitly enabled.
type PromptBundleOptions struct {
	ExcludeSheets        []string `json:"exclude_sheets,omitempty"`
	ExcludeCells         []string `json:"exclude_cells,omitempty"`
	EnableWorkbookSlices bool     `json:"enable_workbook_slices,omitempty"`
	MaxSliceRows         int      `json:"max_slice_rows,omitempty"`
	MaxSliceColumns      int      `json:"max_slice_columns,omitempty"`
	MaxPacketBytes       int      `json:"max_packet_bytes,omitempty"`
	UserObjective        string   `json:"user_objective,omitempty"`
}

// WorkbookSlice holds optional bounded cell context when slices are enabled.
// Slice 2 does not populate values by default; caps apply when enabled in later work.
type WorkbookSlice struct {
	Range  string     `json:"range"`
	Sheet  string     `json:"sheet"`
	Values [][]string `json:"values,omitempty"`
}

// PromptBundleV1 is the versioned prompt contract for manual LLM use.
type PromptBundleV1 struct {
	BundleVersion  string           `json:"bundle_version"`
	PromptVersion  string           `json:"prompt_version"`
	Instructions   string           `json:"instructions"`
	ResponseSchema map[string]any   `json:"response_schema"`
	EvidencePacket EvidencePacketV1 `json:"evidence_packet"`
	WorkbookSlices []WorkbookSlice  `json:"workbook_slices,omitempty"`
	Prompt         string           `json:"prompt"`
}

func (b *PromptBundleV1) CanonicalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(b); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
