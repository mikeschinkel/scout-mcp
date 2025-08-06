package langutil

// PartInfo represents information about a found part in a file
type PartInfo struct {
	StartLine   int    `json:"start_line"`
	EndLine     int    `json:"end_line"`
	StartOffset int    `json:"start_offset"`
	EndOffset   int    `json:"end_offset"`
	Content     string `json:"content"`
	Found       bool   `json:"found"`
}

type PartArgs struct {
	Language   Language
	Content    string
	PartType   PartType
	PartName   string
	NewContent string
	Filepath   string
}

// FindPart finds a part in source code using the appropriate language handler
func FindPart(args PartArgs) (*PartInfo, error) {
	p, err := GetProcessor(args.Language)
	if err != nil {
		return nil, err
	}

	return p.FindPart(args)
}
