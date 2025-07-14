package mcptools

import (
	"time"
)

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
