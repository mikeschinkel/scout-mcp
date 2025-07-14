package mcptools

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
