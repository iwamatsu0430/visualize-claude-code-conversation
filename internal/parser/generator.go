package parser

import (
	_ "embed"
	"encoding/json"
	"fmt"
	htmlpkg "html"
	"strings"
	"time"
)

//go:embed assets/styles.css
var cssContent string

//go:embed assets/script.js
var jsContent string

// GenerateHTML generates the complete HTML output
func GenerateHTML(conv *Conversation) (string, error) {
	messagesHTML := generateMessages(conv)
	tocHTML := generateTOC(conv)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ja">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Claude Code Conversation</title>
  <style>%s</style>
</head>
<body>
  %s
  <div class="container">
    <div class="header">
      <h1>Claude Code Conversation</h1>
      <div class="session-info">
        <div>Session ID: %s</div>
        <div>Start: %s</div>
        <div>End: %s</div>
      </div>
    </div>
    <div class="messages">
      %s
    </div>
  </div>
  <script>%s</script>
</body>
</html>`, cssContent, tocHTML, htmlpkg.EscapeString(conv.SessionID),
		formatTimestamp(conv.StartTime),
		formatTimestamp(conv.EndTime),
		messagesHTML, jsContent), nil
}

func formatTimestamp(t time.Time) string {
	// Format: 2025/10/30 15:14:49
	return t.In(time.Local).Format("2006/01/02 15:04:05")
}

func generateMessages(conv *Conversation) string {
	var html strings.Builder
	userMessageIDs := collectUserMessageIDs(conv.Entries)

	userMsgIndex := 0
	for i, entry := range conv.Entries {
		var nextEntry *LogEntry
		if i+1 < len(conv.Entries) {
			nextEntry = &conv.Entries[i+1]
		}

		if entry.Type == "user" {
			isSessionContinuation := entry.IsVisibleInTranscriptOnly && entry.IsCompactSummary

			if isSessionContinuation {
				html.WriteString(generateSessionContinuationHTML(&entry))
			} else {
				messageID := fmt.Sprintf("user-msg-%d", userMsgIndex)
				messageNumber := userMsgIndex + 1
				hasPrev := userMsgIndex > 0
				hasNext := userMsgIndex < len(userMessageIDs)-1

				msgHTML := generateUserMessageHTML(&entry, messageID, messageNumber, hasPrev, hasNext)
				if msgHTML != "" {
					html.WriteString(msgHTML)
					userMsgIndex++
				}
			}
		} else if entry.Type == "assistant" {
			html.WriteString(generateAssistantMessageHTML(&entry, nextEntry))
		}
	}

	return html.String()
}

func collectUserMessageIDs(entries []LogEntry) []string {
	var ids []string
	userMsgIndex := 0

	for _, entry := range entries {
		if entry.Type != "user" {
			continue
		}

		isSessionContinuation := entry.IsVisibleInTranscriptOnly && entry.IsCompactSummary
		if isSessionContinuation {
			continue
		}

		messageText := extractMessageText(&entry)
		if strings.TrimSpace(messageText) != "" {
			ids = append(ids, fmt.Sprintf("user-msg-%d", userMsgIndex))
			userMsgIndex++
		}
	}

	return ids
}

func extractMessageText(entry *LogEntry) string {
	// Handle string content
	if str, ok := entry.Message.Content.(string); ok {
		return str
	}

	// Handle array content
	var parts []string
	if contentArray, ok := entry.Message.Content.([]interface{}); ok {
		for _, item := range contentArray {
			if contentMap, ok := item.(map[string]interface{}); ok {
				contentType, _ := contentMap["type"].(string)

				if contentType == "text" {
					if text, ok := contentMap["text"].(string); ok {
						parts = append(parts, text)
					}
				} else if contentType != "tool_result" {
					if content, ok := contentMap["content"].(string); ok && content != "" {
						parts = append(parts, content)
					}
				}
			}
		}
	}

	return strings.Join(parts, "\n")
}

func generateUserMessageHTML(entry *LogEntry, messageID string, messageNumber int, hasPrev, hasNext bool) string {
	messageText := extractMessageText(entry)

	if strings.TrimSpace(messageText) == "" {
		return ""
	}

	var html strings.Builder

	// Divider
	html.WriteString(fmt.Sprintf(`<div class="message-divider">#%d</div>`, messageNumber))

	// Message group
	html.WriteString(`<div class="message-group user">`)
	html.WriteString(fmt.Sprintf(`<div class="timestamp-label">%s</div>`, formatTimestamp(entry.Timestamp)))
	html.WriteString(fmt.Sprintf(`<div id="%s" class="message user-message">`, messageID))
	html.WriteString(fmt.Sprintf(`<div class="message-content">%s</div>`, htmlpkg.EscapeString(messageText)))

	// Navigation
	html.WriteString(`<div class="message-navigation">`)
	if hasPrev {
		html.WriteString(`<button class="nav-btn" onclick="jumpToMessage(this, 'prev')">⬆️</button>`)
	}
	if hasNext {
		html.WriteString(`<button class="nav-btn" onclick="jumpToMessage(this, 'next')">⬇️</button>`)
	}
	html.WriteString(`</div>`) // message-navigation
	html.WriteString(`</div>`) // message
	html.WriteString(`</div>`) // message-group

	return html.String()
}

