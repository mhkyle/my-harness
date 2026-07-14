package engine

import (
	"context"
	"fmt"
	"strings"
)

type TerminalReporter struct{}

func NewTerminalReporter() *TerminalReporter {
	return &TerminalReporter{}
}

func (r *TerminalReporter) OnThinking(ctx context.Context) {
	fmt.Printf("💭 Model is thinking...\n")
}

func (r *TerminalReporter) OnToolCall(ctx context.Context, toolName, args string) {
	fmt.Printf("🔧 Tool %s called\n", toolName)
	displayArgs := strings.ReplaceAll(args, "\n", "\\n")
	displayArgs = strings.ReplaceAll(displayArgs, "\r", "\\r")
	if len(displayArgs) > 150 {
		fmt.Printf("🔧 Tool %s called with args (truncated): %s...\n", toolName, displayArgs[:150])
		displayArgs = displayArgs[:150]
	}
	fmt.Printf("Args: %s\n", displayArgs)
}

func (r *TerminalReporter) OnToolResult(ctx context.Context, toolName, result string, isError bool) {
	if isError {
		fmt.Printf("❌ Tool %s returned an error\n", toolName)
		if result != "" {
			fmt.Printf("Error: %s\n", result)
		}
	} else {
		fmt.Printf("✅ Tool %s returned a result\n", toolName)
	}
}

func (r *TerminalReporter) OnMessage(ctx context.Context, content string) {
	if content == "" {
		return
	}
	fmt.Printf("💬 Agent message: %s\n", content)
}
