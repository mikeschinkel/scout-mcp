package scout

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// mustClose closes an io.Closer and terminates the program on error.
func mustClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// validatePath validates that the given path exists and is a directory.
// Returns an error if the path is invalid, doesn't exist, or is not a directory.
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

// toExistenceMap takes a []comparable and returns a map[comparable]struct{}
func toExistenceMap[K comparable](s []K) (m map[K]struct{}) {
	m = make(map[K]struct{}, len(s))
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}
