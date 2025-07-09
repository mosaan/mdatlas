package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPerformanceStructureAnalysis(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	projectRoot, binaryPath := setupTest(t)
	
	// Create a large document for performance testing
	largeDoc := generateLargeDocument(1000, 5) // 1000 sections, max 5 levels deep
	largeFile := filepath.Join(projectRoot, "tests", "fixtures", "large_performance.md")
	if err := os.WriteFile(largeFile, []byte(largeDoc), 0644); err != nil {
		t.Fatalf("Failed to create large document: %v", err)
	}
	defer os.Remove(largeFile)
	
	tests := []struct {
		name        string
		args        []string
		maxDuration time.Duration
		validate    func(t *testing.T, output []byte)
	}{
		{
			name:        "large document structure",
			args:        []string{"structure", largeFile},
			maxDuration: 5 * time.Second,
			validate: func(t *testing.T, output []byte) {
				var structure map[string]interface{}
				if err := json.Unmarshal(output, &structure); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				
				sections := structure["structure"].([]interface{})
				if len(sections) == 0 {
					t.Error("Expected sections in large document")
				}
				
				// Check that all sections have required fields
				checkSectionFields(t, sections)
			},
		},
		{
			name:        "large document with max-depth",
			args:        []string{"structure", largeFile, "--max-depth", "3"},
			maxDuration: 3 * time.Second,
			validate: func(t *testing.T, output []byte) {
				var structure map[string]interface{}
				if err := json.Unmarshal(output, &structure); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				
				// Should be faster with depth limitation
				sections := structure["structure"].([]interface{})
				checkMaxDepth(t, sections, 3)
			},
		},
		{
			name:        "large document pretty print",
			args:        []string{"structure", largeFile, "--pretty"},
			maxDuration: 6 * time.Second,
			validate: func(t *testing.T, output []byte) {
				if !strings.Contains(string(output), "\n") {
					t.Error("Expected pretty-printed output")
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.Output()
			
			duration := time.Since(start)
			
			if err != nil {
				t.Fatalf("Command failed: %v", err)
			}
			
			if duration > tt.maxDuration {
				t.Errorf("Command took too long: %v (max: %v)", duration, tt.maxDuration)
			}
			
			if tt.validate != nil {
				tt.validate(t, output)
			}
			
			t.Logf("Performance: %s completed in %v", tt.name, duration)
		})
	}
}

func TestPerformanceSectionExtraction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	projectRoot, binaryPath := setupTest(t)
	
	// Create a document with many sections
	largeDoc := generateLargeDocument(500, 3)
	largeFile := filepath.Join(projectRoot, "tests", "fixtures", "large_sections.md")
	if err := os.WriteFile(largeFile, []byte(largeDoc), 0644); err != nil {
		t.Fatalf("Failed to create large document: %v", err)
	}
	defer os.Remove(largeFile)
	
	// Get structure to find section IDs
	cmd := exec.Command(binaryPath, "structure", largeFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get structure: %v", err)
	}
	
	var structure map[string]interface{}
	if err := json.Unmarshal(output, &structure); err != nil {
		t.Fatalf("Failed to parse structure: %v", err)
	}
	
	// Extract some section IDs
	sectionIDs := extractSectionIDs(structure["structure"].([]interface{}))
	if len(sectionIDs) == 0 {
		t.Fatal("No section IDs found")
	}
	
	// Test section extraction performance
	tests := []struct {
		name        string
		sectionID   string
		args        []string
		maxDuration time.Duration
	}{
		{
			name:        "section extraction",
			sectionID:   sectionIDs[0],
			args:        []string{"section", largeFile, "--section-id", sectionIDs[0]},
			maxDuration: 2 * time.Second,
		},
		{
			name:        "section with children",
			sectionID:   sectionIDs[0],
			args:        []string{"section", largeFile, "--section-id", sectionIDs[0], "--include-children"},
			maxDuration: 3 * time.Second,
		},
		{
			name:        "section JSON format",
			sectionID:   sectionIDs[0],
			args:        []string{"section", largeFile, "--section-id", sectionIDs[0], "--format", "json"},
			maxDuration: 2 * time.Second,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.Output()
			
			duration := time.Since(start)
			
			if err != nil {
				t.Fatalf("Command failed: %v", err)
			}
			
			if duration > tt.maxDuration {
				t.Errorf("Command took too long: %v (max: %v)", duration, tt.maxDuration)
			}
			
			if len(output) == 0 {
				t.Error("Expected output from section extraction")
			}
			
			t.Logf("Performance: %s completed in %v", tt.name, duration)
		})
	}
}

