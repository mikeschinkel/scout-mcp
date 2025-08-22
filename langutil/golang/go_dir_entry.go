package golang

import (
	"io/fs"
	"os"
	"path/filepath"
)

var _ os.DirEntry = (*goDirEntry)(nil)

// goDirEntry DOES NOT validate to ensure the oath is a directory, it
// assumes that it is.
type goDirEntry struct {
	path string
}

func (g goDirEntry) Name() string {
	return filepath.Base(g.path)
}

func (g goDirEntry) IsDir() bool {
	return true
}

func (g goDirEntry) Type() fs.FileMode {
	return 0755
}

func (g goDirEntry) Info() (fs.FileInfo, error) {
	return os.Stat(g.path)
}

// newGoDirEntry creates a new goDirEntry for the specified directory path
func newGoDirEntry(path string) *goDirEntry {
	return &goDirEntry{path: path}
}
