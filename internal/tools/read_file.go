package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"mhkyle/my-harness/internal/schema"
)

const (
	maxLengthBytes = 8000
)

type ReadFileTool struct {
	workDir string
}

type readFileArgs struct {
	Path string `json:"path"`
}

func NewReadFileTool(workDir string) *ReadFileTool {
	return &ReadFileTool{
		workDir: workDir,
	}
}

func (m *ReadFileTool) Name() string {
	return "read_file"
}

func (m *ReadFileTool) Definition() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        m.Name(),
		Description: "Read the content of a file in the current directory.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to read.",
				},
			},
			"required": []string{"path"},
		},
	}
}

func (m *ReadFileTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var input readFileArgs
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %v", err)
	}

	fullPath := filepath.Join(m.workDir, input.Path)

	file, err := os.Open(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	if len(content) > maxLengthBytes {
		truncatedContent := content[:maxLengthBytes]
		fmt.Printf("content truncated to %d bytes\n", maxLengthBytes)
		return string(truncatedContent), nil
	}

	return string(content), nil
}
