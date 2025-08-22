package langutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// File represents a source code file that can be validated and processed by language-specific processors.
// This structure encapsulates a file's path, detected language, associated processor, and initialization state.
// It provides a high-level interface for performing language-specific operations on source files while
// managing the underlying language detection and processor selection automatically.
//
// # Lifecycle
//
// File instances follow a specific lifecycle:
//  1. Creation with NewFile() - sets the file path
//  2. Initialization with Initialize() - detects language and validates processor availability
//  3. Operations like Validate(), ValidateAs(), ValidateSyntax() - perform file processing
//
// The File structure caches the detected language and processor to avoid repeated detection
// and lookup operations when performing multiple operations on the same file.
//
// # Language Detection
//
// Language detection is performed automatically during initialization based on the file extension.
// The detected language can be overridden by using the *As() methods that accept an explicit language parameter.
//
// # Thread Safety
//
// File instances are not thread-safe and should not be shared between goroutines without external
// synchronization. Each goroutine should use its own File instance or implement appropriate locking.
type File struct {
	filepath    string    // Absolute or relative path to the source file
	language    Language  // Detected or explicitly set language for this file
	processor   Processor // Cached processor instance for the file's language
	initialized bool      // Whether Initialize() has been called successfully
}

// NewFile creates a new File instance for the specified file path.
// This constructor function creates a File structure with the provided path but does not
// perform any file system operations or language detection. The returned File must be
// initialized using Initialize() before it can be used for validation or processing operations.
//
// The filepath parameter can be either absolute or relative. Relative paths will be resolved
// relative to the current working directory when file operations are performed.
//
// # Example Usage
//
//	file := langutil.NewFile("main.go")
//	err := file.Initialize()
//	if err != nil {
//		log.Fatal(err)
//	}
//	language, err := file.Validate()
func NewFile(filepath string) *File {
	return &File{
		filepath: filepath,
	}
}

// LanguageAs determines the language to use for this file, with optional override.
// This method implements the language resolution logic used throughout the File type.
// If a non-empty language parameter is provided, it is returned as-is. Otherwise,
// the method returns the file's detected language if available.
//
// If no language can be determined (neither provided nor detected), the method returns
// an error with information about the file extension and suggestions for resolution.
//
// # Language Resolution Priority
//
//  1. Explicit language parameter (if non-empty)
//  2. Previously detected language from Initialize()
//  3. Error if no language can be determined
//
// This method is used internally by validation methods to support both automatic
// language detection and explicit language specification.
func (f *File) LanguageAs(language Language) (_ Language, err error) {
	if language != "" {
		goto end
	}
	if f.language != "" {
		language = f.language
	}
	if language == "" {
		// TODO: Add information about how to get add support for a language
		err = fmt.Errorf("%s is not aware of how to validate files with a file extension of '%s'", appName, filepath.Ext(f.filepath))
		goto end
	}
end:
	return language, err
}

// ensureInitialized checks that Initialize() has been called and panics if not.
// This method is used internally by all File methods that require initialization
// to provide clear error messages when the File is used incorrectly.
//
// The panic message provides clear guidance about the required initialization step,
// helping developers identify and fix usage errors quickly.
func (f *File) ensureInitialized() {
	if !f.initialized {
		panic("You must call lanutil.File.Initialized() first")
	}
}

// Initialize performs language detection and processor validation for the file.
// This method must be called after creating a File with NewFile() and before performing
// any validation or processing operations. It performs the following steps:
//
//  1. Sets the initialized flag to true
//  2. Detects the file's language based on its extension using DetectLanguage()
//  3. Validates that a processor is available for the detected language
//
// The method does not read the file contents or perform syntax validation. It only
// ensures that the file type is recognized and that appropriate processing capabilities
// are available.
//
// # Error Conditions
//
// Returns an error if:
//   - The file extension is not recognized (language detection fails)
//   - No processor is registered for the detected language
//   - The file path is invalid or inaccessible (future enhancement)
//
// # Performance
//
// This method performs language detection and processor lookup, which are both fast
// operations. It can be called repeatedly, though the results will be the same for
// a given file path.
func (f *File) Initialize() (err error) {
	f.initialized = true
	f.language = DetectLanguage(f.filepath)
	if f.language == "" {
		// TODO: Add information about how to get add support for a language
		err = fmt.Errorf("file %s is not a currently supported language", f.filepath)
		goto end
	}
end:
	return err
}

// Validate performs complete validation of the file using automatic language detection.
// This method reads the file contents from disk and validates that the file is syntactically
// correct according to the detected language's rules. It combines file I/O, language detection,
// and syntax validation into a single convenient operation.
//
// The method returns the detected language, which may be useful for callers that want to
// know what language was automatically detected for the file.
//
// # Validation Process
//
//  1. Ensures the File has been initialized
//  2. Reads the complete file contents from disk
//  3. Determines the appropriate language processor
//  4. Validates the syntax using the language-specific processor
//
// # Error Handling
//
// Returns an error if:
//   - The File has not been initialized
//   - The file cannot be read from disk
//   - Language detection or processor lookup fails
//   - The file contents contain syntax errors
//
// This method is equivalent to calling ValidateAs with an empty language parameter.
func (f *File) Validate() (language Language, err error) {
	return f.ValidateAs("")
}

