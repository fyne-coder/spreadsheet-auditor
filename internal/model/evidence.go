package model

import (
	"bytes"
	"encoding/json"
)

const EvidencePacketVersionV1 = "1"

type EvidenceWorkbook struct {
	Name            string  `json:"name"`
	SupportedFormat string  `json:"supported_format"`
	Summary         Summary `json:"summary"`
}

type EvidenceSheet struct {
	FormulaCells int    `json:"formula_cells"`
	Name         string `json:"name"`
	State        string `json:"state"`
	UsedRange    string `json:"used_range"`
}

type EvidenceImpactFactor struct {
	Code        string `json:"code"`
	Explanation string `json:"explanation"`
}

type EvidenceIssue struct {
	Cell          string                 `json:"cell"`
	Category      string                 `json:"category"`
	Details       map[string]any         `json:"details,omitempty"`
	Formula       string                 `json:"formula,omitempty"`
	ImpactFactors []EvidenceImpactFactor `json:"impact_factors,omitempty"`
	IssueID       string                 `json:"issue_id"`
	Message       string                 `json:"message"`
	Priority      string                 `json:"priority,omitempty"`
	Rationale     string                 `json:"rationale"`
	Remediation   string                 `json:"remediation"`
	RuleID        string                 `json:"rule_id"`
	Severity      string                 `json:"severity"`
	Sheet         string                 `json:"sheet"`
	Title         string                 `json:"title"`
}

type FormulaFamily struct {
	ClusterID       string   `json:"formula_cluster_id"`
	ExpectedPattern string   `json:"expected_pattern"`
	LocalPattern    string   `json:"local_pattern"`
	MemberCells     []string `json:"member_cells"`
	Orientation     string   `json:"orientation"`
	OutlierCell     string   `json:"outlier_cell"`
	Representative  string   `json:"representative_formula,omitempty"`
	Sheet           string   `json:"sheet"`
}

type CitationMap struct {
	FormulaClusterIDs []string `json:"formula_cluster_ids"`
	IssueIDs          []string `json:"issue_ids"`
	RuleIDs           []string `json:"rule_ids"`
	SheetCells        []string `json:"sheet_cells"`
	SheetNames        []string `json:"sheet_names"`
}

type EvidencePacketV1 struct {
	AuditHash       string           `json:"audit_hash"`
	CitationMap     CitationMap      `json:"citation_map"`
	FormulaFamilies []FormulaFamily  `json:"formula_families"`
	Issues          []EvidenceIssue  `json:"audit_findings"`
	PacketVersion   string           `json:"packet_version"`
	Sheets          []EvidenceSheet  `json:"sheets"`
	Workbook        EvidenceWorkbook `json:"workbook"`
}

type UnderstandingClaim struct {
	Citations []string `json:"citations"`
	Claim     string   `json:"claim"`
}

type SheetRoleClaim struct {
	Citations []string `json:"citations"`
	Role      string   `json:"role"`
	Sheet     string   `json:"sheet"`
}

type FlowClaim struct {
	Citations []string `json:"citations"`
	Summary   string   `json:"summary"`
}

type RiskClaim struct {
	Citations []string `json:"citations"`
	Severity  string   `json:"severity"`
	Summary   string   `json:"summary"`
}

type CleanupAction struct {
	Action    string   `json:"action"`
	Citations []string `json:"citations"`
}

type OwnerQuestion struct {
	ContextCitations []string `json:"context_citations"`
	Question         string   `json:"question"`
}

type ConfidenceNote struct {
	Citations []string `json:"citations"`
	Note      string   `json:"note"`
}

type UnderstandingReportV1 struct {
	CleanupPlan     []CleanupAction      `json:"cleanup_plan"`
	ConfidenceNotes []ConfidenceNote     `json:"confidence_notes"`
	KeyFlows        []FlowClaim          `json:"key_flows"`
	MajorRisks      []RiskClaim          `json:"major_risks"`
	OwnerQuestions  []OwnerQuestion      `json:"owner_questions"`
	SheetRoles      []SheetRoleClaim     `json:"sheet_roles"`
	WorkbookPurpose []UnderstandingClaim `json:"workbook_purpose"`
}

type CitationReject struct {
	Citation string `json:"citation"`
	Field    string `json:"field"`
	Index    int    `json:"index"`
	Reason   string `json:"reason"`
}

func (p *EvidencePacketV1) CanonicalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(p); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
