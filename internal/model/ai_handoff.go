package model

// AIHandoffPayload is the single-scan AI handoff surface for desktop preview,
// copy, and export. Prefer BuildAIHandoff over separate build methods so prompt
// text, bundle JSON, and evidence packet JSON stay aligned to one audit report.
type AIHandoffPayload struct {
	AuditHash          string          `json:"audit_hash"`
	Prompt             string          `json:"prompt"`
	PromptBundleJSON   string          `json:"prompt_bundle_json"`
	EvidencePacketJSON string          `json:"evidence_packet_json"`
	Bundle             *PromptBundleV1 `json:"bundle"`
}
