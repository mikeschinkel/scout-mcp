package mcptools_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mikeschinkel/scout-mcp"
	"github.com/mikeschinkel/scout-mcp/mcptools"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const DirPrefix = scout.AppName + "-test"

type productDetectionResultOpts struct {
	AssertNoCurrentProject    bool
	AssertCurrentProject      bool
	AssertNoRecentProjects    bool
	AssertInRecentProjects    []string
	AssertNotInRecentProjects []string
	AssertRecentProjects      bool
	AssertRecentProjectsCount *int
	AssertCurrentProjectName  string
	AssertCurrentProjectDir   string
	AssertError               string // Expected error message substring
}

func requireProductDetectionResult(t *testing.T, result *mcptools.ProjectDetectionResult, err error, opts productDetectionResultOpts) {
	t.Helper()

	// Check for expected error first
	if opts.AssertError != "" {
		require.Error(t, err, "Should have error")
		if err != nil {
			assert.Contains(t, err.Error(), opts.AssertError, "Error message should contain expected text")
		}
		return // Skip other checks for error cases
	}

	// No error expected
	require.NoError(t, err, "Should not have error")

	// Verify current project is detected and valid
	if opts.AssertNoCurrentProject {
		// Verify current_project is nil since no projects found
		require.Nil(t, result.CurrentProject, "CurrentProject should be nil indicating no projects found")
	} else if opts.AssertCurrentProject {
		require.NotNil(t, result.CurrentProject, "Should detect a current project")
	}

	if result.CurrentProject != nil {
		if opts.AssertCurrentProjectName != "" {
			assert.Equal(t, opts.AssertCurrentProjectName, result.CurrentProject.Name, "Should detect correct project name")
		}

		if opts.AssertCurrentProjectDir != "" {
			assert.Equal(t, opts.AssertCurrentProjectDir, result.CurrentProject.Path, "Should have correct project path")
			// Verify the path exists and is a directory
			info, err := os.Stat(result.CurrentProject.Path)
			require.NoError(t, err, "Current project path should exist")
			assert.True(t, info.IsDir(), "Current project path should be a directory")
		}
	}

	if opts.AssertNoRecentProjects {
		assert.Len(t, result.RecentProjects, 0, "RecentProjects should be empty array")
	} else if opts.AssertRecentProjects {
		assert.Greater(t, len(result.RecentProjects), 0, "RecentProjects should not be empty array")
	}

	if len(result.RecentProjects) == 0 {
		// Should not require choice for single project
		assert.False(t, result.RequiresChoice, "Should not require choice for single project")
		return
	}
	// We have recent projects

	if opts.AssertRecentProjectsCount != nil {
		require.Len(t, result.RecentProjects,
			*opts.AssertRecentProjectsCount,
			fmt.Sprintf("Should have exactly %d other recent project(s)", *opts.AssertRecentProjectsCount),
		)
	}

	if opts.AssertInRecentProjects != nil {
		for _, p := range opts.AssertInRecentProjects {
			assert.True(t, containsMatch(p, result.RecentProjects, func(p string, pi mcptools.ProjectInfo) bool {
				return p == pi.Name
			}), fmt.Sprintf("Recent projects should contain the older project"))
		}
	}
	if opts.AssertRecentProjectsCount != nil {
		require.Len(t, result.RecentProjects,
			*opts.AssertRecentProjectsCount,
			fmt.Sprintf("Should have exactly %d other recent project(s)", *opts.AssertRecentProjectsCount),
		)
	}

	// Verify requires choice set correctly for current + recent projects
	if result.CurrentProject != nil {
		if withinTimeframe(t, result.CurrentProject.LastModified, result.RecentProjects[0].LastModified, 24*time.Hour) {
			assert.True(t, result.RequiresChoice, "Should require choice when recent projects found")
		} else {
			assert.False(t, result.RequiresChoice, "Should not require choice when current project more recent than recent projects by at least 24 hours")
		}

		// Verify recent projects does NOT include the current project (no duplication)
		assert.False(t,
			containsMatch(result.CurrentProject.Path, result.RecentProjects, func(path string, p mcptools.ProjectInfo) bool {
				return path == p.Path
			}),
			"CurrentProject path must not appear in RecentProjects",
		)
	}

	// Ensure the the recent projects are ordered reverse chronologically
	rp1 := result.RecentProjects[0]
	for i := 0; i < len(result.RecentProjects)-2; i++ {
		rp2 := result.RecentProjects[i+1]
		// Verify requires choice set correctly for current + recent projects
		assert.True(t, rp1.LastModified.After(rp2.LastModified), "Recent projects should be in reverse chronological order")
		rp1 = rp2
	}

}

