package langutil

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// PartInfo represents information about a found part in a file
type PartInfo struct {
	StartLine   int    `json:"start_line"`
	EndLine     int    `json:"end_line"`
	StartOffset int    `json:"start_offset"`
	EndOffset   int    `json:"end_offset"`
	Content     string `json:"content"`
	Found       bool   `json:"found"`
}

// PartType represents the type of code construct
type PartType string

// Language represents a programming language handler
type Language interface {
	// Name returns the language name (e.g., "go", "javascript")
	Name() string

	// SupportedPartTypes returns the part types this language supports
	SupportedPartTypes() []PartType

	// FindPart finds a specific part in the source code
	FindPart(source string, partType PartType, partName string) (*PartInfo, error)

	// ReplacePart replaces a specific part in the source code
	ReplacePart(source string, partType PartType, partName string, newContent string) (string, error)

	// ValidateContent validates that content is appropriate for the part type
	ValidateContent(partType PartType, content string) error

	// ValidateSyntax validates that the entire source code is syntactically correct
	ValidateSyntax(source string) error
}

// Registry holds registered languages
var languages = make(map[string]Language)

// RegisterLanguage registers a language handler
func RegisterLanguage(lang Language) {
	name := strings.ToLower(lang.Name())
	languages[name] = lang
}

// GetLanguage retrieves a registered language handler
func GetLanguage(name string) (lang Language, err error) {
	name = strings.ToLower(name)
	lang, exists := languages[name]
	if !exists {
		err = fmt.Errorf("language '%s' not supported. Available languages: %v", name, GetSupportedLanguages())
	}
	return lang, err
}

// GetSupportedLanguages returns a list of all registered languages
func GetSupportedLanguages() []string {
	return slices.Collect(maps.Keys(languages))
}

// FindPart finds a part in source code using the appropriate language handler
func FindPart(language, source string, partType PartType, partName string) (*PartInfo, error) {
	lang, err := GetLanguage(language)
	if err != nil {
		return nil, err
	}

	return lang.FindPart(source, partType, partName)
}

// ReplacePart replaces a part in source code using the appropriate language handler
func ReplacePart(language, source string, partType PartType, partName string, newContent string) (string, error) {
	lang, err := GetLanguage(language)
	if err != nil {
		return "", err
	}

	return lang.ReplacePart(source, partType, partName, newContent)
}

// ValidateContent validates content for a specific part type
func ValidateContent(language string, partType PartType, content string) error {
	lang, err := GetLanguage(language)
	if err != nil {
		return err
	}

	return lang.ValidateContent(partType, content)
}

// ValidateSyntax validates syntax for a language
func ValidateSyntax(language, source string) error {
	lang, err := GetLanguage(language)
	if err != nil {
		return err
	}

	return lang.ValidateSyntax(source)
}

// GetSupportedPartTypes returns supported part types for a language
func GetSupportedPartTypes(language string) ([]PartType, error) {
	lang, err := GetLanguage(language)
	if err != nil {
		return nil, err
	}

	return lang.SupportedPartTypes(), nil
}