func generateAssistantMessageHTML(entry *LogEntry, nextEntry *LogEntry) string {
	var thinkingParts []string
	var textParts []string
	var toolUses []ToolUse

	// Parse assistant content
	if contentArray, ok := entry.Message.Content.([]interface{}); ok {
		for _, item := range contentArray {
			if contentMap, ok := item.(map[string]interface{}); ok {
				contentType, _ := contentMap["type"].(string)

				switch contentType {
				case "thinking":
					if t, ok := contentMap["thinking"].(string); ok {
						thinkingParts = append(thinkingParts, t)
					}
				case "text":
					if text, ok := contentMap["text"].(string); ok {
						textParts = append(textParts, text)
					}
				case "tool_use":
					toolUse := ContentItem{
						Type: "tool_use",
					}
					if id, ok := contentMap["id"].(string); ok {
						toolUse.ID = id
					}
					if name, ok := contentMap["name"].(string); ok {
						toolUse.Name = name
					}
					if input, ok := contentMap["input"]; ok {
						toolUse.Input = input
					}
					toolUses = append(toolUses, ToolUse{Tool: toolUse})
				}
			}
		}
	}

	// Find tool results in next entry
	if nextEntry != nil && nextEntry.Type == "user" {
		if nextContentArray, ok := nextEntry.Message.Content.([]interface{}); ok {
			for _, item := range nextContentArray {
				if resultMap, ok := item.(map[string]interface{}); ok {
					if resultType, ok := resultMap["type"].(string); ok && resultType == "tool_result" {
						toolUseID, _ := resultMap["tool_use_id"].(string)
						resultContent, _ := resultMap["content"].(string)

						for i := range toolUses {
							if toolUses[i].Tool.ID == toolUseID {
								toolUses[i].Result = resultContent
								break
							}
						}
					}
				}
			}
		}
	}

	messageText := strings.Join(textParts, "\n")
	thinking := strings.Join(thinkingParts, "\n\n---\n\n")
	hasThinking := len(thinkingParts) > 0
	hasTools := len(toolUses) > 0

	var html strings.Builder

	// Thinking section (show first if present)
	if hasThinking {
		html.WriteString(`<div class="message-group assistant">`)
		html.WriteString(fmt.Sprintf(`<div class="timestamp-label">%s</div>`, formatTimestamp(entry.Timestamp)))
		html.WriteString(`<div class="thinking-section">`)
		html.WriteString(`<button class="meta-btn" onclick="toggleThinking(this)">🧠 ...</button>`)
		html.WriteString(`<div class="thinking-content" style="display: none;">`)
		html.WriteString(fmt.Sprintf(`<pre>%s</pre>`, htmlpkg.EscapeString(thinking)))
		html.WriteString(`</div></div></div>`)
	}

	// Message box
	if messageText != "" {
		html.WriteString(`<div class="message-group assistant">`)
		html.WriteString(fmt.Sprintf(`<div class="timestamp-label">%s</div>`, formatTimestamp(entry.Timestamp)))
		html.WriteString(`<div class="message assistant-message">`)
		html.WriteString(fmt.Sprintf(`<div class="message-content">%s</div>`, htmlpkg.EscapeString(messageText)))
		html.WriteString(`</div></div>`)
	}

	// Tools section
	if hasTools {
		html.WriteString(`<div class="message-group assistant">`)
		html.WriteString(fmt.Sprintf(`<div class="timestamp-label">%s</div>`, formatTimestamp(entry.Timestamp)))
		html.WriteString(`<div class="tools-section">⚒️ `)

		toolHTMLs := make([]string, len(toolUses))
		for i, toolUse := range toolUses {
			toolHTMLs[i] = generateToolUseHTML(toolUse)
		}
		html.WriteString(strings.Join(toolHTMLs, ", "))
		html.WriteString(`</div></div>`)
	}

	return html.String()
}

