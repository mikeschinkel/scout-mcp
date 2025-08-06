package mcputil

import (
	"fmt"
	"os"
)

func WriteFile(c Config, filePath string, content string) (err error) {

	if !c.IsAllowedPath(filePath) {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	err = os.WriteFile(filePath, []byte(content), 0644)

end:
	return err
}

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
