package builtin

import (
	"fmt"
	"os"
	"strings"

	"github.com/houruizhe/go-claude-code/pkg/tools"
)

// FileEditTool edits a file by replacing text
type FileEditTool struct {
	allowedPaths []string
}

// NewFileEditTool creates a new file edit tool
func NewFileEditTool(allowedPaths []string) *FileEditTool {
	return &FileEditTool{allowedPaths: allowedPaths}
}

// Name returns the name of the tool
func (t *FileEditTool) Name() string {
	return "file_edit"
}

// Description returns the description of the tool
func (t *FileEditTool) Description() string {
	return "Edit a file by replacing text. You must use the Read tool at least once before editing. ALWAYS prefer editing existing files in the codebase. NEVER write new files unless they're absolutely necessary."
}

// Parameters returns the JSON schema for the tool's parameters
func (t *FileEditTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"file_path": map[string]interface{}{
			"type":        "string",
			"description": "The absolute path to the file to edit",
		},
		"old_string": map[string]interface{}{
			"type":        "string",
			"description": "The text to replace",
		},
		"new_string": map[string]interface{}{
			"type":        "string",
			"description": "The text to replace with (must be different from old_string)",
		},
		"replace_all": map[string]interface{}{
			"type":        "boolean",
			"description": "Replace all occurrences (default false)",
		},
	}
}

// Execute executes the tool
func (t *FileEditTool) Execute(ctx *tools.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	// Parse file path
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("file_path is required")
	}

	// Parse old string
	oldString, ok := args["old_string"].(string)
	if !ok {
		return nil, fmt.Errorf("old_string is required")
	}

	// Parse new string
	newString, ok := args["new_string"].(string)
	if !ok {
		return nil, fmt.Errorf("new_string is required")
	}

	if oldString == newString {
		return nil, fmt.Errorf("old_string and new_string must be different")
	}

	// Parse replace_all
	replaceAll := false
	if ra, ok := args["replace_all"].(bool); ok {
		replaceAll = ra
	}

	// Check if path is allowed
	if !t.isPathAllowed(filePath) {
		return nil, fmt.Errorf("access to path '%s' is not allowed", filePath)
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Perform replacement
	var newContent string
	if replaceAll {
		newContent = strings.ReplaceAll(string(content), oldString, newString)
	} else {
		// Check if old_string is unique
		count := strings.Count(string(content), oldString)
		if count == 0 {
			return nil, fmt.Errorf("old_string not found in file")
		}
		if count > 1 {
			return nil, fmt.Errorf("old_string appears %d times in file; use replace_all=true or provide more context", count)
		}
		newContent = strings.Replace(string(content), oldString, newString, 1)
	}

	// Write file
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]interface{}{
		"file_path": filePath,
		"size":      len(newContent),
		"status":    "edited",
	}, nil
}

// isPathAllowed checks if the path is in the allowed paths
func (t *FileEditTool) isPathAllowed(path string) bool {
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