func TestDetectCurrentProjectTool(t *testing.T) {
	// Get the tool and set config
	tool := mcputil.GetRegisteredTool("detect_current_project")
	require.NotNil(t, tool, "detect_current_project tool should be registered")

	t.Run("NoProjects", func(t *testing.T) {
		tf := NewTestFixture(DirPrefix)
		defer tf.Cleanup()
		// No projects added - empty directory
		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
		})

		result, err := getToolResult[mcptools.ProjectDetectionResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error with empty directory",
		)

		requireProductDetectionResult(t, result, err, productDetectionResultOpts{
			AssertNoCurrentProject: true,
			AssertNoRecentProjects: true,
		})

	})

	t.Run("SingleProjectWithGit", func(t *testing.T) {
		tf := NewTestFixture(DirPrefix)
		defer tf.Cleanup()
		pf := tf.AddProjectFixture("my-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf.AddFileFixture("README.md", FileFixtureArgs{
			Content:     "# Test Project",
			Permissions: 0644,
		})
		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
		})

		result, err := getToolResult[mcptools.ProjectDetectionResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error with single project",
		)
		requireProductDetectionResult(t, result, err, productDetectionResultOpts{
			AssertCurrentProject:     true,
			AssertNoRecentProjects:   true,
			AssertCurrentProjectName: pf.Name,
			AssertCurrentProjectDir:  pf.Dir,
		})

	})

	t.Run("SingleProjectIgnoreGit", func(t *testing.T) {
		tf := NewTestFixture(DirPrefix)
		defer tf.Cleanup()
		pf := tf.AddProjectFixture("my-project-no-git", ProjectFixtureArgs{
			HasGit:      false,
			Permissions: 0755,
		})
		pf.AddFileFixture("package.json", FileFixtureArgs{
			Content:     `{"name": "test"}`,
			Permissions: 0644,
		})
		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token":          tf.token,
			"ignore_git_requirement": true,
		})

		result, err := getToolResult[mcptools.ProjectDetectionResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error with ignore_git_requirement",
		)
		requireProductDetectionResult(t, result, err, productDetectionResultOpts{
			AssertCurrentProject:    true,
			AssertNoRecentProjects:  true,
			AssertCurrentProjectDir: pf.Dir,
		})

	})

	t.Run("MultipleProjectsClearWinner", func(t *testing.T) {
		tf := NewTestFixture(DirPrefix)
		defer tf.Cleanup()

		// Create old project (48 hours ago)
		oldPf := tf.AddProjectFixture("old-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		oldPf.AddFileFixture("README.md", FileFixtureArgs{
			Content:      "# Old Project",
			Permissions:  0644,
			ModifiedTime: time.Now().Add(-48 * time.Hour),
		})

		// Create recent project (current time)
		recentPf := tf.AddProjectFixture("recent-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		recentPf.AddFileFixture("README.md", FileFixtureArgs{
			Content:      "# Recent Project",
			Permissions:  0644,
			ModifiedTime: time.Now(),
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
		})

		result, err := getToolResult[mcptools.ProjectDetectionResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error with clear winner",
		)

		projectCount := 1
		requireProductDetectionResult(t, result, err, productDetectionResultOpts{
			AssertCurrentProject:      true,
			AssertRecentProjects:      true,
			AssertInRecentProjects:    []string{"old-project"},
			AssertRecentProjectsCount: &projectCount,
			AssertCurrentProjectName:  "recent-project",
			AssertCurrentProjectDir:   recentPf.Dir,
		})
	})

	t.Run("MultipleRecentProjects", func(t *testing.T) {
		tf := NewTestFixture(DirPrefix)
		defer tf.Cleanup()

		// Create project 1 (1 hour ago)
		pf1 := tf.AddProjectFixture("project-1", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf1.AddFileFixture("README.md", FileFixtureArgs{
			Content:      "# Project 1",
			Permissions:  0644,
			ModifiedTime: time.Now().Add(-1 * time.Hour),
		})

		// Create project 2 (2 hours ago)
		pf2 := tf.AddProjectFixture("project-2", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf2.AddFileFixture("README.md", FileFixtureArgs{
			Content:      "# Project 2",
			Permissions:  0644,
			ModifiedTime: time.Now().Add(-2 * time.Hour),
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
		})

		result, err := getToolResult[mcptools.ProjectDetectionResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error with multiple recent projects",
		)

		// For multiple recent projects (within 24 hours), there should be no current project
		// and RequiresChoice should be true
		requireProductDetectionResult(t, result, err, productDetectionResultOpts{
			AssertNoCurrentProject: true,
			AssertRecentProjects:   true,
		})

		// Additional assertion that this should require choice since projects are within 24 hours
		assert.True(t, result.RequiresChoice, "Should require choice when multiple projects are recent")
	})

	t.Run("SkipHiddenDirectories", func(t *testing.T) {
		tf := NewTestFixture(DirPrefix)
		defer tf.Cleanup()

		// Create a hidden directory (should be ignored)
		hiddenPf := tf.AddProjectFixture(".hidden-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		hiddenPf.AddFileFixture("README.md", FileFixtureArgs{
			Content:     "# Hidden Project",
			Permissions: 0644,
		})

		// Create a normal project directory
		normalPf := tf.AddProjectFixture("normal-project", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		normalPf.AddFileFixture("README.md", FileFixtureArgs{
			Content:     "# Normal Project",
			Permissions: 0644,
		})

		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
		})

		result, err := getToolResult[mcptools.ProjectDetectionResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error with hidden directories",
		)

		requireProductDetectionResult(t, result, err, productDetectionResultOpts{
			AssertCurrentProject:     true,
			AssertNoRecentProjects:   true,
			AssertCurrentProjectName: "normal-project",
			AssertCurrentProjectDir:  normalPf.Dir,
		})
	})

	t.Run("MultipleRootPathsAllowed", func(t *testing.T) {
		tf1 := NewTestFixture(DirPrefix)
		defer tf1.Cleanup()
		tf2 := NewTestFixture(DirPrefix)
		defer tf2.Cleanup()

		// Create project in first directory (most recent)
		pf1 := tf1.AddProjectFixture("project-in-dir1", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf1.AddFileFixture("README.md", FileFixtureArgs{
			Content:      "# Project 1",
			Permissions:  0644,
			ModifiedTime: time.Now(),
		})

		// Create project in second directory (older)
		pf2 := tf2.AddProjectFixture("project-in-dir2", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf2.AddFileFixture("README.md", FileFixtureArgs{
			Content:      "# Project 2",
			Permissions:  0644,
			ModifiedTime: time.Now().Add(-48 * time.Hour),
		})

		tf1.Setup(t)
		tf2.Setup(t)

		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{
				tf1.TempDir(),
				tf2.TempDir(),
			},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf1.token,
		})

		result, err := getToolResult[mcptools.ProjectDetectionResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error with multiple allowed paths",
		)

		projectCount := 1
		requireProductDetectionResult(t, result, err, productDetectionResultOpts{
			AssertCurrentProject:      true,
			AssertRecentProjects:      true,
			AssertInRecentProjects:    []string{"project-in-dir2"},
			AssertRecentProjectsCount: &projectCount,
			AssertCurrentProjectName:  "project-in-dir1",
			AssertCurrentProjectDir:   pf1.Dir,
		})
	})

	t.Run("MultipleProjectPathsAllowed", func(t *testing.T) {
		tf := NewTestFixture(DirPrefix)
		defer tf.Cleanup()

		// Create project 1 (most recent) with 5+ files
		pf1 := tf.AddProjectFixture("project-in-path1", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf1.AddFileFixtures(t, FileFixtureArgs{
			ContentFunc: func(ff *FileFixture) string {
				return fmt.Sprintf("# Project 1 - %s", ff.Name)
			},
			Permissions:  0644,
			ModifiedTime: time.Now(),
		}, "README.md", "package.json", "main.go", "config.yaml", "utils.js")

		// Create project 2 (older) with 5+ files
		pf2 := tf.AddProjectFixture("project-in-path2", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		pf2.AddFileFixtures(t, FileFixtureArgs{
			ContentFunc: func(ff *FileFixture) string {
				return fmt.Sprintf("# Project 2 - %s", ff.Name)
			},
			Permissions:  0644,
			ModifiedTime: time.Now().Add(-48 * time.Hour),
		}, "README.md", "package.json", "main.go", "config.yaml", "helpers.js")

		tf.Setup(t)

		// Configure tool with project directories as allowed paths
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{
				pf1.Dir,
				pf2.Dir,
			},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
		})

		result, err := getToolResult[mcptools.ProjectDetectionResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error with multiple project paths",
		)

		projectCount := 1
		requireProductDetectionResult(t, result, err, productDetectionResultOpts{
			AssertCurrentProject:      true,
			AssertRecentProjects:      true,
			AssertInRecentProjects:    []string{"project-in-path2"},
			AssertRecentProjectsCount: &projectCount,
			AssertCurrentProjectName:  "project-in-path1",
			AssertCurrentProjectDir:   pf1.Dir,
		})
	})

	t.Run("NoAllowedPaths", func(t *testing.T) {
		tf := NewTestFixture(DirPrefix)
		defer tf.Cleanup()
		// No projects added, but also no allowed paths configured
		tf.Setup(t)

		// Configure tool with no allowed paths
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
		})

		result, err := getToolResult[mcptools.ProjectDetectionResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should error with no allowed paths",
		)

		requireProductDetectionResult(t, result, err, productDetectionResultOpts{
			AssertError: "no allowed paths configured",
		})
	})

	t.Run("MaxProjectsParameter", func(t *testing.T) {
		tf := NewTestFixture(DirPrefix)
		defer tf.Cleanup()

		pf := tf.AddProjectFixture("project-0", ProjectFixtureArgs{
			HasGit:      true,
			Permissions: 0755,
		})
		// Create multiple files to ensure projects are robust
		pf.AddFileFixture("README.md", FileFixtureArgs{
			Content:      "# Project 0 - README.md",
			Permissions:  0644,
			ModifiedTime: time.Now(),
		})

		// Create more projects than max_projects (8 projects, limit to 3)
		// Use larger time gaps to ensure clear winner logic works
		for i := 1; i < 8; i++ {
			pf := tf.AddProjectFixture(fmt.Sprintf("project-%d", i), ProjectFixtureArgs{
				HasGit:      true,
				Permissions: 0755,
			})

			// Create multiple files to ensure projects are robust
			pf.AddFileFixtures(t, FileFixtureArgs{
				ContentFunc: func(ff *FileFixture) string {
					return fmt.Sprintf("# Project %d - %s", i, ff.Name)
				},
				Permissions:  0644,
				ModifiedTime: time.Now().Add(-time.Duration(i*25) * time.Hour), // 25+ hour gaps
			}, "README.md", "package.json", "main.go", "config.yaml")
		}
		tf.Setup(t)
		tool.SetConfig(testutil.NewMockConfig(testutil.MockConfigArgs{
			AllowedPaths: []string{tf.TempDir()},
		}))

		req := testutil.NewMockRequest(testutil.Params{
			"session_token": tf.token,
			"max_projects":  3,
		})

		result, err := getToolResult[mcptools.ProjectDetectionResult](t,
			callResult(testutil.CallTool(tool, req)),
			"Should not error with max_projects limit",
		)

		// Should have limited results to max_projects (3)
		// The most recent project should be current, others in recent projects
		maxProjects := 2 // 3 total projects: 1 current + 2 recent
		requireProductDetectionResult(t, result, err, productDetectionResultOpts{
			AssertCurrentProject:      true,
			AssertRecentProjects:      true,
			AssertRecentProjectsCount: &maxProjects,
			AssertCurrentProjectName:  "project-0", // Most recent (i=0 means most recent due to proper time gaps)
		})
	})
}
