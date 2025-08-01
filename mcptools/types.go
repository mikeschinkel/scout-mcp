package mcptools

import (
	"time"
)

type FileAction struct {
	Action  string `json:"action"`  // create, update, delete
	Path    string `json:"path"`    // file path
	Purpose string `json:"purpose"` // why this file is being modified
}

func (fa FileAction) Icon() string {
	switch fa.Action {
	case "create":
		return "✨"
	case "update", "modify":
		return "📝"
	case "delete":
		return "🗑️"
	case "move", "rename":
		return "📦"
	default:
		return "📄"
	}
}

type ApprovalRequest struct {
	Operation   string
	FileActions []FileAction
	Preview     string
	Risk        RiskLevel
	Impact      string
}

type TokenRequest struct {
	FileActions []FileAction
	Operations  []string
	SessionID   string
	ExpiresIn   time.Duration
}

type FileAnalysis struct {
	TotalLines   int      `json:"total_lines"`
	Complexity   string   `json:"complexity"`   // "low", "medium", "high"
	Dependencies []string `json:"dependencies"` // New imports/packages
	RiskFactors  []string `json:"risk_factors"` // Security, breaking changes, etc.
}

type FileSearchResult struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
	IsDir    bool   `json:"is_directory"`
}
