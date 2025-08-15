package langutil

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrLanguageNotSupported = errors.New("language/file type not supported")
)

// Processor represents a programming language handler
type Processor interface {
	// Language returns the language name (e.g., "go", "javascript")
	Language() Language

	// SupportedPartTypes returns the part types this language supports
	SupportedPartTypes() []PartType

	// FindPart finds a specific part in the source code
	FindPart(PartArgs) (*PartInfo, error)

	// ReplacePart replaces a specific part in the source code
	ReplacePart(PartArgs) (string, error)

	// ValidateContent validates that content is appropriate for the part type
	ValidateContent(PartArgs) error

	// ValidateSyntax validates that the entire source code is syntactically correct
	ValidateSyntax(source string) error
}

// Registry holds registered languages
var processors = make(map[string]Processor)

// RegisterProcessor registers a language handler
func RegisterProcessor(p Processor) {
	name := strings.ToLower(string(p.Language()))
	processors[name] = p
}

// GetProcessor retrieves a registered language handler
func GetProcessor(name Language) (p Processor, err error) {
	p, exists := processors[strings.ToLower(string(name))]
	if !exists {
		err = errors.Join(ErrLanguageNotSupported,
			fmt.Errorf("language=%s", name),
			fmt.Errorf("available=[%v]", GetLanguages()),
		)
	}
	return p, err
}

// GetLanguages returns a list of all registered processors
func GetLanguages() (langs []Language) {
	langs = make([]Language, 0, len(processors))
	for _, p := range processors {
		langs = append(langs, p.Language())
	}
	return langs
}
