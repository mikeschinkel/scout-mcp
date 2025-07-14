package mcptools

type RiskLevel string

func (r RiskLevel) String() string {
	return string(r)
}

func (r RiskLevel) TitleCase() string {
	return titleCase(string(r))
}

const LowRisk RiskLevel = "low"
const MediumRisk RiskLevel = "medium"
const HighRisk RiskLevel = "high"

func getRiskIcon(rl RiskLevel) string {
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