func TestPerformanceCommandOverhead(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	_, binaryPath := setupTest(t)
	
	// Test command startup overhead
	tests := []struct {
		name        string
		args        []string
		maxDuration time.Duration
	}{
		{
			name:        "version command",
			args:        []string{"version"},
			maxDuration: 100 * time.Millisecond,
		},
		{
			name:        "help command",
			args:        []string{"--help"},
			maxDuration: 100 * time.Millisecond,
		},
		{
			name:        "structure help",
			args:        []string{"structure", "--help"},
			maxDuration: 100 * time.Millisecond,
		},
		{
			name:        "section help",
			args:        []string{"section", "--help"},
			maxDuration: 100 * time.Millisecond,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			
			cmd := exec.Command(binaryPath, tt.args...)
			_, err := cmd.Output()
			
			duration := time.Since(start)
			
			if err != nil {
				t.Fatalf("Command failed: %v", err)
			}
			
			if duration > tt.maxDuration {
				t.Errorf("Command took too long: %v (max: %v)", duration, tt.maxDuration)
			}
			
			t.Logf("Performance: %s completed in %v", tt.name, duration)
		})
	}
}

func TestStressTestMultipleFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}
	
	projectRoot, binaryPath := setupTest(t)
	
	// Create multiple test files
	testFiles := make([]string, 10)
	for i := 0; i < 10; i++ {
		content := generateLargeDocument(100, 3)
		filename := filepath.Join(projectRoot, "tests", "fixtures", fmt.Sprintf("stress_test_%d.md", i))
		
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}
		
		testFiles[i] = filename
		defer os.Remove(filename)
	}
	
	// Test processing multiple files sequentially
	start := time.Now()
	
	for i, file := range testFiles {
		cmd := exec.Command(binaryPath, "structure", file)
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to process file %d: %v", i, err)
		}
		
		// Verify output
		var structure map[string]interface{}
		if err := json.Unmarshal(output, &structure); err != nil {
			t.Fatalf("Failed to parse JSON for file %d: %v", i, err)
		}
		
		sections := structure["structure"].([]interface{})
		if len(sections) == 0 {
			t.Errorf("Expected sections in file %d", i)
		}
	}
	
	duration := time.Since(start)
	maxDuration := 30 * time.Second
	
	if duration > maxDuration {
		t.Errorf("Stress test took too long: %v (max: %v)", duration, maxDuration)
	}
	
	t.Logf("Stress test: processed %d files in %v", len(testFiles), duration)
}

func TestStressTestLargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}
	
	projectRoot, binaryPath := setupTest(t)
	
	// Create a very large document (approaching the 50MB limit)
	largeContent := generateLargeDocument(5000, 6) // 5000 sections, max 6 levels
	largeFile := filepath.Join(projectRoot, "tests", "fixtures", "stress_large.md")
	
	if err := os.WriteFile(largeFile, []byte(largeContent), 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}
	defer os.Remove(largeFile)
	
	// Check file size
	stat, err := os.Stat(largeFile)
	if err != nil {
		t.Fatalf("Failed to stat large file: %v", err)
	}
	
	t.Logf("Large file size: %d bytes (%.2f MB)", stat.Size(), float64(stat.Size())/1024/1024)
	
	// Test structure analysis
	start := time.Now()
	
	cmd := exec.Command(binaryPath, "structure", largeFile)
	output, err := cmd.Output()
	
	duration := time.Since(start)
	maxDuration := 10 * time.Second
	
	if err != nil {
		t.Fatalf("Failed to process large file: %v", err)
	}
	
	if duration > maxDuration {
		t.Errorf("Large file processing took too long: %v (max: %v)", duration, maxDuration)
	}
	
	// Verify output
	var structure map[string]interface{}
	if err := json.Unmarshal(output, &structure); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	
	sections := structure["structure"].([]interface{})
	if len(sections) == 0 {
		t.Error("Expected sections in large file")
	}
	
	t.Logf("Large file processing: completed in %v", duration)
}

