package handlers

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// splitLogs separates combined Docker logs into stdout and stderr.
func splitLogs(logs io.Reader) (string, string) {
	var stdout, stderr strings.Builder
	reader := bufio.NewReader(logs) // Use bufio.Reader to handler partial reads and incomplete headers
	headerBuf := make([]byte, 8)    // Temporary buffer for header storage

	for {
		// Read the 8-byte header
		_, err := io.ReadFull(reader, headerBuf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Sprintf("error reading logs: %v", err)
		}

		// Extract header info
		header := headerBuf[0] // First byte determines stdout/stderr

		// Read the actual log message
		content, err := reader.ReadBytes('\n') // Read until newline
		if err != nil && err != io.EOF {
			return "", fmt.Sprintf("error reading log content: %v", err)
		}

		if header == 1 {
			stdout.Write(content)
		} else if header == 2 {
			stderr.Write(content)
		}
	}

	return stdout.String(), stderr.String()
}
