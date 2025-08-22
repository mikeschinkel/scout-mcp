package langutil

// PartType represents the type of code construct that can be found or manipulated within source files.
// This type is used to specify what kind of language element should be targeted during AST operations
// such as finding, replacing, or validating code constructs.
//
// Different programming languages support different sets of part types. For example, Go supports
// functions, types, constants, variables, imports, and packages, while other languages may have
// different or additional construct types like classes, interfaces, or modules.
//
// The available part types for a specific language can be determined using GetSupportedPartTypes,
// which queries the language's registered processor for its capabilities.
//
// # Usage Examples
//
// Common part types include:
//   - "func" for functions and methods
//   - "type" for type definitions (structs, interfaces, aliases)
//   - "const" for constant declarations
//   - "var" for variable declarations
//   - "import" for import statements
//   - "package" for package declarations
//   - "class" for class definitions (in OOP languages)
//   - "interface" for interface definitions
//
// The exact string values and availability depend on the specific language processor implementation.
type PartType string

// GetSupportedPartTypes returns supported part types for a language.
// This function queries the registered processor for the specified language and returns
// a slice of all PartType values that the processor can handle. This allows callers to
// determine what kinds of code constructs can be found or manipulated for a given language.
//
// The function is useful for:
//   - Validating user input before attempting AST operations
//   - Building user interfaces that show available options
//   - Runtime discovery of language capabilities
//   - Debugging processor implementations
//
// Returns ErrLanguageNotSupported if no processor is registered for the specified language.
// The returned slice should not be modified by callers as it may be shared across calls.
//
// # Example Usage
//
//	partTypes, err := langutil.GetSupportedPartTypes(langutil.GoLanguage)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, partType := range partTypes {
//		fmt.Printf("Supported part type: %s\n", partType)
//	}
//
// # Performance
//
// This function involves a map lookup to find the processor and a method call to retrieve
// the supported types. The processor may cache the results, so repeated calls should be fast.
func GetSupportedPartTypes(language Language) ([]PartType, error) {
	p, err := GetProcessor(language)
	if err != nil {
		return nil, err
	}

	return p.SupportedPartTypes(), nil
}
