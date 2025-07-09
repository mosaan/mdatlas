package integration

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCLIStructureCommandComprehensive(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	testFile := filepath.Join(projectRoot, "tests", "fixtures", "sample.md")

	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T, output []byte, err error)
	}{
		{
			name: "basic structure extraction",
			args: []string{"structure", testFile},
			validate: func(t *testing.T, output []byte, err error) {
				if err != nil {
					t.Fatalf("Command failed: %v", err)
				}
				var structure map[string]interface{}
				if err := json.Unmarshal(output, &structure); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				// Verify required fields
				requiredFields := []string{"file_path", "total_chars", "total_lines", "structure", "last_modified"}
				for _, field := range requiredFields {
					if _, exists := structure[field]; !exists {
						t.Errorf("Missing required field: %s", field)
					}
				}
			},
		},
		{
			name: "pretty-printed JSON",
			args: []string{"structure", testFile, "--pretty"},
			validate: func(t *testing.T, output []byte, err error) {
				if err != nil {
					t.Fatalf("Command failed: %v", err)
				}
				// Check if output is formatted (contains newlines and indentation)
				if !strings.Contains(string(output), "\n") {
					t.Error("Expected pretty-printed JSON to contain newlines")
				}
				if !strings.Contains(string(output), "  ") {
					t.Error("Expected pretty-printed JSON to contain indentation")
				}
			},
		},
		{
			name: "max-depth limitation",
			args: []string{"structure", testFile, "--max-depth", "2"},
			validate: func(t *testing.T, output []byte, err error) {
				if err != nil {
					t.Fatalf("Command failed: %v", err)
				}
				var structure map[string]interface{}
				if err := json.Unmarshal(output, &structure); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				// Check that no sections exceed level 2
				sections := structure["structure"].([]interface{})
				checkMaxDepth(t, sections, 2)
			},
		},
		{
			name: "base-dir flag",
			args: []string{"--base-dir", filepath.Join(projectRoot, "tests", "fixtures"), "structure", "sample.md"},
			validate: func(t *testing.T, output []byte, err error) {
				if err != nil {
					t.Fatalf("Command failed: %v", err)
				}
				var structure map[string]interface{}
				if err := json.Unmarshal(output, &structure); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				// Should still work with base-dir
				if structure["total_chars"].(float64) == 0 {
					t.Error("Expected non-zero total_chars")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.Output()
			tt.validate(t, output, err)
		})
	}
}

func TestCLISectionCommandComprehensive(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	testFile := filepath.Join(projectRoot, "tests", "fixtures", "sample.md")

	// First get structure to obtain section IDs
	cmd := exec.Command(binaryPath, "structure", testFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get structure: %v", err)
	}

	var structure map[string]interface{}
	if err := json.Unmarshal(output, &structure); err != nil {
		t.Fatalf("Failed to parse structure: %v", err)
	}

	sections := structure["structure"].([]interface{})
	if len(sections) == 0 {
		t.Fatal("No sections found")
	}

	firstSection := sections[0].(map[string]interface{})
	sectionID := firstSection["id"].(string)

	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T, output []byte, err error)
	}{
		{
			name: "basic section extraction",
			args: []string{"section", testFile, "--section-id", sectionID},
			validate: func(t *testing.T, output []byte, err error) {
				if err != nil {
					t.Fatalf("Command failed: %v", err)
				}
				content := string(output)
				if !strings.HasPrefix(content, "#") {
					t.Error("Expected markdown content to start with #")
				}
			},
		},
		{
			name: "section with children",
			args: []string{"section", testFile, "--section-id", sectionID, "--include-children"},
			validate: func(t *testing.T, output []byte, err error) {
				if err != nil {
					t.Fatalf("Command failed: %v", err)
				}
				// Should have more content when including children
				if len(output) < 10 {
					t.Error("Expected more content when including children")
				}
			},
		},
		{
			name: "section in JSON format",
			args: []string{"section", testFile, "--section-id", sectionID, "--format", "json"},
			validate: func(t *testing.T, output []byte, err error) {
				if err != nil {
					t.Fatalf("Command failed: %v", err)
				}
				var sectionContent map[string]interface{}
				if err := json.Unmarshal(output, &sectionContent); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				// Check required fields
				requiredFields := []string{"id", "title", "content", "format"}
				for _, field := range requiredFields {
					if _, exists := sectionContent[field]; !exists {
						t.Errorf("Missing required field: %s", field)
					}
				}
			},
		},
		{
			name: "section in plain format",
			args: []string{"section", testFile, "--section-id", sectionID, "--format", "plain"},
			validate: func(t *testing.T, output []byte, err error) {
				if err != nil {
					t.Fatalf("Command failed: %v", err)
				}
				// Plain format should not contain markdown formatting
				content := string(output)
				if strings.Contains(content, "```") || strings.Contains(content, "**") {
					t.Error("Plain format should not contain markdown formatting")
				}
			},
		},
		{
			name: "JSON format with pretty printing",
			args: []string{"section", testFile, "--section-id", sectionID, "--format", "json", "--pretty"},
			validate: func(t *testing.T, output []byte, err error) {
				if err != nil {
					t.Fatalf("Command failed: %v", err)
				}
				// Should be pretty-printed JSON
				if !strings.Contains(string(output), "\n") {
					t.Error("Expected pretty-printed JSON to contain newlines")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.Output()
			tt.validate(t, output, err)
		})
	}
}

