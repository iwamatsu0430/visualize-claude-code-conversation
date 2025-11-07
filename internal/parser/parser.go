package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ParseJSONL parses a JSONL file and returns a Conversation
func ParseJSONL(filePath string) (*Conversation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var rawEntries []LogEntry
	var sessionID string
	var startTime, endTime time.Time

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024) // 10MB max token size

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue // Skip invalid lines
		}

		// Skip file-history-snapshot entries
		if entry.Type == "file-history-snapshot" {
			continue
		}

		// Extract session info
		if sessionID == "" && entry.SessionID != "" {
			sessionID = entry.SessionID
		}

		if startTime.IsZero() && !entry.Timestamp.IsZero() {
			startTime = entry.Timestamp
		}

		if !entry.Timestamp.IsZero() {
			endTime = entry.Timestamp
		}

		// Add user and assistant entries
		if entry.Type == "user" || entry.Type == "assistant" {
			rawEntries = append(rawEntries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan file: %w", err)
	}

	// Group assistant messages by message ID
	entries := groupAssistantMessages(rawEntries)

	return &Conversation{
		SessionID: sessionID,
		StartTime: startTime,
		EndTime:   endTime,
		Entries:   entries,
	}, nil
}

// groupAssistantMessages groups assistant messages with the same message ID
func groupAssistantMessages(rawEntries []LogEntry) []LogEntry {
	var entries []LogEntry
	assistantMsgMap := make(map[string]*LogEntry)

	for _, entry := range rawEntries {
		if entry.Type == "user" {
			entries = append(entries, entry)
		} else if entry.Type == "assistant" {
			messageID := entry.Message.ID

			if existing, found := assistantMsgMap[messageID]; found {
				// Merge content into existing message (if both are arrays)
				if existingArr, ok := existing.Message.Content.([]interface{}); ok {
					if entryArr, ok := entry.Message.Content.([]interface{}); ok {
						existing.Message.Content = append(existingArr, entryArr...)
					}
				}
			} else {
				// Create new entry
				newEntry := entry
				assistantMsgMap[messageID] = &newEntry
				entries = append(entries, newEntry)
			}
		}
	}

	return entries
}
