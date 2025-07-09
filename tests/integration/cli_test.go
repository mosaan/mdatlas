package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"strings"
)

// Helper function to get project root and build binary
func setupTest(t *testing.T) (string, string) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	
	projectRoot := filepath.Join(wd, "..", "..")
	binaryPath := filepath.Join(projectRoot, "bin", "mdatlas")
	
	// Check if binary exists, build if not
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		buildCmd := exec.Command("make", "build")
		buildCmd.Dir = projectRoot
		if err := buildCmd.Run(); err != nil {
			t.Fatalf("Failed to build binary: %v", err)
		}
	}
	
	return projectRoot, binaryPath
}

func TestCLIStructureCommand(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	// Create test fixture
	testFile := filepath.Join(projectRoot, "tests", "fixtures", "sample.md")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", testFile)
	}
	
	// Run structure command
	cmd := exec.Command(binaryPath, "structure", testFile, "--pretty")
	
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("CLI command failed: %v", err)
	}
	
	// Parse JSON output
	var structure map[string]interface{}
	if err := json.Unmarshal(output, &structure); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}
	
	// Verify structure
	if _, ok := structure["file_path"]; !ok {
		t.Error("Expected file_path in output")
	}
	
	if _, ok := structure["structure"]; !ok {
		t.Error("Expected structure in output")
	}
	
	if _, ok := structure["total_chars"]; !ok {
		t.Error("Expected total_chars in output")
	}
	
	if _, ok := structure["total_lines"]; !ok {
		t.Error("Expected total_lines in output")
	}
}

func TestCLISectionCommand(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	testFile := filepath.Join(projectRoot, "tests", "fixtures", "sample.md")
	
	// First get the structure to find a section ID
	structureOutput, err := exec.Command(binaryPath, "structure", testFile).Output()
	if err != nil {
		t.Fatalf("Failed to get structure: %v", err)
	}
	
	// Parse structure to get section ID
	var structure map[string]interface{}
	if err := json.Unmarshal(structureOutput, &structure); err != nil {
		t.Fatalf("Failed to parse structure: %v", err)
	}
	
	sections, ok := structure["structure"].([]interface{})
	if !ok || len(sections) == 0 {
		t.Skip("No sections found in test file")
	}
	
	firstSection, ok := sections[0].(map[string]interface{})
	if !ok {
		t.Skip("Invalid section structure")
	}
	
	sectionID, ok := firstSection["id"].(string)
	if !ok {
		t.Skip("No section ID found")
	}
	
	// Run section command
	sectionOutput, err := exec.Command(binaryPath, "section", testFile, "--section-id", sectionID).Output()
	if err != nil {
		t.Fatalf("Section command failed: %v", err)
	}
	
	// Verify we got content
	if len(sectionOutput) == 0 {
		t.Error("Expected section content in output")
	}
	
	// Content should be markdown by default
	content := string(sectionOutput)
	if !strings.HasPrefix(content, "#") {
		t.Error("Expected markdown content to start with #")
	}
}

func TestCLIVersionCommand(t *testing.T) {
	_, binaryPath := setupTest(t)
	
	// Run version command
	output, err := exec.Command(binaryPath, "version").Output()
	if err != nil {
		t.Fatalf("Version command failed: %v", err)
	}
	
	// Verify version output
	content := string(output)
	if !strings.Contains(content, "mdatlas version") {
		t.Error("Expected version information in output")
	}
	
	if !strings.Contains(content, "Build date") {
		t.Error("Expected build date in output")
	}
}

func TestCLIHelpCommand(t *testing.T) {
	_, binaryPath := setupTest(t)
	
	// Run help command
	output, err := exec.Command(binaryPath, "--help").Output()
	if err != nil {
		t.Fatalf("Help command failed: %v", err)
	}
	
	// Verify help output
	content := string(output)
	expectedCommands := []string{"structure", "section", "version"}
	
	for _, command := range expectedCommands {
		if !strings.Contains(content, command) {
			t.Errorf("Expected command %s in help output", command)
		}
	}
}

func TestCLIInvalidFile(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("make", "build")
	buildCmd.Dir = filepath.Join("..", "..")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	
	// Run structure command with invalid file
	binaryPath := filepath.Join("..", "..", "bin", "mdatlas")
	cmd := exec.Command(binaryPath, "structure", "nonexistent.md")
	cmd.Dir = filepath.Join("..", "..")
	
	_, err := cmd.Output()
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
	
	// Check that it's the right kind of error
	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() == 0 {
			t.Error("Expected non-zero exit code for invalid file")
		}
	}
}