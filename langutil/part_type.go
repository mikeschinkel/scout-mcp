package langutil

// PartType represents the type of code construct
type PartType string

// GetSupportedPartTypes returns supported part types for a language
func GetSupportedPartTypes(language Language) ([]PartType, error) {
	p, err := GetProcessor(language)
	if err != nil {
		return nil, err
	}

	return p.SupportedPartTypes(), nil
}
