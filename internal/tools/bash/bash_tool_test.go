package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mhkyle/my-harness/internal/schema"
)

func TestBashToolExecuteSuccess(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	result := tool.Execute(context.Background(), schema.ToolCall{
		ID:        "call_success",
		Name:      "bash",
		Arguments: []byte(`{"command":"pwd"}`),
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", result.Output)
	}
	if result.ToolCallID != "call_success" {
		t.Fatalf("unexpected tool call id: %s", result.ToolCallID)
	}
	if strings.TrimSpace(result.Output) == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestBashToolExecuteInvalidJSON(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	result := tool.Execute(context.Background(), schema.ToolCall{
		ID:        "call_invalid_json",
		Name:      "bash",
		Arguments: []byte(`{"command":`),
	})

	if !result.IsError {
		t.Fatal("expected error for invalid json")
	}
	if !strings.Contains(result.Output, "invalid arguments") {
		t.Fatalf("unexpected error output: %s", result.Output)
	}
}

func TestBashToolExecuteMissingCommand(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	result := tool.Execute(context.Background(), schema.ToolCall{
		ID:        "call_missing_command",
		Name:      "bash",
		Arguments: []byte(`{"command":"   "}`),
	})

	if !result.IsError {
		t.Fatal("expected error for empty command")
	}
	if result.Output != "command is required" {
		t.Fatalf("unexpected output: %s", result.Output)
	}
}

func TestBashToolExecuteCommandFailure(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	result := tool.Execute(context.Background(), schema.ToolCall{
		ID:        "call_failure",
		Name:      "bash",
		Arguments: []byte(`{"command":"ls /definitely-not-existing-path"}`),
	})

	if !result.IsError {
		t.Fatal("expected command failure to return error")
	}
	if !strings.Contains(result.Output, "No such file or directory") {
		t.Fatalf("unexpected failure output: %s", result.Output)
	}
}

func TestBashToolExecuteWithWorkDir(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "work")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	tool := &BashTool{}
	result := tool.Execute(context.Background(), schema.ToolCall{
		ID:   "call_workdir",
		Name: "bash",
		Arguments: []byte(`{"command":"pwd","workdir":"` +
			subDir +
			`"}`),
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", result.Output)
	}
	if result.Output != subDir {
		t.Fatalf("expected output %q, got %q", subDir, result.Output)
	}
}
