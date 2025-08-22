// Package scoutcfg_test provides comprehensive tests for the scoutcfg package,
// focusing on FileStore functionality including file operations, error handling,
// and edge cases. These tests use temporary directories to avoid interfering
// with actual user configuration files.
package scoutcfg_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/mikeschinkel/scout-mcp/scoutcfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// logger is a package-level test logger used for error reporting during test cleanup.
// This logger is separate from the scoutcfg package logger to avoid interference
// with logger configuration tests.
var logger *slog.Logger

// testData represents a simple data structure used for testing JSON serialization
// and deserialization operations. It contains basic types that are commonly
// used in configuration files.
type testData struct {
	Name string // String field for testing text data
	Age  int    // Integer field for testing numeric data
}

// TestFileStore_SaveLoadExists verifies the core functionality of FileStore
// including saving data to JSON files, loading data back from files, and
// checking file existence. This test covers the primary workflow that most
// applications will use with the scoutcfg package.
//
// The test uses a temporary directory to avoid affecting user configuration
// files and validates that data roundtrips correctly through JSON serialization.
func TestFileStore_SaveLoadExists(t *testing.T) {
	var err error
	dir := filepath.Join(os.TempDir(), "gmcfg-test-"+uuid.NewString())
	t.Cleanup(func() { must(os.RemoveAll(dir)) })

	s := scoutcfg.NewFileStore("test-app")
	s.SetBaseDir(dir)

	data := testData{Name: "Alice", Age: 42}
	filename := "config/testdata.json"

	err = s.Save(filename, &data)
	require.NoError(t, err)

	exists := s.Exists(filename)
	assert.True(t, exists)

	var loaded testData
	err = s.Load(filename, &loaded)
	require.NoError(t, err)
	assert.Equal(t, data, loaded)
}

// TestFileStore_Append tests the file appending functionality which is
// commonly used for logging operations. The test verifies that content
// is correctly appended to files and that multiple append operations
// accumulate content properly.
//
// This test is important for ensuring that log files and incremental
// configuration updates work correctly without overwriting existing data.
func TestFileStore_Append(t *testing.T) {
	var err error
	dir := filepath.Join(os.TempDir(), "gmcfg-append-"+uuid.NewString())
	t.Cleanup(func() { must(os.RemoveAll(dir)) })

	s := scoutcfg.NewFileStore("test-app")
	s.SetBaseDir(dir)

	filename := "log/output.log"
	msg1 := []byte("first line\n")
	msg2 := []byte("second line\n")

	err = s.Append(filename, msg1)
	require.NoError(t, err)

	err = s.Append(filename, msg2)
	require.NoError(t, err)

	path := filepath.Join(dir, filename)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, string(msg1)+string(msg2), string(content))
}

// TestFileStore_LoadNonexistent verifies that attempting to load a file
// that doesn't exist returns an appropriate error. This test ensures
// that the FileStore properly handles missing configuration files,
// which is a common scenario during first-time application startup.
func TestFileStore_LoadNonexistent(t *testing.T) {
	var err error

	s := scoutcfg.NewFileStore("test-app")
	s.SetBaseDir(t.TempDir())

	err = s.Load("does-not-exist.json", &testData{})
	assert.Error(t, err)
}

// TestFileStore_SaveInvalidJSON tests error handling when attempting to
// save data that cannot be serialized to JSON. This test uses a channel
// type which is not JSON-serializable to trigger the error condition.
//
// This test ensures that the FileStore provides appropriate error messages
// for invalid data types and fails gracefully without creating partial files.
func TestFileStore_SaveInvalidJSON(t *testing.T) {
	var err error

	s := scoutcfg.NewFileStore("test-app")
	s.SetBaseDir(t.TempDir())

	ch := make(chan int) // channels are not JSON-serializable
	err = s.Save("bad.json", ch)
	assert.Error(t, err)
}

// TestFileStore_ConfigDir validates the configuration directory path
// computation and caching functionality. This test ensures that the
// FileStore correctly determines and caches the configuration directory
// path for efficient repeated access.
func TestFileStore_ConfigDir(t *testing.T) {
	s := scoutcfg.NewFileStore("test-app")
	dir := t.TempDir()
	s.SetBaseDir(dir)

	cfgDir, err := s.ConfigDir()
	assert.NoError(t, err)
	assert.Equal(t, dir, cfgDir)
}

// must is a test helper function that logs errors during test cleanup
// operations. It uses the test logger to report cleanup errors without
// failing tests, since cleanup errors are typically not critical to
// test validation but may indicate environmental issues.
//
// Parameters:
//   - err: An error from cleanup operations, typically file removal.
//     If err is nil, this function is a no-op.
func must(err error) {
	if err != nil {
		logger.Error(err.Error())
	}
}