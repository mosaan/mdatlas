package core

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/mosaan/mdatlas/pkg/types"
)

func TestParseStructure(t *testing.T) {
	parser := NewParser()
	
	// Test with simple content
	content := []byte(`# Title

Some content

## Section 1

Content of section 1

### Subsection 1.1

Content of subsection 1.1

## Section 2

Content of section 2`)
	
	structure, err := parser.ParseStructure(content)
	if err != nil {
		t.Fatalf("ParseStructure failed: %v", err)
	}
	
	if structure.TotalChars != len(content) {
		t.Errorf("Expected total chars %d, got %d", len(content), structure.TotalChars)
	}
	
	if len(structure.Structure) != 1 {
		t.Errorf("Expected 1 top-level section, got %d", len(structure.Structure))
	}
	
	// Check the title section
	title := structure.Structure[0]
	if title.Title != "Title" {
		t.Errorf("Expected title 'Title', got '%s'", title.Title)
	}
	
	if title.Level != 1 {
		t.Errorf("Expected level 1, got %d", title.Level)
	}
	
	if len(title.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(title.Children))
	}
	
	// Check first child
	section1 := title.Children[0]
	if section1.Title != "Section 1" {
		t.Errorf("Expected title 'Section 1', got '%s'", section1.Title)
	}
	
	if section1.Level != 2 {
		t.Errorf("Expected level 2, got %d", section1.Level)
	}
	
	if len(section1.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(section1.Children))
	}
	
	// Check subsection
	subsection := section1.Children[0]
	if subsection.Title != "Subsection 1.1" {
		t.Errorf("Expected title 'Subsection 1.1', got '%s'", subsection.Title)
	}
	
	if subsection.Level != 3 {
		t.Errorf("Expected level 3, got %d", subsection.Level)
	}
}

func TestParseEmptyContent(t *testing.T) {
	parser := NewParser()
	
	structure, err := parser.ParseStructure([]byte(""))
	if err != nil {
		t.Fatalf("ParseStructure failed on empty content: %v", err)
	}
	
	if len(structure.Structure) != 0 {
		t.Errorf("Expected no sections for empty content, got %d", len(structure.Structure))
	}
}

func TestParseNoHeadings(t *testing.T) {
	parser := NewParser()
	
	content := []byte(`This is just regular text.

Some more text without any headings.

Even more text.`)
	
	structure, err := parser.ParseStructure(content)
	if err != nil {
		t.Fatalf("ParseStructure failed: %v", err)
	}
	
	if len(structure.Structure) != 0 {
		t.Errorf("Expected no sections for content without headings, got %d", len(structure.Structure))
	}
}

func TestGetSectionContent(t *testing.T) {
	parser := NewParser()
	
	content := []byte(`# Title

Some content

## Section 1

Content of section 1

### Subsection 1.1

Content of subsection 1.1

## Section 2

Content of section 2`)
	
	structure, err := parser.ParseStructure(content)
	if err != nil {
		t.Fatalf("ParseStructure failed: %v", err)
	}
	
	// Get section content
	if len(structure.Structure) > 0 && len(structure.Structure[0].Children) > 0 {
		sectionID := structure.Structure[0].Children[0].ID
		sectionContent, err := parser.GetSectionContent(content, sectionID, false)
		if err != nil {
			t.Fatalf("GetSectionContent failed: %v", err)
		}
		
		if sectionContent.Title != "Section 1" {
			t.Errorf("Expected title 'Section 1', got '%s'", sectionContent.Title)
		}
		
		if sectionContent.Format != "markdown" {
			t.Errorf("Expected format 'markdown', got '%s'", sectionContent.Format)
		}
	}
}

func TestParseFixtureFiles(t *testing.T) {
	parser := NewParser()
	
	// Get the fixtures directory
	fixturesDir := filepath.Join("..", "..", "tests", "fixtures")
	
	// Test sample.md
	samplePath := filepath.Join(fixturesDir, "sample.md")
	if _, err := os.Stat(samplePath); err == nil {
		content, err := os.ReadFile(samplePath)
		if err != nil {
			t.Fatalf("Failed to read sample.md: %v", err)
		}
		
		structure, err := parser.ParseStructure(content)
		if err != nil {
			t.Fatalf("ParseStructure failed on sample.md: %v", err)
		}
		
		if len(structure.Structure) == 0 {
			t.Error("Expected at least one section in sample.md")
		}
		
		// Verify structure has sections
		if structure.TotalChars == 0 {
			t.Error("Expected non-zero total chars")
		}
		
		if structure.TotalLines == 0 {
			t.Error("Expected non-zero total lines")
		}
	}
	
	// Test complex.md
	complexPath := filepath.Join(fixturesDir, "complex.md")
	if _, err := os.Stat(complexPath); err == nil {
		content, err := os.ReadFile(complexPath)
		if err != nil {
			t.Fatalf("Failed to read complex.md: %v", err)
		}
		
		structure, err := parser.ParseStructure(content)
		if err != nil {
			t.Fatalf("ParseStructure failed on complex.md: %v", err)
		}
		
		if len(structure.Structure) == 0 {
			t.Error("Expected at least one section in complex.md")
		}
		
		// Check for deep nesting
		hasDeepNesting := false
		checkNesting(structure.Structure, &hasDeepNesting, 0)
		
		if !hasDeepNesting {
			t.Error("Expected deep nesting in complex.md")
		}
	}
}

func checkNesting(sections []types.Section, hasDeepNesting *bool, level int) {
	if level > 3 {
		*hasDeepNesting = true
		return
	}
	
	for _, section := range sections {
		checkNesting(section.Children, hasDeepNesting, level+1)
	}
}

func TestGenerateSectionID(t *testing.T) {
	parser := NewParser()
	
	// Test that IDs are generated consistently
	content := []byte(`# Same Title

## Same Title`)
	
	structure, err := parser.ParseStructure(content)
	if err != nil {
		t.Fatalf("ParseStructure failed: %v", err)
	}
	
	if len(structure.Structure) == 0 {
		t.Fatal("Expected at least one section")
	}
	
	// IDs should be unique even for same titles at different levels
	if len(structure.Structure[0].Children) > 0 {
		titleID := structure.Structure[0].ID
		childID := structure.Structure[0].Children[0].ID
		
		if titleID == childID {
			t.Error("Expected different IDs for sections with same title at different levels")
		}
	}
}