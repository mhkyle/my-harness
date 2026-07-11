package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"mhkyle/my-harness/internal/schema"
)

const (
	defaultBashTimeout   = 30 * time.Second
	defaultMaxOutputSize = 8000
)

type BashTool struct {
}

type bashInput struct {
	Command string `json:"command"`
	WorkDir string `json:"workdir,omitempty"`
}

func NewBashTool() BaseTool {
	return &BashTool{}
}

func (m *BashTool) Name() string {
	return "bash"
}

func (m *BashTool) Definition() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        "bash",
		Description: "Execute bash commands in the current directory.",
		InputSchema: string(`{"type":"object","properties":{"command":{"type":"string","description":"Bash command to execute."},"workdir":{"type":"string","description":"Optional working directory for command execution."}},"required":["command"]}`),
	}
}

func (m *BashTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var input bashInput
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %v", err)
	}

	if strings.TrimSpace(input.Command) == "" {
		return "", fmt.Errorf("command is required")
	}

	workDir := input.WorkDir
	if workDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %v", err)
		}
		workDir = cwd
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultBashTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "bash", "-lc", input.Command)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()

	if timeoutCtx.Err() == context.DeadlineExceeded {
		return fmt.Sprintf("timeout execute the command after %s, current output is %s", defaultBashTimeout, strings.TrimRight(string(output), "\n")), nil
	}

	if err != nil {
		return fmt.Sprintf("command failed: %v, output: %s", err, strings.TrimRight(string(output), "\n")), nil
	}

	if len(output) == 0 {
		return "command executed successfully without any output", nil
	}

	if len(output) > defaultMaxOutputSize {
		return fmt.Sprintf("%s\n\nCommand executed successfully, but output is too long (%d bytes). First %d bytes: %s", strings.TrimRight(string(output[:defaultMaxOutputSize]), "\n"), len(output), defaultMaxOutputSize, strings.TrimRight(string(output[:defaultMaxOutputSize]), "\n")), nil
	}

	return strings.TrimRight(string(output), "\n"), nil
}
