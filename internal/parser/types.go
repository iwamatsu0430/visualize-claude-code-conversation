package parser

import "time"

// Message represents a conversation message
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Can be string or []ContentItem
	ID      string      `json:"id,omitempty"`
	Type    string      `json:"type,omitempty"`
	Model   string      `json:"model,omitempty"`
}

// ContentItem represents a content item in a message
type ContentItem struct {
	Type        string `json:"type"`
	Text        string `json:"text,omitempty"`
	Thinking    string `json:"thinking,omitempty"`
	Content     string `json:"content,omitempty"`
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Input       any    `json:"input,omitempty"`
	ToolUseID   string `json:"tool_use_id,omitempty"`
}

// LogEntry represents a single JSONL entry
type LogEntry struct {
	Type                     string    `json:"type"`
	UUID                     string    `json:"uuid"`
	SessionID                string    `json:"sessionId"`
	Timestamp                time.Time `json:"timestamp"`
	Message                  Message   `json:"message"`
	IsVisibleInTranscriptOnly bool      `json:"isVisibleInTranscriptOnly,omitempty"`
	IsCompactSummary         bool      `json:"isCompactSummary,omitempty"`
}

// Conversation represents a parsed conversation
type Conversation struct {
	SessionID string
	StartTime time.Time
	EndTime   time.Time
	Entries   []LogEntry
}

// ToolUse represents a tool usage with its result
type ToolUse struct {
	Tool   ContentItem
	Result string
}
