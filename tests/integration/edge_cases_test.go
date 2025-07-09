package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEdgeCasesEmptyFile(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	// Create empty test file
	emptyFile := filepath.Join(projectRoot, "tests", "fixtures", "empty.md")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	defer os.Remove(emptyFile)
	
	// Test structure command with empty file
	cmd := exec.Command(binaryPath, "structure", emptyFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	var structure map[string]interface{}
	if err := json.Unmarshal(output, &structure); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	
	// Should have empty structure
	sections := structure["structure"].([]interface{})
	if len(sections) != 0 {
		t.Errorf("Expected empty structure, got %d sections", len(sections))
	}
	
	// Should have correct metadata
	if structure["total_chars"].(float64) != 0 {
		t.Errorf("Expected 0 total_chars, got %v", structure["total_chars"])
	}
	
	if structure["total_lines"].(float64) != 1 {
		t.Errorf("Expected 1 total_lines, got %v", structure["total_lines"])
	}
}

func TestEdgeCasesNoHeadings(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	// Create file with no headings
	content := `This is just plain text.

Some more text without any headings.

Even more text with **bold** and *italic* formatting.

A list:
- Item 1
- Item 2
- Item 3

And some code:
` + "```" + `
function example() {
    return "hello";
}
` + "```" + `

But no headings at all.`
	
	noHeadingsFile := filepath.Join(projectRoot, "tests", "fixtures", "no_headings.md")
	if err := os.WriteFile(noHeadingsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create no headings file: %v", err)
	}
	defer os.Remove(noHeadingsFile)
	
	// Test structure command
	cmd := exec.Command(binaryPath, "structure", noHeadingsFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	var structure map[string]interface{}
	if err := json.Unmarshal(output, &structure); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	
	// Should have no sections
	sections := structure["structure"].([]interface{})
	if len(sections) != 0 {
		t.Errorf("Expected no sections, got %d sections", len(sections))
	}
	
	// Should have correct character count
	if structure["total_chars"].(float64) != float64(len(content)) {
		t.Errorf("Expected %d total_chars, got %v", len(content), structure["total_chars"])
	}
}

func TestEdgeCasesUnicodeContent(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	// Create file with Unicode content
	unicodeContent := `# 日本語のタイトル

これは日本語のコンテンツです。

## 中文标题

这是中文内容。

### Русский заголовок

Это русский контент.

#### العربية عنوان

هذا محتوى باللغة العربية.

##### 🚀 Emoji Title

Content with emojis: 🎉 🎊 🎈

###### Mixed: 日本語 + English + 中文

Mixed language content.`
	
	unicodeFile := filepath.Join(projectRoot, "tests", "fixtures", "unicode.md")
	if err := os.WriteFile(unicodeFile, []byte(unicodeContent), 0644); err != nil {
		t.Fatalf("Failed to create unicode file: %v", err)
	}
	defer os.Remove(unicodeFile)
	
	// Test structure command
	cmd := exec.Command(binaryPath, "structure", unicodeFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	var structure map[string]interface{}
	if err := json.Unmarshal(output, &structure); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	
	// Should have 6 sections
	sections := structure["structure"].([]interface{})
	if len(sections) != 1 {
		t.Errorf("Expected 1 top-level section, got %d sections", len(sections))
	}
	
	// Check that Unicode titles are properly handled
	topSection := sections[0].(map[string]interface{})
	if topSection["title"].(string) != "日本語のタイトル" {
		t.Errorf("Expected Unicode title, got %s", topSection["title"])
	}
}

func TestEdgeCasesVeryLongLines(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	// Create file with very long lines
	longLine := strings.Repeat("This is a very long line that goes on and on and on. ", 100)
	longContent := `# Title

` + longLine + `

## Another Section

` + longLine + `

### Subsection

` + longLine
	
	longFile := filepath.Join(projectRoot, "tests", "fixtures", "long_lines.md")
	if err := os.WriteFile(longFile, []byte(longContent), 0644); err != nil {
		t.Fatalf("Failed to create long lines file: %v", err)
	}
	defer os.Remove(longFile)
	
	// Test structure command
	cmd := exec.Command(binaryPath, "structure", longFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	var structure map[string]interface{}
	if err := json.Unmarshal(output, &structure); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	
	// Should handle long lines properly
	sections := structure["structure"].([]interface{})
	if len(sections) != 1 {
		t.Errorf("Expected 1 section, got %d sections", len(sections))
	}
	
	// Check total characters
	if structure["total_chars"].(float64) != float64(len(longContent)) {
		t.Errorf("Expected %d total_chars, got %v", len(longContent), structure["total_chars"])
	}
}

func TestEdgeCasesDeepNesting(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	// Create file with maximum nesting (H1 to H6)
	deepContent := `# Level 1
## Level 2
### Level 3
#### Level 4
##### Level 5
###### Level 6

Content at level 6.

##### Back to Level 5
#### Back to Level 4
### Back to Level 3
## Back to Level 2
# Another Level 1`
	
	deepFile := filepath.Join(projectRoot, "tests", "fixtures", "deep_nesting.md")
	if err := os.WriteFile(deepFile, []byte(deepContent), 0644); err != nil {
		t.Fatalf("Failed to create deep nesting file: %v", err)
	}
	defer os.Remove(deepFile)
	
	// Test structure command
	cmd := exec.Command(binaryPath, "structure", deepFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	var structure map[string]interface{}
	if err := json.Unmarshal(output, &structure); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	
	// Should handle deep nesting properly
	sections := structure["structure"].([]interface{})
	if len(sections) != 2 {
		t.Errorf("Expected 2 top-level sections, got %d sections", len(sections))
	}
	
	// Check deep nesting
	firstSection := sections[0].(map[string]interface{})
	if firstSection["level"].(float64) != 1 {
		t.Errorf("Expected level 1, got %v", firstSection["level"])
	}
	
	// Verify the hierarchy goes deep
	hasLevel6 := checkForLevel(firstSection, 6)
	if !hasLevel6 {
		t.Error("Expected to find level 6 section in hierarchy")
	}
}

func TestEdgeCasesSpecialCharacters(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	// Create file with special characters in headings
	specialContent := `# Title with !@#$%^&*()
## Title with "quotes" and 'apostrophes'
### Title with <tags> and &entities;
#### Title with [brackets] and {braces}
##### Title with \backslashes\ and /slashes/
###### Title with |pipes| and ~tildes~

Content with special characters.`
	
	specialFile := filepath.Join(projectRoot, "tests", "fixtures", "special_chars.md")
	if err := os.WriteFile(specialFile, []byte(specialContent), 0644); err != nil {
		t.Fatalf("Failed to create special chars file: %v", err)
	}
	defer os.Remove(specialFile)
	
	// Test structure command
	cmd := exec.Command(binaryPath, "structure", specialFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	var structure map[string]interface{}
	if err := json.Unmarshal(output, &structure); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	
	// Should handle special characters properly
	sections := structure["structure"].([]interface{})
	if len(sections) != 1 {
		t.Errorf("Expected 1 top-level section, got %d sections", len(sections))
	}
	
	// Check that special characters are preserved
	topSection := sections[0].(map[string]interface{})
	if !strings.Contains(topSection["title"].(string), "!@#$%^&*()") {
		t.Error("Expected special characters to be preserved in title")
	}
}

func TestEdgeCasesMarkdownFormatting(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	// Create file with markdown formatting in headings
	formattedContent := `# Title with **bold** and *italic*
## Title with ` + "`code`" + ` and [link](http://example.com)
### Title with ~~strikethrough~~ and ==highlight==
#### Title with ^superscript^ and ~subscript~

Content with various formatting.`
	
	formattedFile := filepath.Join(projectRoot, "tests", "fixtures", "formatted_headings.md")
	if err := os.WriteFile(formattedFile, []byte(formattedContent), 0644); err != nil {
		t.Fatalf("Failed to create formatted headings file: %v", err)
	}
	defer os.Remove(formattedFile)
	
	// Test structure command
	cmd := exec.Command(binaryPath, "structure", formattedFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	var structure map[string]interface{}
	if err := json.Unmarshal(output, &structure); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	
	// Should handle formatted headings properly
	sections := structure["structure"].([]interface{})
	if len(sections) != 1 {
		t.Errorf("Expected 1 top-level section, got %d sections", len(sections))
	}
	
	// Check that markdown formatting is extracted as plain text
	topSection := sections[0].(map[string]interface{})
	title := topSection["title"].(string)
	if strings.Contains(title, "**") || strings.Contains(title, "*") {
		t.Error("Expected markdown formatting to be removed from title")
	}
}

func TestEdgeCasesFilePermissions(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	// Create file with restricted permissions
	restrictedFile := filepath.Join(projectRoot, "tests", "fixtures", "restricted.md")
	if err := os.WriteFile(restrictedFile, []byte("# Restricted File\n\nContent"), 0644); err != nil {
		t.Fatalf("Failed to create restricted file: %v", err)
	}
	defer os.Remove(restrictedFile)
	
	// Make file unreadable
	if err := os.Chmod(restrictedFile, 0000); err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}
	
	// Restore permissions for cleanup
	defer func() {
		os.Chmod(restrictedFile, 0644)
	}()
	
	// Test structure command with unreadable file
	cmd := exec.Command(binaryPath, "structure", restrictedFile)
	_, err := cmd.Output()
	if err == nil {
		t.Error("Expected error for unreadable file")
	}
}

func TestEdgeCasesLargeNumberOfSections(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	// Create file with many sections
	var content strings.Builder
	content.WriteString("# Main Title\n\n")
	
	for i := 1; i <= 50; i++ {
		content.WriteString(fmt.Sprintf("## Section %d\n\n", i))
		content.WriteString("Some content here.\n\n")
		
		for j := 1; j <= 3; j++ {
			content.WriteString(fmt.Sprintf("### Subsection %d.%d\n\n", i, j))
			content.WriteString("More content here.\n\n")
		}
	}
	
	manyFile := filepath.Join(projectRoot, "tests", "fixtures", "many_sections.md")
	if err := os.WriteFile(manyFile, []byte(content.String()), 0644); err != nil {
		t.Fatalf("Failed to create many sections file: %v", err)
	}
	defer os.Remove(manyFile)
	
	// Test structure command
	cmd := exec.Command(binaryPath, "structure", manyFile)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	var structure map[string]interface{}
	if err := json.Unmarshal(output, &structure); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	
	// Should handle many sections properly
	sections := structure["structure"].([]interface{})
	if len(sections) != 1 {
		t.Errorf("Expected 1 top-level section, got %d sections", len(sections))
	}
	
	// Check total section count
	topSection := sections[0].(map[string]interface{})
	children := topSection["children"].([]interface{})
	if len(children) != 50 {
		t.Errorf("Expected 50 child sections, got %d", len(children))
	}
}

func TestEdgeCasesInvalidSectionID(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	testFile := filepath.Join(projectRoot, "tests", "fixtures", "sample.md")
	
	invalidIDs := []string{
		"",
		"invalid",
		"section_invalid",
		"nonexistent_id",
		"section_123456789",
		"null",
		"undefined",
	}
	
	for _, invalidID := range invalidIDs {
		t.Run(fmt.Sprintf("invalid_id_%s", invalidID), func(t *testing.T) {
			cmd := exec.Command(binaryPath, "section", testFile, "--section-id", invalidID)
			_, err := cmd.Output()
			if err == nil {
				t.Errorf("Expected error for invalid section ID: %s", invalidID)
			}
		})
	}
}

// Helper function to check if a section hierarchy contains a specific level
func checkForLevel(section map[string]interface{}, targetLevel int) bool {
	level := int(section["level"].(float64))
	if level == targetLevel {
		return true
	}
	
	if children, exists := section["children"]; exists {
		for _, child := range children.([]interface{}) {
			if checkForLevel(child.(map[string]interface{}), targetLevel) {
				return true
			}
		}
	}
	
	return false
}

