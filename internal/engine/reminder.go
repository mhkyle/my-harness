package engine

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"mhkyle/my-harness/internal/schema"
)

const (
	failureThreshold = 3
)

// Reminder is an interface that defines the behavior of a reminder prompt injector.
// It uses maxTurns to limit the number of turns in the loop, or monitoring the token usage budget;
type Reminder interface {
	CheckAndInject(schema.ToolCall, schema.ToolResult) *schema.Message
}

// ReminderInjector will track the number of consecutive failures for each tool call and inject a reminder message into the session when the failure threshold is reached.
type ReminderInjector struct {
	consecutiveFailures map[string]int
}

func NewReminderInjector() Reminder {
	return &ReminderInjector{
		consecutiveFailures: make(map[string]int),
	}
}

func generateFingerprint(toolName string, args []byte) string {
	hasher := md5.New()
	hasher.Write([]byte(toolName))
	hasher.Write(args)
	return hex.EncodeToString(hasher.Sum(nil))
}

func (r *ReminderInjector) CheckAndInject(lastToolCall schema.ToolCall, lastResult schema.ToolResult) *schema.Message {
	fingerprint := generateFingerprint(lastToolCall.Name, lastToolCall.Arguments)

	if !lastResult.IsError {
		r.consecutiveFailures = make(map[string]int)
		return nil
	}

	r.consecutiveFailures[fingerprint]++
	failureCount := r.consecutiveFailures[fingerprint]

	log.Printf("Reminder: Tool %s failed %d consecutive times with fingerprint %s", lastToolCall.Name, failureCount, fingerprint)

	if failureCount >= failureThreshold {
		log.Println("Reminder: Failure threshold reached. Injecting reminder message.")
		return &schema.Message{
			Role: schema.RoleUser,
			Content: fmt.Sprintf("[SYSTEM REMINDER WARNING]: The tool '%s' has failed %d consecutive times. Please STOP the useless loop. You need to \n"+
				"1. Stop guessing blindly and check the tool's arguments and the current state of the system.\n"+
				"2. Change your approach or strategy to avoid repeated failures.\n"+
				"3. If you cannot resolve the issue by current system tools, you should STOP the process and seek assistance from a human operator.\n", lastToolCall.Name, failureCount),
		}
	}
	return nil
}