// ValidateAs performs complete validation of the file using the specified or detected language.
// This method provides the core file validation functionality with support for both automatic
// language detection and explicit language specification. It reads the file from disk and
// validates the syntax using the appropriate language processor.
//
// # Language Resolution
//
// If the language parameter is non-empty, it will be used for validation. Otherwise,
// the method uses the language detected during initialization. This allows for flexible
// validation scenarios where the file extension might not accurately reflect the content type.
//
// # Validation Process
//
//  1. Ensures the File has been initialized
//  2. Reads the complete file contents from disk using os.ReadFile
//  3. Resolves the target language (explicit or detected)
//  4. Validates the syntax using ValidateSyntaxAs
//
// # File I/O
//
// The method reads the entire file into memory for validation. For very large files,
// this may have memory implications. The file contents are not cached, so repeated
// calls will re-read the file from disk.
//
// # Error Reporting
//
// File I/O errors are wrapped with additional context to help identify the source of the problem.
// Syntax errors from the language processor are returned as-is and typically include line
// number and position information.
func (f *File) ValidateAs(language Language) (_ Language, err error) {
	var content []byte

	// Read file content
	content, err = os.ReadFile(f.filepath)
	if err != nil {
		err = fmt.Errorf("failed to read file: %v", err)
		goto end
	}

	// Validate syntax
	language, err = f.ValidateSyntaxAs(string(content), language)
	if err != nil {
		goto end
	}

end:
	return language, err
}

// ValidateSyntaxAs validates the provided source content as the specified or detected language.
// This method performs syntax validation without reading from disk, making it suitable for
// validating generated or modified content before writing it to files. It provides the core
// syntax validation logic used by other validation methods.
//
// # Language Resolution
//
// The method uses the same language resolution logic as ValidateAs:
//   - If language parameter is non-empty, it is used for validation
//   - Otherwise, the file's detected language is used
//   - An error is returned if no language can be determined
//
// # Processor Selection
//
// The method automatically selects the appropriate processor for the resolved language
// and delegates syntax validation to that processor. This ensures that language-specific
// syntax rules and features are properly handled.
//
// # Performance
//
// This method is more efficient than ValidateAs for scenarios where the content is already
// in memory, as it avoids file I/O operations. It's commonly used for validating replacement
// content before performing AST-based replacements.
//
// # Example Usage
//
//	content := "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}"
//	language, err := file.ValidateSyntaxAs(content, langutil.GoLanguage)
//	if err != nil {
//		log.Printf("Syntax error: %v", err)
//	}
func (f *File) ValidateSyntaxAs(content string, language Language) (_ Language, err error) {
	var p Processor

	f.ensureInitialized()
	p, language, err = f.ProcessorAs(language)
	if err != nil {
		goto end
	}
	err = p.ValidateSyntax(content)
end:
	return language, err
}

// ValidateSyntax validates the provided source content using the file's detected language.
// This method is a convenience wrapper around ValidateSyntaxAs that uses the file's
// automatically detected language. It's useful when you want to validate content against
// the same language that was detected for the file.
//
// # Language Detection
//
// The method uses the language that was detected during Initialize(). If Initialize()
// was not called or if language detection failed, this method will return an error.
//
// # Error Handling
//
// Returns an error if:
//   - The File has not been initialized
//   - No language was detected during initialization
//   - No processor is available for the detected language
//   - The content contains syntax errors
//
// This method is equivalent to calling ValidateSyntaxAs with an empty language parameter.
func (f *File) ValidateSyntax(content string) (err error) {
	var p Processor

	f.ensureInitialized()

	p, err = f.Processor()
	if err != nil {
		goto end
	}
	err = p.ValidateSyntax(content)
end:
	return err
}

// Processor returns the processor for the file's detected language.
// This method provides access to the language-specific processor that can handle the file.
// It caches the processor instance to avoid repeated lookup operations when performing
// multiple operations on the same file.
//
// # Caching Behavior
//
// The first call to this method performs processor lookup and caches the result.
// Subsequent calls return the cached processor without additional lookup overhead.
// The cache is tied to the File instance and is not shared between File instances.
//
// # Initialization Requirement
//
// This method requires the File to be initialized, as it depends on language detection
// having been performed. Calling this method on an uninitialized File will panic.
//
// # Thread Safety
//
// The caching behavior is not thread-safe. Multiple goroutines calling this method
// concurrently on the same File instance may result in race conditions.
func (f *File) Processor() (_ Processor, err error) {
	if f.processor != nil {
		goto end
	}
	f.processor, err = GetProcessor(f.language)
end:
	return f.processor, err
}

// ProcessorAs returns the processor for the specified or detected language.
// This method provides flexible processor access with support for both automatic
// language detection and explicit language specification. It's used internally by
// validation methods that need to support language override capabilities.
//
// # Language Resolution
//
// The method uses LanguageAs to resolve the target language:
//   - If language parameter is non-empty, that language is used
//   - Otherwise, the file's detected language is used
//   - An error is returned if no language can be determined
//
// # Return Values
//
// Returns three values:
//   - processor: The processor instance for the resolved language
//   - language: The resolved language (may differ from input if auto-detected)
//   - err: Any error that occurred during language resolution or processor lookup
//
// # Processor Lookup
//
// Unlike the Processor() method, this method does not cache the result since the
// processor may vary depending on the language parameter. Each call performs a
// fresh processor lookup.
//
// # Example Usage
//
//	processor, language, err := file.ProcessorAs(langutil.GoLanguage)
//	if err != nil {
//		log.Fatal(err)
//	}
//	partInfo, err := processor.FindPart(args)
func (f *File) ProcessorAs(language Language) (processor Processor, _ Language, err error) {
	language, err = f.LanguageAs(language)
	if err != nil {
		goto end
	}
	processor, err = GetProcessor(language)
end:
	return processor, language, err
}
