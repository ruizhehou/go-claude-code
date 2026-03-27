package builtin

import (
	"fmt"
	"os"
	"strings"

	"github.com/houruizhe/go-claude-code/pkg/tools"
)

// FileWriteTool writes content to a file
type FileWriteTool struct {
	allowedPaths []string
}

// NewFileWriteTool creates a new file write tool
func NewFileWriteTool(allowedPaths []string) *FileWriteTool {
	return &FileWriteTool{allowedPaths: allowedPaths}
}

// Name returns the name of the tool
func (t *FileWriteTool) Name() string {
	return "file_write"
}

// Description returns the description of the tool
func (t *FileWriteTool) Description() string {
	return "Write content to a file. The file_path parameter must be an absolute path. This will overwrite the file if it already exists. ALWAYS prefer editing existing files in the codebase. NEVER write new files unless they're absolutely necessary for achieving your goal."
}

// Parameters returns the JSON schema for the tool's parameters
func (t *FileWriteTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"file_path": map[string]interface{}{
			"type":        "string",
			"description": "The absolute path to the file to write",
		},
		"content": map[string]interface{}{
			"type":        "string",
			"description": "The content to write to the file",
		},
	}
}

// Execute executes the tool
func (t *FileWriteTool) Execute(ctx *tools.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	// Parse file path
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("file_path is required")
	}

	// Parse content
	content, ok := args["content"].(string)
	if !ok {
		content = ""
	}

	// Check if path is allowed
	if !t.isPathAllowed(filePath) {
		return nil, fmt.Errorf("access to path '%s' is not allowed", filePath)
	}

	// Write file
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]interface{}{
		"file_path": filePath,
		"size":      len(content),
		"status":    "written",
	}, nil
}

// isPathAllowed checks if the path is in the allowed paths
func (t *FileWriteTool) isPathAllowed(path string) bool {
	// If no allowed paths specified, allow all
	if len(t.allowedPaths) == 0 {
		return true
	}

	// Check if path is within any allowed path
	for _, allowed := range t.allowedPaths {
		if strings.HasPrefix(path, allowed) {
			return true
		}
	}

	return false
}
