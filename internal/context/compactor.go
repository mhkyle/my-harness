package context

import (
	"fmt"
	"log"

	"mhkyle/my-harness/internal/schema"
)

type Compactor interface {
	Compact(messages []schema.Message) []schema.Message
}

// StaticCompactor removes the earlier messages to make sure the total length of the context is less than maxChars, and keeps the last retainLastMsgs messages
type StaticCompactor struct {
	MaxChars       int
	RetainLastMsgs int
}

func NewStaticCompactor(maxChars, retainLastMsgs int) *StaticCompactor {
	return &StaticCompactor{
		MaxChars:       maxChars,
		RetainLastMsgs: retainLastMsgs,
	}
}

func (c *StaticCompactor) Compact(messages []schema.Message) []schema.Message {
	currentLen := c.estimateLength(messages)

	// If the current length is within the limit, return the original messages
	if currentLen < c.MaxChars {
		return messages
	}

	log.Printf("[Compactor] Current context length %d exceeds max %d, compacting...\n", currentLen, c.MaxChars)

	var compacted []schema.Message
	msgCount := len(messages)

	protectStartIndex := msgCount - c.RetainLastMsgs
	if protectStartIndex < 0 {
		protectStartIndex = 0
	}

	for i, msg := range messages {
		if msg.Role == schema.RoleSystem {
			compacted = append(compacted, msg)
			continue
		}

		newMsg := msg
		isInWorkingMemory := i >= protectStartIndex

		if msg.Role == schema.RoleUser && msg.ToolCallID != "" {

			// for earier user messages, just remove all the contents for long messages
			if !isInWorkingMemory {
				if len(newMsg.Content) > 200 {
					newMsg.Content = fmt.Sprintf("... [In order to save space, earlier user message content has been truncated] ...")
				}
			} else {
				// for latest user messages, only keep the head and tail of the content, and truncate the middle part if it's too long
				const maxKeep = 1000
				if len(msg.Content) > maxKeep {
					head := msg.Content[:maxKeep/2]
					tail := msg.Content[len(msg.Content)-maxKeep/2:]
					newMsg.Content = fmt.Sprintf("%s\n\n...[Contents too long, %d characters in the middle have been truncated]...\n\n%s", head, len(msg.Content)-maxKeep, tail)
				}
			}
		} else if msg.Role == schema.RoleAssistant && msg.Content != "" {
			if !isInWorkingMemory && len(msg.Content) > 200 {
				newMsg.Content = fmt.Sprintf("... [In order to save space, earlier assistant message content has been truncated] ...")
			}
		}
		compacted = append(compacted, newMsg)
	}

	newLength := c.estimateLength(compacted)
	log.Printf("[Compactor] Compaction complete. New context length: %d (was %d)\n", newLength, currentLen)

	return compacted
}

func (c *StaticCompactor) estimateLength(messages []schema.Message) int {
	totalLen := 0
	for _, msg := range messages {
		totalLen += len(msg.Content)
		for _, tc := range msg.ToolCalls {
			totalLen += len(tc.Name) + len(tc.Arguments)

		}
	}
	return totalLen
}
