package scoutcfg

import (
	"io"
)

// mustClose ensures that an io.Closer is properly closed and logs any
// errors that occur during the close operation. This function is designed
// to be used in defer statements where close errors cannot be easily
// returned to the caller but should still be logged for debugging and
// monitoring purposes.
//
// The function provides consistent error handling for file close operations
// throughout the scoutcfg package. While close errors are relatively rare
// in practice, they can indicate serious issues such as:
//   - Filesystem corruption or hardware problems
//   - Network issues when writing to network-mounted filesystems
//   - Resource exhaustion or quota exceeded conditions
//   - Interrupted system calls during cleanup
//
// Parameters:
//   - c: Any object implementing io.Closer interface, typically *os.File
//     instances created during file operations. The closer may be nil,
//     in which case this function is a no-op.
//
// The function logs close errors using the package logger at ERROR level,
// providing visibility into potential filesystem or resource issues that
// might otherwise go unnoticed.
//
// Usage pattern:
//
//	file, err := os.Create(filename)
//	if err != nil {
//		return err
//	}
//	defer mustClose(file)
//
// This function requires that SetLogger has been called to configure
// the package logger. If no logger is configured, ensureLogger will
// panic with an appropriate error message.
//
// Note: This function is designed for use in defer statements and does
// not return errors. For cases where close errors need to be handled
// explicitly, use direct Close() calls with proper error checking.
func mustClose(c io.Closer) {
	if c == nil {
		return
	}
	
	err := c.Close()
	if err != nil {
		ensureLogger()
		logger.Error("Error closing resource", "error", err)
	}
}