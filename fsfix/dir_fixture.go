package fsfix

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// NilableBool allows for nil, true, or false values in fixture configurations.
type NilableBool any

var _ Fixture = (*DirFixture)(nil)

// DirFixture represents a dir directory fixture with optional Git repository.
type DirFixture struct {
	Name          string         // Name of the dir directory
	HasGit        NilableBool    // Whether to create a .git directory
	FileFixtures  []*FileFixture // Files to create within this dir
	ChildFixtures []Fixture      // Subdirectories or Projects to create within this dir
	ModifiedTime  time.Time      // Modification time for the dir directory
	Permissions   int            // Directory permissions (e.g., 0755)
	dir           string         // Full path to the created directory
	Parent        Fixture        // Parent test fixture
}

func (df *DirFixture) Dir() string {
	return df.dir
}

// DirFixtureArgs contains arguments for creating a DirFixture.
type DirFixtureArgs struct {
	Files        []*FileFixture // Files to create within this dir
	Permissions  int            // Directory permissions
	ModifiedTime time.Time      // Modification time for the directory
	Parent       Fixture        // Parent test fixture
}

func NewDirFixture(name string, args *DirFixtureArgs) *DirFixture {
	if args == nil {
		args = &DirFixtureArgs{}
	}
	if args.Permissions == 0 {
		args.Permissions = 0755
	}
	return &DirFixture{
		Name:         name,
		Parent:       args.Parent,
		FileFixtures: args.Files,
		ModifiedTime: args.ModifiedTime,
		Permissions:  args.Permissions,
	}
}

func (df *DirFixture) MakeDir(fp string) string {
	return filepath.Join(df.dir, fp)
}

func (df *DirFixture) SetupWithParent(t *testing.T, pf Fixture) {
	t.Helper()

	// Create a single dir directory with .git
	df.dir = filepath.Join(pf.Dir(), df.Name)
	require.NotEqual(t, 0, df.Permissions, "File permissions not set for", df.dir)
	err := os.MkdirAll(df.dir, os.FileMode(df.Permissions))
	require.NoError(t, err, "Failed to create dir directory for test:", df.dir)
	for _, file := range df.FileFixtures {
		file.Setup(t, df)
	}
	for _, child := range df.ChildFixtures {
		child.SetupWithParent(t, df)
	}
}

func (df *DirFixture) AddDirFixture(name string, args *DirFixtureArgs) *DirFixture {
	cf := NewDirFixture(filepath.Join(df.dir, df.Name, name), args)
	cf.Parent = df
	df.ChildFixtures = append(df.ChildFixtures, cf)
	return cf
}

func (df *DirFixture) AddRepoFixture(name string, args *RepoFixtureArgs) *RepoFixture {
	cf := NewRepoFixture(filepath.Join(df.dir, df.Name, name), args)
	cf.Parent = df
	df.ChildFixtures = append(df.ChildFixtures, cf)
	return cf
}

// AddFileFixture adds a file fixture to a dir fixture
func (df *DirFixture) AddFileFixture(name string, args *FileFixtureArgs) *FileFixture {
	ff := NewFileFixture(name, args)
	ff.Parent = df
	df.FileFixtures = append(df.FileFixtures, ff)
	return ff
}

// AddFileFixtures adds multiple files at once using defaults if as
// FileFixtureArgs when one of args is passed just as a string(string) and it
// gets its content from ContentFunc, or a FileFixtureArgs is passed which much
// include Name.
func (df *DirFixture) AddFileFixtures(t *testing.T, defaults *FileFixtureArgs, args ...any) {
	var ff *FileFixture
	for _, f := range args {
		switch ffa := f.(type) {
		case string:
			ff = df.AddFileFixture(ffa, defaults)
			if defaults.ContentFunc != nil {
				ff.Content = defaults.ContentFunc(ff)
			}
		case *FileFixtureArgs:
			if ffa.Name == "" {
				t.Fatalf("Name not set for file fixure being added to dir fixture '%s'", df.Name)
			}
			df.AddFileFixture(ffa.Name, ffa)
		default:
			t.Fatalf("Invalid type '%T' passed for file fixure being added to dir fixture: '%v'", f, f)
		}
	}
}
