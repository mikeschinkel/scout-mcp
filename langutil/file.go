package langutil

import (
	"fmt"
	"os"
	"path/filepath"
)

type File struct {
	filepath    string
	language    Language
	processor   Processor
	initialized bool
}

func NewFile(filepath string) *File {
	return &File{
		filepath: filepath,
	}
}

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

func (f *File) ensureInitialized() {
	if !f.initialized {
		panic("You must call lanutil.File.Initialized() first")
	}
}

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

func (f *File) Validate() (language Language, err error) {
	return f.ValidateAs("")
}

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

func (f *File) Processor() (_ Processor, err error) {
	if f.processor != nil {
		goto end
	}
	f.processor, err = GetProcessor(f.language)
end:
	return f.processor, err
}

func (f *File) ProcessorAs(language Language) (processor Processor, _ Language, err error) {
	language, err = f.LanguageAs(language)
	if err != nil {
		goto end
	}
	processor, err = GetProcessor(language)
end:
	return processor, language, err
}
