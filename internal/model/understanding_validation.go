package model

// UnderstandingValidationResult is the UI-friendly outcome of pasted LLM JSON
// validation against the current evidence packet citation map.
type UnderstandingValidationResult struct {
	// CitationsResolved means every cited ID in the pasted report exists in the
	// evidence packet citation map. It does not verify claim substance.
	CitationsResolved bool `json:"citations_resolved"`
	// Valid is a compatibility alias for CitationsResolved during this branch.
	Valid      bool                   `json:"valid"`
	Report     *UnderstandingReportV1 `json:"report,omitempty"`
	Rejects    []CitationReject       `json:"rejects,omitempty"`
	ParseError string                 `json:"parse_error,omitempty"`
}
