package handlers

import (
	"fmt"
	"io"
)

// splitLogs separates the combined Docker logs into stdout and stderr
func splitLogs(logs io.Reader) (string, string) {
	stdout := new(string)
	stderr := new(string)
	buf := make([]byte, 8192)

	for {
		n, err := logs.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Sprintf("error reading logs: %v", err)
		}

		// Docker log format: first 8 bytes are header, then content
		if n > 8 {
			header := buf[0]
			content := string(buf[8:n])

			if header == 1 { // stdout
				*stdout += content
			} else if header == 2 { // stderr
				*stderr += content
			}
		}
	}

	return *stdout, *stderr
}
