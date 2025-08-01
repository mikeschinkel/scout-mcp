package mcputil

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type RiskLevel string

func (r RiskLevel) String() string {
	return string(r)
}

func (r RiskLevel) TitleCase() string {
	return cases.Title(language.English).String(string(r))
}

const LowRisk RiskLevel = "low"
const MediumRisk RiskLevel = "medium"
const HighRisk RiskLevel = "high"

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