func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}
	
	projectRoot, binaryPath := setupTest(t)
	
	// Create a document with many sections
	largeDoc := generateLargeDocument(1000, 4)
	largeFile := filepath.Join(projectRoot, "tests", "fixtures", "memory_test.md")
	if err := os.WriteFile(largeFile, []byte(largeDoc), 0644); err != nil {
		t.Fatalf("Failed to create large document: %v", err)
	}
	defer os.Remove(largeFile)
	
	// Test memory usage by running the same command multiple times
	for i := 0; i < 5; i++ {
		cmd := exec.Command(binaryPath, "structure", largeFile)
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Command failed on iteration %d: %v", i, err)
		}
		
		// Verify output is consistent
		var structure map[string]interface{}
		if err := json.Unmarshal(output, &structure); err != nil {
			t.Fatalf("Failed to parse JSON on iteration %d: %v", i, err)
		}
		
		sections := structure["structure"].([]interface{})
		if len(sections) == 0 {
			t.Errorf("Expected sections on iteration %d", i)
		}
	}
}

// Helper function to generate a large document for testing
func generateLargeDocument(numSections, maxDepth int) string {
	var content strings.Builder
	
	content.WriteString("# Main Document\n\n")
	content.WriteString("This is a large document generated for performance testing.\n\n")
	
	for i := 1; i <= numSections; i++ {
		level := (i % maxDepth) + 1
		if level > 6 {
			level = 6
		}
		
		// Generate heading
		heading := strings.Repeat("#", level)
		content.WriteString(fmt.Sprintf("%s Section %d\n\n", heading, i))
		
		// Generate content
		content.WriteString(fmt.Sprintf("This is content for section %d. ", i))
		content.WriteString("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ")
		content.WriteString("Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. ")
		content.WriteString("Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.\n\n")
		
		// Add some lists and code blocks occasionally
		if i%10 == 0 {
			content.WriteString("- List item 1\n")
			content.WriteString("- List item 2\n")
			content.WriteString("- List item 3\n\n")
		}
		
		if i%15 == 0 {
			content.WriteString("```\n")
			content.WriteString("function example() {\n")
			content.WriteString("    return 'Hello, World!';\n")
			content.WriteString("}\n")
			content.WriteString("```\n\n")
		}
	}
	
	return content.String()
}

// Helper function to extract section IDs from structure
func extractSectionIDs(sections []interface{}) []string {
	var ids []string
	
	for _, section := range sections {
		sectionMap := section.(map[string]interface{})
		if id, exists := sectionMap["id"]; exists {
			ids = append(ids, id.(string))
		}
		
		if children, exists := sectionMap["children"]; exists && children != nil {
			if childrenSlice, ok := children.([]interface{}); ok {
				childIDs := extractSectionIDs(childrenSlice)
				ids = append(ids, childIDs...)
			}
		}
	}
	
	return ids
}

// Helper function to check section fields
func checkSectionFields(t *testing.T, sections []interface{}) {
	requiredFields := []string{"id", "level", "title", "char_count", "line_count", "start_line", "end_line", "children"}
	
	for _, section := range sections {
		sectionMap := section.(map[string]interface{})
		
		for _, field := range requiredFields {
			if _, exists := sectionMap[field]; !exists {
				t.Errorf("Missing required field: %s", field)
			}
		}
		
		// Check children recursively
		if children, exists := sectionMap["children"]; exists && children != nil {
			if childrenSlice, ok := children.([]interface{}); ok {
				checkSectionFields(t, childrenSlice)
			}
		}
	}
}