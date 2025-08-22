// Package scoutcfg_test provides test setup and initialization for the
// scoutcfg package test suite. This file contains the TestMain function
// which coordinates test execution and ensures proper test environment
// configuration.
package scoutcfg_test

import (
	"os"
	"testing"

	"github.com/mikeschinkel/scout-mcp/scoutcfg"
	"github.com/mikeschinkel/scout-mcp/testutil"
)

// TestMain is the entry point for all tests in the scoutcfg_test package.
// It sets up the test environment by configuring the scoutcfg package
// logger with a test-appropriate logger instance, then runs all tests
// and performs any necessary cleanup.
//
// The function ensures that the scoutcfg package has a properly configured
// logger before any tests run, preventing panics from ensureLogger() calls
// during test execution. This is essential because FileStore operations
// require a configured logger for error reporting.
//
// Test setup includes:
//   - Configuring the scoutcfg package logger with testutil.NewTestLogger()
//   - Running the complete test suite via m.Run()
//   - Performing any necessary cleanup after tests complete
//
// The function follows Go testing conventions by calling os.Exit with
// the test result code, ensuring that test failures are properly reported
// to the test runner and CI systems.
//
// Parameters:
//   - m: The testing.M instance provided by the Go test runner, containing
//     all test functions to be executed and test configuration options.
func TestMain(m *testing.M) {
	// Setup: Configure the scoutcfg package logger to prevent panics
	// during test execution. The test logger provides appropriate
	// output formatting and error handling for test environments.
	scoutcfg.SetLogger(testutil.NewTestLogger())

	// Run all tests in the package. The test runner will execute each
	// test function and collect results, returning an exit code that
	// indicates overall success or failure.
	code := m.Run()

	// Cleanup: Currently no cleanup is required, but this section
	// provides a location for future cleanup operations such as
	// removing temporary files or shutting down test services.

	// Exit with the test result code to properly report test outcomes
	// to the test runner, CI systems, and command-line users.
	os.Exit(code)
}