package model

// Analyst priority bands are derived routing hints; rule Severity is unchanged.
const (
	PriorityCritical = "critical"
	PriorityHigh     = "high"
	PriorityMedium   = "medium"
	PriorityLow      = "low"
)

// ImpactFactor explains one deterministic input to the derived priority band.
type ImpactFactor struct {
	Code        string
	Explanation string
}
