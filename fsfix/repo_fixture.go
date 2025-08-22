package fsfix

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var _ Fixture = (*RepoFixture)(nil)

// RepoFixture represents a project directory fixture with optional Git repository.
type RepoFixture struct {
	*DirFixture
	HasGit NilableBool // Whether to create a .git directory
}

func (pf *RepoFixture) SetupWithParent(t *testing.T, parent Fixture) {
	t.Helper()
	pf.DirFixture.SetupWithParent(t, parent)

	// Create .git directory to make it a valid repo
	gitDir := filepath.Join(pf.dir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	require.NoError(t, err, "Failed to create .git directory")
}

// RepoFixtureArgs contains arguments for creating a RepoFixture.
type RepoFixtureArgs struct {
	HasGit       NilableBool    // Whether to create a .git directory
	Files        []*FileFixture // Files to create within this project
	Permissions  int            // Directory permissions
	ModifiedTime time.Time      // Modification time for the directory
	Parent       Fixture        // Parent test fixture
}

func NewRepoFixture(name string, args *RepoFixtureArgs) *RepoFixture {
	if args == nil {
		args = &RepoFixtureArgs{}
	}
	if args.Permissions == 0 {
		args.Permissions = 0755
	}
	return &RepoFixture{
		DirFixture: NewDirFixture(name, &DirFixtureArgs{
			ModifiedTime: args.ModifiedTime,
			Permissions:  args.Permissions,
			Parent:       args.Parent,
		}),
	}
}

func (pf *RepoFixture) MakeDir(fp string) string {
	return filepath.Join(pf.dir, fp)
}

func (pf *RepoFixture) AddRepoFixture(name string, args *RepoFixtureArgs) *RepoFixture {
	rf := NewRepoFixture(filepath.Join(pf.dir, pf.Name, name), args)
	rf.Parent = pf
	pf.ChildFixtures = append(pf.ChildFixtures, rf)
	return rf
}

func (pf *RepoFixture) AddDirFixture(name string, args *DirFixtureArgs) *DirFixture {
	rf := NewDirFixture(filepath.Join(pf.dir, pf.Name, name), args)
	rf.Parent = pf
	pf.ChildFixtures = append(pf.ChildFixtures, rf)
	return rf
}

// AddFileFixture adds a file fixture to a project fixture
func (pf *RepoFixture) AddFileFixture(name string, args *FileFixtureArgs) *FileFixture {
	ff := NewFileFixture(name, args)
	ff.Parent = pf
	pf.FileFixtures = append(pf.FileFixtures, ff)
	return ff
}

// AddFileFixtures adds multiple files at once using defaults if as
// FileFixtureArgs when one of args is passed just as a string(string) and it
// gets its content from ContentFunc, or a FileFixtureArgs is passed which much
// include Name.
func (pf *RepoFixture) AddFileFixtures(t *testing.T, defaults *FileFixtureArgs, args ...any) {
	var ff *FileFixture
	for _, f := range args {
		switch ffa := f.(type) {
		case string:
			ff = pf.AddFileFixture(ffa, defaults)
			if defaults.ContentFunc != nil {
				ff.Content = defaults.ContentFunc(ff)
			}
		case *FileFixtureArgs:
			if ffa.Name == "" {
				t.Fatalf("Name not set for file fixure being added to project fixture '%s'", pf.Name)
			}
			pf.AddFileFixture(ffa.Name, ffa)
		default:
			t.Fatalf("Invalid type '%T' passed for file fixure being added to project fixture: '%v'", f, f)
		}
	}
}