func TestCLIVersionCommandComprehensive(t *testing.T) {
	_, binaryPath := setupTest(t)

	cmd := exec.Command(binaryPath, "version")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Version command failed: %v", err)
	}

	content := string(output)
	if !strings.Contains(content, "mdatlas version") {
		t.Error("Expected version information in output")
	}
	if !strings.Contains(content, "Build date") {
		t.Error("Expected build date in output")
	}
}

func TestCLICompletionCommand(t *testing.T) {
	_, binaryPath := setupTest(t)

	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			cmd := exec.Command(binaryPath, "completion", shell)
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("Completion command failed for %s: %v", shell, err)
			}
			if len(output) == 0 {
				t.Errorf("Expected completion script for %s", shell)
			}
		})
	}
}

func TestCLIHelpCommands(t *testing.T) {
	_, binaryPath := setupTest(t)

	tests := []struct {
		name string
		args []string
		expectedContent []string
	}{
		{
			name: "general help",
			args: []string{"--help"},
			expectedContent: []string{"Usage:", "Available Commands:", "structure", "section", "version"},
		},
		{
			name: "structure help",
			args: []string{"structure", "--help"},
			expectedContent: []string{"Usage:", "mdatlas structure", "--max-depth", "--pretty"},
		},
		{
			name: "section help",
			args: []string{"section", "--help"},
			expectedContent: []string{"Usage:", "mdatlas section", "--section-id", "--format", "--include-children"},
		},
		{
			name: "version help",
			args: []string{"version", "--help"},
			expectedContent: []string{"Usage:", "Print the version information"},
		},
		{
			name: "help command",
			args: []string{"help"},
			expectedContent: []string{"Usage:", "Available Commands:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("Help command failed: %v", err)
			}
			
			content := string(output)
			for _, expected := range tt.expectedContent {
				if !strings.Contains(content, expected) {
					t.Errorf("Expected help content to contain: %s", expected)
				}
			}
		})
	}
}

func TestCLIGlobalFlags(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	_ = filepath.Join(projectRoot, "tests", "fixtures", "sample.md")

	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T, output []byte, err error)
	}{
		{
			name: "base-dir with absolute path",
			args: []string{"--base-dir", filepath.Join(projectRoot, "tests", "fixtures"), "structure", "sample.md"},
			validate: func(t *testing.T, output []byte, err error) {
				if err != nil {
					t.Fatalf("Command failed: %v", err)
				}
				// Should succeed with absolute base-dir
				var structure map[string]interface{}
				if err := json.Unmarshal(output, &structure); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
			},
		},
		{
			name: "base-dir with relative path",
			args: []string{"--base-dir", "tests/fixtures", "structure", "sample.md"},
			validate: func(t *testing.T, output []byte, err error) {
				if err != nil {
					t.Fatalf("Command failed: %v", err)
				}
				// Should succeed with relative base-dir
				var structure map[string]interface{}
				if err := json.Unmarshal(output, &structure); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Dir = projectRoot
			output, err := cmd.Output()
			tt.validate(t, output, err)
		})
	}
}

