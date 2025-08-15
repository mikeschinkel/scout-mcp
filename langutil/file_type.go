package langutil

type FileType struct {
	Name string
	Ext  string
}

func GetFileTypes() []FileType {
	return fileTypes
}

var fileTypes = []FileType{
	// Go project files
	{Ext: ".go"},
	{Ext: ".mod"},
	{Ext: ".sum"},
	{Ext: ".work"},
	{Ext: ".work.sum"},

	// JavaScript/TypeScript project files
	{Ext: ".js"},
	{Ext: ".mjs"},
	{Ext: ".cjs"},
	{Ext: ".jsx"},
	{Ext: ".ts"},
	{Ext: ".tsx"},
	{Ext: ".json"},
	{Name: "package", Ext: ".json"},
	{Name: "package-lock", Ext: ".json"},
	{Name: "yarn", Ext: ".lock"},
	{Name: "pnpm-lock", Ext: ".yaml"},
	{Name: "tsconfig", Ext: ".json"},
	{Name: "jsconfig", Ext: ".json"},
	{Name: ".eslintrc", Ext: ".js"},
	{Name: ".eslintrc", Ext: ".json"},
	{Name: "webpack", Ext: ".config.js"},
	{Name: "vite", Ext: ".config.js"},
	{Name: "next", Ext: ".config.js"},
	{Name: "tailwind", Ext: ".config.js"},

	// Python project files
	{Ext: ".py"},
	{Ext: ".pyi"},
	{Ext: ".pyw"},
	{Ext: ".pyx"},
	{Ext: ".pxd"},
	{Ext: ".pxi"},
	{Name: "requirements", Ext: ".txt"},
	{Name: "requirements-dev", Ext: ".txt"},
	{Name: "Pipfile"},
	{Name: "Pipfile", Ext: ".lock"},
	{Name: "pyproject", Ext: ".toml"},
	{Name: "setup", Ext: ".py"},
	{Name: "setup", Ext: ".cfg"},
	{Name: "tox", Ext: ".ini"},
	{Name: "pytest", Ext: ".ini"},
	{Name: ".python-version"},
	{Ext: ".toml"},
	{Ext: ".cfg"},
	{Ext: ".ini"},

	// PHP project files
	{Ext: ".php"},
	{Ext: ".phar"},
	{Name: "composer", Ext: ".json"},
	{Name: "composer", Ext: ".lock"},
	{Name: "phpunit", Ext: ".xml"},
	{Name: "phpunit", Ext: ".xml.dist"},
	{Name: ".php-version"},

	// Salesforce DX project files
	{Ext: ".cls"},
	{Ext: ".trigger"},
	{Ext: ".page"},
	{Ext: ".component"},
	{Ext: ".app"},
	{Ext: ".evt"},
	{Ext: ".intf"},
	{Ext: ".tokens"},
	{Ext: ".auradoc"},
	{Ext: ".cmp"},
	{Ext: ".css"},
	{Ext: ".design"},
	{Ext: ".svg"},
	{Ext: ".lwc"},
	{Ext: ".html"},
	{Ext: ".xml"},
	{Name: "sfdx-project", Ext: ".json"},
	{Name: ".foreign"},
	{Name: "package", Ext: ".xml"},
	{Name: "destructiveChanges", Ext: ".xml"},
	{Name: "destructiveChangesPre", Ext: ".xml"},
	{Name: "destructiveChangesPost", Ext: ".xml"},

	// Common configuration and documentation files
	{Ext: ".md"},
	{Ext: ".txt"},
	{Ext: ".log"},
	{Ext: ".yml"},
	{Ext: ".yaml"},
	{Ext: ".env"},
	{Ext: ".example"},
	{Ext: ".sample"},
	{Name: "Makefile"},
	{Name: "Dockerfile"},
	{Name: "docker-compose", Ext: ".yml"},
	{Name: "docker-compose", Ext: ".yaml"},
	{Name: ".gitignore"},
	{Name: ".gitattributes"},
	{Name: "README", Ext: ".md"},
	{Name: "CHANGELOG", Ext: ".md"},
	{Name: "LICENSE"},
	{Name: "LICENSE", Ext: ".txt"},
	{Name: ".env"},
	{Name: ".env", Ext: ".example"},
	{Name: ".env", Ext: ".local"},
	{Name: ".env", Ext: ".development"},
	{Name: ".env", Ext: ".production"},
	{Name: ".editorconfig"},
	{Name: ".prettierrc"},
	{Name: ".prettierrc", Ext: ".json"},
	{Name: ".prettierignore"},
}

func (f FileType) String() string {
	switch {
	case f.Ext == "":
		return f.Name
	case f.Name == "":
		return f.Ext
	}
	return f.Name + f.Ext
}
