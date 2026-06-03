package promptpack

import (
	"fmt"

	"spreadsheet-auditor/internal/evidence"
	"spreadsheet-auditor/internal/model"
)

const (
	defaultMaxSliceRows    = 50
	defaultMaxSliceColumns = 20
	defaultMaxPacketBytes  = 512 * 1024
)

// PreparePacket builds an evidence packet, applies exclusions and redaction, and
// never mutates deterministic scan semantics in the source audit report.
func PreparePacket(report *model.AuditReport, options model.PromptBundleOptions) (*model.EvidencePacketV1, error) {
	if options.EnableWorkbookSlices {
		return nil, fmt.Errorf("workbook slices are not implemented; disable enable_workbook_slices")
	}
	packet, err := evidence.BuildPacket(report)
	if err != nil {
		return nil, err
	}
	applyExclusions(packet, options)
	redactPacket(packet)
	packet.CitationMap = rebuildCitationMap(packet)
	return packet, nil
}

// BuildBundle assembles a versioned prompt bundle for copy/export.
func BuildBundle(report *model.AuditReport, options model.PromptBundleOptions) (*model.PromptBundleV1, error) {
	packet, err := PreparePacket(report, options)
	if err != nil {
		return nil, err
	}
	packetJSON, err := packet.CanonicalJSON()
	if err != nil {
		return nil, err
	}
	if err := enforcePacketByteCap(packetJSON, options); err != nil {
		return nil, err
	}

	bundle := &model.PromptBundleV1{
		BundleVersion:  model.PromptBundleVersionV1,
		PromptVersion:  model.PromptContractVersionV1,
		Instructions:   InstructionsV1,
		ResponseSchema: UnderstandingReportSchemaV1(),
		EvidencePacket: *packet,
		WorkbookSlices: workbookSlices(options),
		Prompt:         assemblePrompt(InstructionsV1, packetJSON, options),
	}
	return bundle, nil
}

func redactPacket(packet *model.EvidencePacketV1) {
	packet.Workbook.Name = redactWorkbookName(packet.Workbook.Name)
	for index := range packet.Sheets {
		packet.Sheets[index].Name = redactString(packet.Sheets[index].Name)
		packet.Sheets[index].UsedRange = redactString(packet.Sheets[index].UsedRange)
	}
	for index := range packet.Issues {
		issue := &packet.Issues[index]
		issue.IssueID = redactString(issue.IssueID)
		issue.Sheet = redactString(issue.Sheet)
		issue.Cell = redactString(issue.Cell)
		issue.Formula = redactString(issue.Formula)
		issue.Message = redactString(issue.Message)
		issue.Rationale = redactString(issue.Rationale)
		issue.Remediation = redactString(issue.Remediation)
		issue.Title = redactString(issue.Title)
		if len(issue.Details) > 0 {
			issue.Details = redactAny(issue.Details).(map[string]any)
		}
	}
	for index := range packet.FormulaFamilies {
		family := &packet.FormulaFamilies[index]
		family.Sheet = redactString(family.Sheet)
		family.ExpectedPattern = redactString(family.ExpectedPattern)
		family.LocalPattern = redactString(family.LocalPattern)
		family.OutlierCell = redactString(family.OutlierCell)
		family.Representative = redactString(family.Representative)
		for cellIndex, cell := range family.MemberCells {
			family.MemberCells[cellIndex] = redactString(cell)
		}
	}
}

func workbookSlices(options model.PromptBundleOptions) []model.WorkbookSlice {
	if !options.EnableWorkbookSlices {
		return nil
	}
	_, _, _ = sliceCaps(options)
	return []model.WorkbookSlice{}
}

func sliceCaps(options model.PromptBundleOptions) (maxRows, maxColumns, maxBytes int) {
	maxRows = options.MaxSliceRows
	if maxRows <= 0 {
		maxRows = defaultMaxSliceRows
	}
	maxColumns = options.MaxSliceColumns
	if maxColumns <= 0 {
		maxColumns = defaultMaxSliceColumns
	}
	maxBytes = options.MaxPacketBytes
	if maxBytes <= 0 {
		maxBytes = defaultMaxPacketBytes
	}
	return maxRows, maxColumns, maxBytes
}

func enforcePacketByteCap(packetJSON []byte, options model.PromptBundleOptions) error {
	_, _, maxBytes := sliceCaps(options)
	if len(packetJSON) > maxBytes {
		return fmt.Errorf("evidence packet exceeds max_packet_bytes cap (%d > %d)", len(packetJSON), maxBytes)
	}
	return nil
}
