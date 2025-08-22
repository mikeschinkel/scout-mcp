package fsfix

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type FileFixture struct {
	Filepath       string
	Name           string
	Content        string
	Permissions    int
	DirPermissions int
	ModifiedTime   time.Time
	Missing        bool
	Pending        bool
	Parent         Fixture
}

type FileFixtureArgs struct {
	Name           string
	Content        string
	ContentFunc    func(ff *FileFixture) string
	ModifiedTime   time.Time
	Permissions    int
	DirPermissions int
	Missing        bool
	Pending        bool
	Parent         Fixture
}

func NewFileFixture(name string, args *FileFixtureArgs) *FileFixture {
	if args == nil {
		args = &FileFixtureArgs{}
	}
	if args.Permissions == 0 {
		args.Permissions = 0644
	}
	if args.DirPermissions == 0 {
		args.DirPermissions = 0755
	}
	return &FileFixture{
		Name:           name,
		Content:        args.Content,
		Permissions:    args.Permissions,
		DirPermissions: args.DirPermissions,
		ModifiedTime:   args.ModifiedTime,
		Missing:        args.Missing,
		Pending:        args.Pending,
		Parent:         args.Parent,
	}
}

func (ff *FileFixture) Setup(t *testing.T, pf Fixture) {
	t.Helper()
	ff.Parent = pf
	ff.Filepath = filepath.Join(pf.Dir(), ff.Name)
	ff.createFile(t)
}

// createFile handles the common file creation logic
func (ff *FileFixture) createFile(t *testing.T) {
	var err error
	t.Helper()

	// Skip file creation if it's marked as Missing or Pending
	if ff.Missing || ff.Pending {
		goto end
	}

	require.NotEqual(t, 0, ff.Permissions, "File permissions not set for", ff.Filepath)

	err = os.MkdirAll(filepath.Dir(ff.Filepath), os.FileMode(ff.DirPermissions))
	require.NoError(t, err, "Failed to create test file directory", filepath.Dir(ff.Filepath))
	err = os.WriteFile(ff.Filepath, []byte(ff.Content), os.FileMode(ff.Permissions))
	require.NoError(t, err, "Failed to create test file", ff.Filepath)

	// Set modification time if specified
	if !ff.ModifiedTime.IsZero() {
		err = os.Chtimes(ff.Filepath, ff.ModifiedTime, ff.ModifiedTime)
		require.NoError(t, err, "Failed to set file modification time", ff.Filepath)
	}
end:
}
