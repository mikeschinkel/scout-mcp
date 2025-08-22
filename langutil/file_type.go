package langutil

// FileType represents a file type definition that includes both name and extension components.
// This structure is used to define comprehensive file type recognition that goes beyond
// simple extension matching. Some files are identified by specific names (like "Makefile"),
// others by extensions (like ".go"), and some by both (like "package.json").
//
// The FileType system supports complex file identification patterns commonly found in
// software development projects, where configuration files, build scripts, and project
// metadata files often have specific naming conventions that include both the base name
// and extension.
//
// # Usage Patterns
//
// The FileType structure supports several identification patterns:
//   - Extension-only: {Ext: ".go"} matches any .go file
//   - Name-only: {Name: "Makefile"} matches files named exactly "Makefile"
//   - Name+Extension: {Name: "package", Ext: ".json"} matches "package.json"
//
// This flexibility allows for precise file type detection in complex project structures
// where different file types may share extensions but serve different purposes.
type FileType struct {
	Name string // Base filename (without extension) for files with specific names
	Ext  string // File extension including the leading dot (e.g., ".go", ".json")
}

// GetFileTypes returns a comprehensive list of all supported file types across multiple programming languages.
// This function provides access to the complete registry of file types that the langutil package
// recognizes and can potentially process. The list includes source code files, configuration files,
// build scripts, dependency manifests, and documentation files for various programming ecosystems.
//
// The returned slice is a copy of the internal registry and can be safely modified by callers
// without affecting the internal state of the package.
//
// # File Type Categories
//
// The returned file types are organized into several categories:
//   - Source code files: .go, .js, .py, .php, etc.
//   - Project configuration: package.json, composer.json, pyproject.toml, etc.
//   - Build and dependency files: Makefile, Dockerfile, requirements.txt, etc.
//   - Documentation and metadata: README.md, LICENSE, CHANGELOG.md, etc.
//   - Version control and tooling: .gitignore, .editorconfig, .prettierrc, etc.
//
// # Usage Examples
//
// This function is commonly used for:
//   - File filtering in development tools
//   - Project analysis and discovery
//   - IDE and editor file type associations
//   - Build system file inclusion/exclusion rules
//
// Example usage:
//
//	fileTypes := langutil.GetFileTypes()
//	for _, ft := range fileTypes {
//		fmt.Printf("File type: %s\n", ft.String())
//	}
func GetFileTypes() []FileType {
	return fileTypes
}

