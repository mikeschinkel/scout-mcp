package testutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mikeschinkel/scout-mcp/mcptools"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/stretchr/testify/require"
)

type TestFixture struct {
	DirPrefix       string
	tempDir         string
	ProjectFixtures []*ProjectFixture
	FileFixtures    []*FileFixture
	cleanupFunc     func()
}

//tool.SetConfig(testutil.NewMockConfig(allowedPaths))

func NewTestFixture(dirPrefix string) *TestFixture {
	return &TestFixture{
		DirPrefix:       dirPrefix,
		ProjectFixtures: []*ProjectFixture{},
		FileFixtures:    []*FileFixture{},
	}
}

func (tf *TestFixture) AddProjectFixture(name string, args ProjectFixtureArgs) *ProjectFixture {
	args.TestFixture = tf
	pf := NewProjectFixture(name, args)
	tf.ProjectFixtures = append(tf.ProjectFixtures, pf)
	return pf
}

// AddFileFixture adds a file fixture directly to the TestFixture temp directory
func (tf *TestFixture) AddFileFixture(name string, args FileFixtureArgs) *FileFixture {
	args.TestFixture = tf
	ff := NewFileFixture(name, args)
	tf.FileFixtures = append(tf.FileFixtures, ff)
	return ff
}

func (tf *TestFixture) Setup(t *testing.T) {
	t.Helper()

	logger := QuietLogger()
	mcptools.SetLogger(logger)
	mcputil.SetLogger(logger)

	// Create temp directory (this can fail, so it belongs in Setup)
	var err error
	tf.tempDir, err = os.MkdirTemp("", tf.DirPrefix+"-*")
	require.NoError(t, err, "Failed to create temp directory")

	tf.cleanupFunc = func() {
		must(os.RemoveAll(tf.tempDir))
	}

	// Set up all the project fixtures
	tf.RemoveFiles(t)
	for _, pf := range tf.ProjectFixtures {
		pf.Setup(t, tf)
	}

	// Set up all the test fixture files (directly in temp directory)
	for _, ff := range tf.FileFixtures {
		ff.SetupInTestFixture(t, tf)
	}
}

func (tf *TestFixture) TempDir() string {
	return tf.tempDir
}

func (tf *TestFixture) Cleanup() {
	tf.cleanupFunc()
}

func (tf *TestFixture) RemoveFiles(t *testing.T) {
	var err error
	t.Helper()
	if tf.tempDir == "" {
		goto end
	}
	if tf.tempDir == "/" {
		goto end
	}
	if len(tf.tempDir) <= len("/tmp/x") {
		goto end
	}
	err = os.RemoveAll(tf.tempDir)
	if err != nil {
		t.Fatalf("failed to remove temporary files: %s", err.Error())
	}
end:
}

type ProjectFixture struct {
	Name         string
	HasGit       NilableBool
	FileFixtures []*FileFixture
	ModifiedTime time.Time
	Permissions  int
	Dir          string
	TestFixture  *TestFixture
}

type NilableBool any

type ProjectFixtureArgs struct {
	HasGit       NilableBool
	Files        []*FileFixture
	Permissions  int
	ModifiedTime time.Time
	TestFixture  *TestFixture
}

func NewProjectFixture(name string, args ProjectFixtureArgs) *ProjectFixture {
	return &ProjectFixture{
		TestFixture:  args.TestFixture,
		Name:         name,
		HasGit:       args.HasGit,
		FileFixtures: args.Files,
		ModifiedTime: args.ModifiedTime,
		Permissions:  args.Permissions,
	}
}

func (pf *ProjectFixture) Setup(t *testing.T, tf *TestFixture) {
	t.Helper()

	// Create a single project directory with .git
	pf.Dir = filepath.Join(tf.tempDir, pf.Name)
	require.NotEqual(t, 0, pf.Permissions, "File permissions not set for", pf.Dir)
	err := os.MkdirAll(pf.Dir, os.FileMode(pf.Permissions))
	require.NoError(t, err, "Failed to create project directory for test:", pf.Dir)
	if pf.HasGit.(bool) {
		// Create .git directory to make it a valid project
		gitDir := filepath.Join(pf.Dir, ".git")
		err = os.MkdirAll(gitDir, 0755)
		require.NoError(t, err, "Failed to create .git directory")
	}
	for _, file := range pf.FileFixtures {
		file.Setup(t, pf)
	}
}

type FileFixture struct {
	Filepath       string
	Name           string
	Content        string
	Permissions    int
	DirPermissions int
	ModifiedTime   time.Time
	Missing        bool
	Pending        bool
	ProjectFixture *ProjectFixture
	TestFixture    *TestFixture
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
	ProjectFixture *ProjectFixture
	TestFixture    *TestFixture
}

func NewFileFixture(name string, args FileFixtureArgs) *FileFixture {
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
		ProjectFixture: args.ProjectFixture,
		TestFixture:    args.TestFixture,
	}
}

func (ff *FileFixture) Setup(t *testing.T, pf *ProjectFixture) {
	t.Helper()
	ff.ProjectFixture = pf
	ff.Filepath = filepath.Join(pf.Dir, string(ff.Name))
	ff.createFile(t)
}

// SetupInTestFixture sets up a file fixture directly in the TestFixture temp directory
func (ff *FileFixture) SetupInTestFixture(t *testing.T, tf *TestFixture) {
	t.Helper()
	ff.TestFixture = tf
	ff.Filepath = filepath.Join(tf.tempDir, string(ff.Name))
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

// AddFileFixture adds a file fixture to a project fixture
func (pf *ProjectFixture) AddFileFixture(name string, args FileFixtureArgs) *FileFixture {
	args.ProjectFixture = pf
	ff := NewFileFixture(name, args)
	pf.FileFixtures = append(pf.FileFixtures, ff)
	return ff
}

// AddFileFixtures adds multiple files at once using defaults if as
// FileFixtureArgs when one of args is passed just as a string(string) and it
// gets its content from ContentFunc, or a FileFixtureArgs is passed which much
// include Name.
func (pf *ProjectFixture) AddFileFixtures(t *testing.T, defaults FileFixtureArgs, args ...any) {
	var ff *FileFixture
	for _, f := range args {
		switch ffa := f.(type) {
		case string:
			ff = pf.AddFileFixture(ffa, defaults)
			if defaults.ContentFunc != nil {
				ff.Content = defaults.ContentFunc(ff)
			}
		case FileFixtureArgs:
			if ffa.Name == "" {
				t.Fatalf("Name not set for file fixure being added to project fixture '%s'", pf.Name)
			}
			pf.AddFileFixture(ffa.Name, ffa)
		default:
			t.Fatalf("Invalid type '%T' passed for file fixure being added to project fixture: '%v'", f, f)
		}
	}
}
