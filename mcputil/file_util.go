package mcputil

import (
	"fmt"
	"os"
)

// WriteFile writes content to a file after validating the path is allowed.
// This function provides secure file writing with path validation against the server's
// allowed paths configuration to prevent unauthorized file system access.
func WriteFile(c Config, filePath string, content string) (err error) {

	if !c.IsAllowedPath(filePath) {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	err = os.WriteFile(filePath, []byte(content), 0644)

end:
	return err
}

// ReadFile reads content from a file after validating the path is allowed.
// This function provides secure file reading with path validation against the server's
// allowed paths configuration to prevent unauthorized file system access.
func ReadFile(c Config, filePath string) (content string, err error) {
	var fileData []byte

	if !c.IsAllowedPath(filePath) {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	fileData, err = os.ReadFile(filePath)
	if err != nil {
		goto end
	}

	content = string(fileData)

end:
	return content, err
}