func TestCLIErrorHandling(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedInError string
	}{
		{
			name:           "nonexistent file",
			args:           []string{"structure", "nonexistent.md"},
			expectError:    true,
			expectedInError: "does not exist",
		},
		{
			name:           "missing section ID",
			args:           []string{"section", filepath.Join(projectRoot, "tests", "fixtures", "sample.md")},
			expectError:    true,
			expectedInError: "required",
		},
		{
			name:           "invalid section ID",
			args:           []string{"section", filepath.Join(projectRoot, "tests", "fixtures", "sample.md"), "--section-id", "invalid"},
			expectError:    true,
			expectedInError: "section not found",
		},
		{
			name:           "invalid format",
			args:           []string{"section", filepath.Join(projectRoot, "tests", "fixtures", "sample.md"), "--section-id", "test", "--format", "invalid"},
			expectError:    true,
			expectedInError: "unsupported format",
		},
		{
			name:           "invalid max-depth",
			args:           []string{"structure", filepath.Join(projectRoot, "tests", "fixtures", "sample.md"), "--max-depth", "-1"},
			expectError:    false, // Should handle gracefully
			expectedInError: "",
		},
		{
			name:           "missing file argument",
			args:           []string{"structure"},
			expectError:    true,
			expectedInError: "accepts 1 arg",
		},
		{
			name:           "too many arguments",
			args:           []string{"structure", "file1.md", "file2.md"},
			expectError:    true,
			expectedInError: "accepts 1 arg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but command succeeded. Output: %s", string(output))
				}
				if tt.expectedInError != "" && !strings.Contains(string(output), tt.expectedInError) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.expectedInError, string(output))
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %v. Output: %s", err, string(output))
				}
			}
		})
	}
}

func TestCLIOutputFormats(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	testFile := filepath.Join(projectRoot, "tests", "fixtures", "sample.md")

	// Get a valid section ID first
	cmd := exec.Command(binaryPath, "structure", testFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get structure: %v", err)
	}

	var structure map[string]interface{}
	if err := json.Unmarshal(output, &structure); err != nil {
		t.Fatalf("Failed to parse structure: %v", err)
	}

	sections := structure["structure"].([]interface{})
	if len(sections) == 0 {
		t.Fatal("No sections found")
	}

	firstSection := sections[0].(map[string]interface{})
	sectionID := firstSection["id"].(string)

	tests := []struct {
		name     string
		format   string
		validate func(t *testing.T, output []byte)
	}{
		{
			name:   "markdown format",
			format: "markdown",
			validate: func(t *testing.T, output []byte) {
				content := string(output)
				if !strings.HasPrefix(content, "#") {
					t.Error("Markdown format should start with #")
				}
			},
		},
		{
			name:   "plain format",
			format: "plain",
			validate: func(t *testing.T, output []byte) {
				content := string(output)
				// Plain format should not have markdown headers
				if strings.Contains(content, "##") {
					t.Error("Plain format should not contain markdown headers")
				}
			},
		},
		{
			name:   "json format",
			format: "json",
			validate: func(t *testing.T, output []byte) {
				var sectionContent map[string]interface{}
				if err := json.Unmarshal(output, &sectionContent); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				if sectionContent["format"] != "json" {
					t.Error("JSON format should have format field set to 'json'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, "section", testFile, "--section-id", sectionID, "--format", tt.format)
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("Command failed: %v", err)
			}
			tt.validate(t, output)
		})
	}
}

func TestCLIPerformance(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	testFile := filepath.Join(projectRoot, "tests", "fixtures", "complex.md")

	tests := []struct {
		name      string
		args      []string
		maxTime   time.Duration
	}{
		{
			name:    "structure analysis performance",
			args:    []string{"structure", testFile},
			maxTime: 2 * time.Second,
		},
		{
			name:    "structure with pretty printing",
			args:    []string{"structure", testFile, "--pretty"},
			maxTime: 2 * time.Second,
		},
		{
			name:    "version command performance",
			args:    []string{"version"},
			maxTime: 100 * time.Millisecond,
		},
		{
			name:    "help command performance",
			args:    []string{"--help"},
			maxTime: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			cmd := exec.Command(binaryPath, tt.args...)
			_, err := cmd.Output()
			elapsed := time.Since(start)
			
			if err != nil {
				t.Fatalf("Command failed: %v", err)
			}
			
			if elapsed > tt.maxTime {
				t.Errorf("Command took too long: %v (max: %v)", elapsed, tt.maxTime)
			}
		})
	}
}

// Helper function to check max depth recursively
func checkMaxDepth(t *testing.T, sections []interface{}, maxDepth int) {
	for _, section := range sections {
		sectionMap := section.(map[string]interface{})
		level := int(sectionMap["level"].(float64))
		if level > maxDepth {
			t.Errorf("Section level %d exceeds max depth %d", level, maxDepth)
		}
		if children, exists := sectionMap["children"]; exists {
			checkMaxDepth(t, children.([]interface{}), maxDepth)
		}
	}
}