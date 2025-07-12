package scout

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func mustClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func validatePath(path string) (err error) {
	var absPath string
	var pathInfo os.FileInfo

	absPath, err = filepath.Abs(path)
	if err != nil {
		err = fmt.Errorf("invalid path '%s': %v", path, err)
		goto end
	}

	pathInfo, err = os.Stat(absPath)
	if err != nil {
		err = fmt.Errorf("path '%s' does not exist: %v", absPath, err)
		goto end
	}

	if !pathInfo.IsDir() {
		err = fmt.Errorf("path '%s' is not a directory", absPath)
		goto end
	}

end:
	return err
}

// Add this utility function (probably in util.go)
func homeRelativePath(path string) string {
	var homeDir string
	var err error

	homeDir, err = os.UserHomeDir()
	if err != nil {
		return path // Return original path if we can't get home dir
	}

	// Check if path starts with home directory
	if strings.HasPrefix(path, homeDir) {
		return "~" + path[len(homeDir):]
	}

	return path
}
