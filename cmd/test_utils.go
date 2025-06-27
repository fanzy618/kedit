
package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// executeCommandC captures the output of a command by redirecting stdout and stderr.
func executeCommandC(t *testing.T, args ...string) string {
	t.Helper()

	// Keep old stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Create pipes to capture stdout and stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Reset command for a clean run
	rootCmd.SetArgs(nil)
	// Set the arguments for the root command.
	rootCmd.SetArgs(args)

	// Execute the command.
	_ = rootCmd.Execute()
	// assert.NoError(t, err) // Do not assert error here, as some tests expect errors

	// Close the writers and restore stdout and stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read the output from the pipes
	var bufOut, bufErr bytes.Buffer
	_, errOut := io.Copy(&bufOut, rOut)
	_, errErr := io.Copy(&bufErr, rErr)
	assert.NoError(t, errOut)
	assert.NoError(t, errErr)

	// Return the captured output, combining stdout and stderr, and trimming any extra space
	return strings.TrimSpace(bufOut.String() + bufErr.String())
}