func generateToolUseHTML(toolUse ToolUse) string {
	inputJSON, _ := json.MarshalIndent(toolUse.Tool.Input, "", "  ")

	var html strings.Builder
	html.WriteString(`<span class="tool-item" onclick="toggleToolDetails(event)">`)
	html.WriteString(htmlpkg.EscapeString(toolUse.Tool.Name))
	html.WriteString(` ▼<div class="tool-details" style="display: none;">`)
	html.WriteString(`<div class="tool-section"><div class="tool-section-title">Input Parameters:</div>`)
	html.WriteString(fmt.Sprintf(`<pre class="tool-input">%s</pre></div>`, htmlpkg.EscapeString(string(inputJSON))))

	if toolUse.Result != "" {
		html.WriteString(`<div class="tool-section"><div class="tool-section-title">Result:</div>`)
		html.WriteString(fmt.Sprintf(`<pre class="tool-result">%s</pre></div>`, htmlpkg.EscapeString(toolUse.Result)))
	}

	html.WriteString(`</div></span>`)
	return html.String()
}

func generateSessionContinuationHTML(entry *LogEntry) string {
	messageText := extractMessageText(entry)

	return fmt.Sprintf(`
<div id="session-continuation" class="message session-continuation-message">
  <div class="session-continuation-header">
    <span class="session-continuation-icon">⚠️</span>
    <span class="session-continuation-title">Session Continued</span>
    <span class="timestamp">%s</span>
  </div>
  <div class="session-continuation-notice">
    This session was continued from a previous conversation that ran out of context.
  </div>
  <div class="session-continuation-toggle">
    <button class="toggle-btn" onclick="toggleSessionSummary(this)">📋 View conversation summary</button>
  </div>
  <div class="session-continuation-content" style="display: none;">
    <pre>%s</pre>
  </div>
</div>`, formatTimestamp(entry.Timestamp), htmlpkg.EscapeString(messageText))
}

func generateTOC(conv *Conversation) string {
	var tocItems []string
	userMsgIndex := 0

	for _, entry := range conv.Entries {
		if entry.Type != "user" {
			continue
		}

		isSessionContinuation := entry.IsVisibleInTranscriptOnly && entry.IsCompactSummary
		if isSessionContinuation {
			continue
		}

		messageText := extractMessageText(&entry)
		if strings.TrimSpace(messageText) == "" {
			continue
		}

		lines := strings.Split(messageText, "\n")
		preview := lines[0]
		if len(preview) > 60 {
			preview = preview[:60] + "..."
		}

		messageID := fmt.Sprintf("user-msg-%d", userMsgIndex)
		tocItems = append(tocItems, fmt.Sprintf(`
<div class="toc-item" onclick="scrollToMessage('%s')">
  <div class="toc-number">#%d</div>
  <div class="toc-content">
    <div class="toc-preview">%s</div>
    <div class="toc-timestamp">%s</div>
  </div>
</div>`, messageID, userMsgIndex+1, htmlpkg.EscapeString(preview), formatTimestamp(entry.Timestamp)))

		userMsgIndex++
	}

	return fmt.Sprintf(`
<div class="sidebar" id="sidebar">
  <div class="toc">
    %s
  </div>
</div>
<button class="scroll-top-btn" id="scrollTopBtn" onclick="scrollToTop()">↑ Top</button>
`, strings.Join(tocItems, ""))
}
