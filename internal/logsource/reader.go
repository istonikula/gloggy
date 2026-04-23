package logsource

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// ReadFile reads all log entries from the file at path.
// Returns an error if the file cannot be opened.
// Entries are returned in order with 1-based line numbers.
func ReadFile(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	defer f.Close()
	return scanEntries(f)
}

func scanEntries(r io.Reader) ([]Entry, error) {
	var entries []Entry
	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		// Copy since scanner reuses its buffer.
		lineCopy := make([]byte, len(line))
		copy(lineCopy, line)
		switch Classify(lineCopy) {
		case LineTypeJSONL:
			entries = append(entries, ParseJSONL(lineCopy, lineNum))
		default:
			entries = append(entries, NewRawEntry(lineCopy, lineNum))
		}
	}
	if err := scanner.Err(); err != nil {
		return entries, fmt.Errorf("scan entries: %w", err)
	}
	return entries, nil
}
