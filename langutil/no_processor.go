package langutil

var _ Processor = (*NoProcessor)(nil)

// NoProcessor implements the Processer interface for files with no language (plain text. etc.)
type NoProcessor struct{}

func init() {
	RegisterProcessor(&NoProcessor{})
}

func (n NoProcessor) Language() Language {
	return NoLanguage
}

func (n NoProcessor) SupportedPartTypes() []PartType {
	return make([]PartType, 0)
}

func (n NoProcessor) FindPart(PartArgs) (*PartInfo, error) {
	panic("implement me")
	return nil, nil
}

func (n NoProcessor) ReplacePart(PartArgs) (string, error) {
	panic("implement me")
	return "", nil
}

func (n NoProcessor) ValidateContent(PartArgs) error {
	return nil
}

func (n NoProcessor) ValidateSyntax(string) error {
	return nil
}
