// Package fsfix provides testing utilities including file fixtures and mock configurations.
// It supports creating temporary directories, files, and project structures for testing.
package fsfix

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var _ Fixture = (*RootFixture)(nil)

// RootFixture manages temporary directories and files for testing purposes.
type RootFixture struct {
	DirPrefix     string         // Prefix for temporary directory names
	tempDir       string         // Path to the temporary directory
	FileFixtures  []*FileFixture // File-level fixtures in the root temp directory
	ChildFixtures []Fixture      // Project-level fixtures (directories with .git)
	cleanupFunc   func()         // Function to clean up resources
}

func (rf *RootFixture) Dir() string {
	return rf.tempDir
}

func (rf *RootFixture) SetupWithParent(*testing.T, Fixture) {
	panic("SetupWithParent is not relevant as RootFixture should be the root")
}

func (rf *RootFixture) Setup(t *testing.T) {
	t.Helper()

	// Create temp directory (this can fail, so it belongs in Setup)
	var err error
	rf.tempDir, err = os.MkdirTemp("", rf.DirPrefix+"-*")
	require.NoError(t, err, "Failed to create temp directory")

	rf.cleanupFunc = func() {
		must(os.RemoveAll(rf.tempDir))
	}

	// Set up all the project fixtures
	// rf.RemoveFiles(t) // BUG: This removes the directory we just created
	for _, cf := range rf.ChildFixtures {
		cf.SetupWithParent(t, rf)
	}

	// Set up all the test fixture files (directly in temp directory)
	for _, ff := range rf.FileFixtures {
		ff.Setup(t, rf)
	}
}

// NewRootFixture creates a new TestFixture with the specified directory prefix.
func NewRootFixture(dirPrefix string) *RootFixture {
	return &RootFixture{
		DirPrefix:     dirPrefix,
		FileFixtures:  []*FileFixture{},
		ChildFixtures: []Fixture{},
	}
}

// AddRepoFixture adds a project-level fixture (directory with .git) to the TestFixture.
func (rf *RootFixture) AddRepoFixture(name string, args *RepoFixtureArgs) *RepoFixture {
	pf := NewRepoFixture(name, args)
	pf.Parent = rf
	rf.ChildFixtures = append(rf.ChildFixtures, pf)
	return pf
}

// AddDirFixture adds a directory fixture (directory with optional .git) to the TestFixture.
func (rf *RootFixture) AddDirFixture(name string, args *DirFixtureArgs) *DirFixture {
	df := NewDirFixture(name, args)
	df.Parent = rf
	rf.ChildFixtures = append(rf.ChildFixtures, df)
	return df
}

// AddFileFixture adds a file fixture directly to the TestFixture temp directory
func (rf *RootFixture) AddFileFixture(name string, args *FileFixtureArgs) *FileFixture {
	ff := NewFileFixture(name, args)
	ff.Parent = rf
	rf.FileFixtures = append(rf.FileFixtures, ff)
	return ff
}

func (rf *RootFixture) TempDir() string {
	return rf.tempDir
}

func (rf *RootFixture) Cleanup() {
	rf.cleanupFunc()
}

func (rf *RootFixture) RemoveFiles(t *testing.T) {
	var err error
	t.Helper()
	if rf.tempDir == "" {
		goto end
	}
	if rf.tempDir == "/" {
		goto end
	}
	if len(rf.tempDir) <= len("/tmp/x") {
		goto end
	}
	err = os.RemoveAll(rf.tempDir)
	if err != nil {
		t.Fatalf("failed to remove temporary files: %s", err.Error())
	}
end:
}