// fileTypes contains the comprehensive registry of supported file types across multiple programming languages and ecosystems.
// This registry is used by various parts of the langutil package to identify and categorize files during processing operations.
//
// The registry is organized by programming language and project type, covering the most common file types encountered
// in software development projects. It includes source files, configuration files, build scripts, dependency manifests,
// and tooling configuration files for major programming languages and frameworks.
var fileTypes = []FileType{
	// Go project files - Go programming language ecosystem
	{Ext: ".go"},       // Go source files
	{Ext: ".mod"},      // Go module definition files (go.mod)
	{Ext: ".sum"},      // Go module checksum files (go.sum)
	{Ext: ".work"},     // Go workspace files (go.work)
	{Ext: ".work.sum"}, // Go workspace checksum files (go.work.sum)

	// JavaScript/TypeScript project files - Node.js and web development ecosystem
	{Ext: ".js"},                          // JavaScript source files
	{Ext: ".mjs"},                         // ECMAScript module files
	{Ext: ".cjs"},                         // CommonJS module files
	{Ext: ".jsx"},                         // React JavaScript files
	{Ext: ".ts"},                          // TypeScript source files
	{Ext: ".tsx"},                         // React TypeScript files
	{Ext: ".json"},                        // Generic JSON files
	{Name: "package", Ext: ".json"},       // NPM package manifest
	{Name: "package-lock", Ext: ".json"},  // NPM lock file
	{Name: "yarn", Ext: ".lock"},          // Yarn lock file
	{Name: "pnpm-lock", Ext: ".yaml"},     // PNPM lock file
	{Name: "tsconfig", Ext: ".json"},      // TypeScript configuration
	{Name: "jsconfig", Ext: ".json"},      // JavaScript configuration
	{Name: ".eslintrc", Ext: ".js"},       // ESLint configuration (JavaScript)
	{Name: ".eslintrc", Ext: ".json"},     // ESLint configuration (JSON)
	{Name: "webpack", Ext: ".config.js"},  // Webpack configuration
	{Name: "vite", Ext: ".config.js"},     // Vite configuration
	{Name: "next", Ext: ".config.js"},     // Next.js configuration
	{Name: "tailwind", Ext: ".config.js"}, // Tailwind CSS configuration

	// Python project files - Python development ecosystem
	{Ext: ".py"},                            // Python source files
	{Ext: ".pyi"},                           // Python stub files
	{Ext: ".pyw"},                           // Python Windows files
	{Ext: ".pyx"},                           // Cython source files
	{Ext: ".pxd"},                           // Cython definition files
	{Ext: ".pxi"},                           // Cython include files
	{Name: "requirements", Ext: ".txt"},     // Pip requirements file
	{Name: "requirements-dev", Ext: ".txt"}, // Development requirements
	{Name: "Pipfile"},                       // Pipenv configuration
	{Name: "Pipfile", Ext: ".lock"},         // Pipenv lock file
	{Name: "pyproject", Ext: ".toml"},       // Python project configuration
	{Name: "setup", Ext: ".py"},             // Python setup script
	{Name: "setup", Ext: ".cfg"},            // Python setup configuration
	{Name: "tox", Ext: ".ini"},              // Tox testing configuration
	{Name: "pytest", Ext: ".ini"},           // Pytest configuration
	{Name: ".python-version"},               // Python version specification
	{Ext: ".toml"},                          // TOML configuration files
	{Ext: ".cfg"},                           // Configuration files
	{Ext: ".ini"},                           // INI configuration files

	// PHP project files - PHP development ecosystem
	{Ext: ".php"},                       // PHP source files
	{Ext: ".phar"},                      // PHP Archive files
	{Name: "composer", Ext: ".json"},    // Composer dependency manifest
	{Name: "composer", Ext: ".lock"},    // Composer lock file
	{Name: "phpunit", Ext: ".xml"},      // PHPUnit configuration
	{Name: "phpunit", Ext: ".xml.dist"}, // PHPUnit distribution configuration
	{Name: ".php-version"},              // PHP version specification

	// Salesforce DX project files - Salesforce development ecosystem
	{Ext: ".cls"},                        // Apex classes
	{Ext: ".trigger"},                    // Apex triggers
	{Ext: ".page"},                       // Visualforce pages
	{Ext: ".component"},                  // Visualforce components
	{Ext: ".app"},                        // Lightning applications
	{Ext: ".evt"},                        // Lightning events
	{Ext: ".intf"},                       // Lightning interfaces
	{Ext: ".tokens"},                     // Lightning design tokens
	{Ext: ".auradoc"},                    // Aura documentation
	{Ext: ".cmp"},                        // Aura components
	{Ext: ".css"},                        // CSS stylesheets
	{Ext: ".design"},                     // Lightning design files
	{Ext: ".svg"},                        // SVG graphics
	{Ext: ".lwc"},                        // Lightning Web Components
	{Ext: ".html"},                       // HTML files
	{Ext: ".xml"},                        // XML files
	{Name: "sfdx-project", Ext: ".json"}, // Salesforce DX project configuration
	{Name: ".foreign"},                   // Salesforce foreign file markers
	{Name: "package", Ext: ".xml"},       // Salesforce package manifest
	{Name: "destructiveChanges", Ext: ".xml"},     // Salesforce destructive changes
	{Name: "destructiveChangesPre", Ext: ".xml"},  // Salesforce pre-deployment destructive changes
	{Name: "destructiveChangesPost", Ext: ".xml"}, // Salesforce post-deployment destructive changes

	// Common configuration and documentation files - Cross-language project files
	{Ext: ".md"},                           // Markdown documentation
	{Ext: ".txt"},                          // Plain text files
	{Ext: ".log"},                          // Log files
	{Ext: ".yml"},                          // YAML configuration files
	{Ext: ".yaml"},                         // YAML configuration files (alternative extension)
	{Ext: ".env"},                          // Environment variable files
	{Ext: ".example"},                      // Example/template files
	{Ext: ".sample"},                       // Sample files
	{Name: "Makefile"},                     // Make build scripts
	{Name: "Dockerfile"},                   // Docker container definitions
	{Name: "docker-compose", Ext: ".yml"},  // Docker Compose configuration (YAML)
	{Name: "docker-compose", Ext: ".yaml"}, // Docker Compose configuration (YAML alternative)
	{Name: ".gitignore"},                   // Git ignore rules
	{Name: ".gitattributes"},               // Git attributes configuration
	{Name: "README", Ext: ".md"},           // Project README documentation
	{Name: "CHANGELOG", Ext: ".md"},        // Project changelog
	{Name: "LICENSE"},                      // License files (no extension)
	{Name: "LICENSE", Ext: ".txt"},         // License files (with .txt extension)
	{Name: ".env"},                         // Environment variables (root)
	{Name: ".env", Ext: ".example"},        // Environment variables example
	{Name: ".env", Ext: ".local"},          // Local environment variables
	{Name: ".env", Ext: ".development"},    // Development environment variables
	{Name: ".env", Ext: ".production"},     // Production environment variables
	{Name: ".editorconfig"},                // Editor configuration
	{Name: ".prettierrc"},                  // Prettier formatter configuration
	{Name: ".prettierrc", Ext: ".json"},    // Prettier configuration (JSON)
	{Name: ".prettierignore"},              // Prettier ignore rules
}

// String returns a string representation of the FileType.
// This method provides a convenient way to convert a FileType into a human-readable string
// that represents the complete file pattern. The method handles the different FileType patterns:
//
//   - Extension-only types return just the extension (e.g., ".go")
//   - Name-only types return just the name (e.g., "Makefile")
//   - Name+extension types return the concatenated result (e.g., "package.json")
//
// # Usage
//
// This method is commonly used for:
//   - Display purposes in user interfaces
//   - Debug output and logging
//   - File pattern matching in tools
//   - Converting FileType definitions to glob patterns
//
// # Example Output
//
//	{Ext: ".go"}.String()                    // Returns: ".go"
//	{Name: "Makefile"}.String()              // Returns: "Makefile"
//	{Name: "package", Ext: ".json"}.String() // Returns: "package.json"
//
// The method handles edge cases gracefully and always returns a non-empty string
// representing the file type pattern.
func (f FileType) String() string {
	switch {
	case f.Ext == "":
		return f.Name
	case f.Name == "":
		return f.Ext
	}
	return f.Name + f.Ext
}
