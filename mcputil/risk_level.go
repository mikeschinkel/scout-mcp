package mcputil

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// RiskLevel represents the risk assessment level for MCP tool operations.
// Risk levels are used to determine approval requirements and safety measures
// for file system operations and other potentially dangerous actions.
type RiskLevel string

// String returns the string representation of the risk level.
// This method provides the lowercase string value of the risk level.
func (r RiskLevel) String() string {
	return string(r)
}

// TitleCase returns the title-cased representation of the risk level
// for display purposes in user interfaces.
func (r RiskLevel) TitleCase() string {
	return cases.Title(language.English).String(string(r))
}

const (
	// LowRisk indicates operations that are read-only or otherwise safe.
	// These operations typically don't require user approval and include
	// file reading, directory listing, and configuration queries.
	LowRisk RiskLevel = "low"

	// MediumRisk indicates operations that modify files but with user approval.
	// These operations include file creation, content updates, and similar
	// modifications that should be reviewed by the user before execution.
	MediumRisk RiskLevel = "medium"

	// HighRisk indicates operations that could cause data loss or system changes.
	// These operations include file deletion, bulk modifications, and other
	// potentially destructive actions that require explicit user confirmation.
	HighRisk RiskLevel = "high"
)

// GetRiskIcon returns an emoji icon representing the risk level.
// This function provides visual indicators for risk assessment in user interfaces.
func GetRiskIcon(rl RiskLevel) string {
	switch rl {
	case LowRisk:
		return "🟢"
	case MediumRisk:
		return "🟡"
	case HighRisk:
		return "🔴"
	default:
		return "⚪"
	}
}
