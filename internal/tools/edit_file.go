package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mhkyle/my-harness/internal/schema"
)

type EditFileTool struct {
	workDir string
}

func NewEditFileTool(workDir string) *EditFileTool {
	return &EditFileTool{
		workDir: workDir,
	}
}

type editFileArgs struct {
	Path    string `json:"path"`
	OldText string `json:"old_text"`
	NewText string `json:"new_text"`
}

func (t *EditFileTool) Name() string {
	return "edit_file"
}

func (t *EditFileTool) Definition() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        t.Name(),
		Description: "Edit a file in the workspace. You can use this tool to modify the content of an existing file.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The relative path of the file to edit, relative to the workspace root.",
				},
				"old_text": map[string]interface{}{
					"type":        "string",
					"description": "The text to be replaced in the file. It should contain enough context (suggested to contain more than 1 lines)",
				},
				"new_text": map[string]interface{}{
					"type":        "string",
					"description": "The new text to replace the old text with.",
				},
			},
			"required": []string{"path", "old_text", "new_text"},
		},
	}
}

func (t *EditFileTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var input editFileArgs
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	fullPath := filepath.Join(t.workDir, input.Path)

	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file, please check if the path is correct: %w", err)
	}
	originalContent := string(contentBytes)

	newContent, err := fuzzyReplace(originalContent, input.OldText, input.NewText)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(fullPath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write back to file: %w", err)
	}

	return fmt.Sprintf("✅ Successfully edited file: %s", input.Path), nil
}

// fuzzyReplace performs a fuzzy replacement of oldText with newText in originalContent. It tries multiple levels of matching to find the best match for oldText in originalContent.
func fuzzyReplace(originalContent, oldText, newText string) (string, error) {
	// L1: Exact Match
	count := strings.Count(originalContent, oldText)
	if count == 1 {
		return strings.Replace(originalContent, oldText, newText, 1), nil
	}
	if count > 1 {
		return "", fmt.Errorf("old_text matched %d times, please provide more context to ensure uniqueness", count)
	}

	// L2: Normalize Line Endings Match
	normalizedContent := strings.ReplaceAll(originalContent, "\r\n", "\n")
	normalizedOld := strings.ReplaceAll(oldText, "\r\n", "\n")

	count = strings.Count(normalizedContent, normalizedOld)
	if count == 1 {
		return strings.Replace(normalizedContent, normalizedOld, newText, 1), nil
	}

	// L3: Trim Space Match
	trimmedOld := strings.TrimSpace(normalizedOld)
	if trimmedOld != "" {
		count = strings.Count(normalizedContent, trimmedOld)
		if count == 1 {
			return strings.Replace(normalizedContent, trimmedOld, newText, 1), nil
		}
	}

	// L4: Line-by-Line Match
	return lineByLineReplace(normalizedContent, normalizedOld, newText)
}

// lineByLineReplace performs a line-by-line replacement of oldText with newText in originalContent. It looks for a block of lines that matches oldText and replaces it with newText.
func lineByLineReplace(content, oldText, newText string) (string, error) {
	contentLines := strings.Split(content, "\n")
	oldLines := strings.Split(strings.TrimSpace(oldText), "\n")

	if len(oldLines) == 0 || len(contentLines) < len(oldLines) {
		return "", fmt.Errorf("old_text is empty or longer than the content, cannot perform line-by-line replacement")
	}

	for i := range oldLines {
		oldLines[i] = strings.TrimSpace(oldLines[i])
	}

	matchCount := 0
	matchStartIndex := -1
	matchEndIndex := -1

	for i := 0; i <= len(contentLines)-len(oldLines); i++ {
		isMatch := true
		for j := 0; j < len(oldLines); j++ {
			if strings.TrimSpace(contentLines[i+j]) != oldLines[j] {
				isMatch = false
				break
			}
		}

		if isMatch {
			matchCount++
			matchStartIndex = i
			matchEndIndex = i + len(oldLines)
		}
	}

	if matchCount == 0 {
		return "", fmt.Errorf("old_text not found in the file, please use read_file to carefully check the file content and indentation")
	}
	if matchCount > 1 {
		return "", fmt.Errorf("fuzzy matched %d similar code blocks, please provide more surrounding lines for precise location", matchCount)
	}

	var newContentLines []string
	newContentLines = append(newContentLines, contentLines[:matchStartIndex]...)
	newContentLines = append(newContentLines, newText)
	newContentLines = append(newContentLines, contentLines[matchEndIndex:]...)

	return strings.Join(newContentLines, "\n"), nil
}
