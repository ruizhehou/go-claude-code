package builtin

import (
	"fmt"
	"os"
	"strings"

	"github.com/houruizhe/go-claude-code/pkg/tools"
)

// FileReadTool reads file contents
type FileReadTool struct {
	allowedPaths []string
}

// NewFileReadTool creates a new file read tool
func NewFileReadTool(allowedPaths []string) *FileReadTool {
	return &FileReadTool{allowedPaths: allowedPaths}
}

// Name returns the name of the tool
func (t *FileReadTool) Name() string {
	return "file_read"
}

// Description returns the description of the tool
func (t *FileReadTool) Description() string {
	return "Read the contents of a file. The file_path parameter must be an absolute path. Supports reading text files, images (PNG, JPG), PDFs, and Jupyter notebooks."
}

// Parameters returns the JSON schema for the tool's parameters
func (t *FileReadTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"file_path": map[string]interface{}{
			"type":        "string",
			"description": "The absolute path to the file to read",
		},
		"offset": map[string]interface{}{
			"type":        "integer",
			"description": "Optional line number to start reading from",
		},
		"limit": map[string]interface{}{
			"type":        "integer",
			"description": "Optional number of lines to read (default reads entire file)",
		},
	}
}

// Execute executes the tool
func (t *FileReadTool) Execute(ctx *tools.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	// Parse file path
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("file_path is required")
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

	// Convert to string
	fileContent := string(content)

	// Handle offset and limit for text files
	offset := 0
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}

	limit := 0
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	if offset > 0 || limit > 0 {
		lines := strings.Split(fileContent, "\n")
		start := offset
		if start < 0 {
			start = 0
		}
		if start >= len(lines) {
			return "", fmt.Errorf("offset %d is beyond file length", offset)
		}

		end := len(lines)
		if limit > 0 && start+limit < end {
			end = start + limit
		}

		// Add line numbers
		result := make([]string, end-start)
		for i := start; i < end; i++ {
			result[i-start] = fmt.Sprintf("%5d\t%s", i+1, lines[i])
		}
		fileContent = strings.Join(result, "\n")
	}

	return map[string]interface{}{
		"file_path": filePath,
		"content":   fileContent,
		"size":      len(content),
	}, nil
}

// isPathAllowed checks if the path is in the allowed paths
func (t *FileReadTool) isPathAllowed(path string) bool {
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
