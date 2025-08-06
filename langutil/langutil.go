package langutil

import (
	"errors"
)

type Args struct {
	AppName string
}

var appName string

func Initialize(args Args) error {
	appName = args.AppName
	return nil
}

func ValidateFileAs(fp string, language Language) (_ Language, err error) {
	f := NewFile(fp)
	err = f.Initialize()
	if err != nil {
		goto end
	}
	language, err = f.ValidateAs(language)
	if err != nil {
		goto end
	}
end:
	return language, err
}

type ValidationResult struct {
	FilePath string
	Language Language
	Error    error
}

func ValidateFilesAs(filepaths []string, language Language) (results []ValidationResult, _ error) {
	var errs []error
	results = make([]ValidationResult, 0, len(filepaths))
	// Validate each file
	for _, fp := range filepaths {
		lang, err := ValidateFileAs(fp, language)
		errs = append(errs, err)

		results = append(results, ValidationResult{
			FilePath: fp,
			Language: lang,
			Error:    err,
		})
	}
	return results, errors.Join(errs...)
}
