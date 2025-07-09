package core

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mosaan/mdatlas/pkg/types"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Parser handles Markdown parsing and structure extraction
type Parser struct {
	md goldmark.Markdown
}

// NewParser creates a new Parser instance
func NewParser() *Parser {
	return &Parser{
		md: goldmark.New(
			goldmark.WithExtensions(
			// Add necessary extensions here
			),
		),
	}
}

// ParseStructure parses the content and extracts document structure
func (p *Parser) ParseStructure(content []byte) (*types.DocumentStructure, error) {
	doc := p.md.Parser().Parse(text.NewReader(content))

	structure := &types.DocumentStructure{
		TotalChars:   len(content),
		TotalLines:   bytes.Count(content, []byte("\n")) + 1,
		Structure:    []types.Section{},
		LastModified: time.Now(),
	}

	// Extract sections from AST
	sections := p.extractSections(doc, content)
	structure.Structure = p.buildHierarchy(sections)

	return structure, nil
}

// extractSections walks through the AST and extracts section information
func (p *Parser) extractSections(doc ast.Node, content []byte) []types.Section {
	var sections []types.Section

	err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && node.Kind() == ast.KindHeading {
			section := p.extractSection(node, content)
			sections = append(sections, section)
		}
		return ast.WalkContinue, nil
	})

	if err != nil {
		// Handle error gracefully
		return sections
	}

	return sections
}

// extractSection extracts section information from a heading node
func (p *Parser) extractSection(node ast.Node, content []byte) types.Section {
	heading := node.(*ast.Heading)

	title := p.extractHeadingText(heading, content)
	startLine := p.getLineNumber(node, content)
	endLine := p.calculateEndLine(node, content)

	return types.Section{
		ID:        p.generateSectionID(heading, title),
		Level:     heading.Level,
		Title:     title,
		StartLine: startLine,
		EndLine:   endLine,
		CharCount: p.calculateCharCount(node, content, startLine, endLine),
		LineCount: endLine - startLine + 1,
		Children:  []types.Section{},
	}
}

// extractHeadingText extracts the text content from a heading node
func (p *Parser) extractHeadingText(heading *ast.Heading, content []byte) string {
	var text strings.Builder

	for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
		if textNode, ok := child.(*ast.Text); ok {
			text.Write(textNode.Value(content))
		}
	}

	return strings.TrimSpace(text.String())
}

// generateSectionID generates a unique ID for a section
func (p *Parser) generateSectionID(heading *ast.Heading, title string) string {
	// Create a hash-based ID for uniqueness
	hash := sha256.Sum256([]byte(title + strconv.Itoa(heading.Level)))
	return fmt.Sprintf("section_%x", hash[:8])
}

// getLineNumber calculates the line number of a node in the content
func (p *Parser) getLineNumber(node ast.Node, content []byte) int {
	segment := node.Lines().At(0)
	if segment.IsEmpty() {
		return 1
	}

	// Count newlines before the segment start
	beforeSegment := content[:segment.Start]
	return bytes.Count(beforeSegment, []byte("\n")) + 1
}

// calculateEndLine calculates the end line of a section
func (p *Parser) calculateEndLine(node ast.Node, content []byte) int {
	// For now, use a simple approach - this can be enhanced
	lines := node.Lines()
	if lines.Len() == 0 {
		return p.getLineNumber(node, content)
	}

	lastSegment := lines.At(lines.Len() - 1)
	beforeEnd := content[:lastSegment.Stop]
	return bytes.Count(beforeEnd, []byte("\n")) + 1
}

// calculateCharCount calculates the character count for a section
func (p *Parser) calculateCharCount(node ast.Node, content []byte, startLine, endLine int) int {
	// Simple implementation - can be enhanced for more accurate counting
	lines := strings.Split(string(content), "\n")
	if startLine > len(lines) || endLine > len(lines) || startLine < 1 {
		return 0
	}

	var charCount int
	for i := startLine - 1; i < endLine && i < len(lines); i++ {
		charCount += len(lines[i]) + 1 // +1 for newline
	}

	return charCount
}

// buildHierarchy builds a hierarchical structure from flat sections
func (p *Parser) buildHierarchy(sections []types.Section) []types.Section {
	if len(sections) == 0 {
		return sections
	}

	var result []types.Section
	stack := make([]*types.Section, 0)

	for _, section := range sections {
		// Find the correct parent level
		for len(stack) > 0 && stack[len(stack)-1].Level >= section.Level {
			stack = stack[:len(stack)-1]
		}

		sectionCopy := section
		if len(stack) == 0 {
			result = append(result, sectionCopy)
		} else {
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, sectionCopy)
		}

		stack = append(stack, &sectionCopy)
	}

	return result
}

// GetSectionContent retrieves the content of a specific section
func (p *Parser) GetSectionContent(content []byte, sectionID string, includeChildren bool) (*types.SectionContent, error) {
	structure, err := p.ParseStructure(content)
	if err != nil {
		return nil, err
	}

	section := p.findSection(structure.Structure, sectionID)
	if section == nil {
		return nil, fmt.Errorf("section not found: %s", sectionID)
	}

	sectionContent := &types.SectionContent{
		ID:              section.ID,
		Title:           section.Title,
		Format:          "markdown",
		IncludeChildren: includeChildren,
	}

	// Extract content based on line numbers
	lines := strings.Split(string(content), "\n")
	if section.StartLine > 0 && section.StartLine <= len(lines) {
		endLine := section.EndLine
		if !includeChildren {
			// Find the next section at the same or higher level
			endLine = p.findSectionEnd(structure.Structure, section)
		}

		if endLine > len(lines) {
			endLine = len(lines)
		}

		contentLines := lines[section.StartLine-1 : endLine]
		sectionContent.Content = strings.Join(contentLines, "\n")
	}

	return sectionContent, nil
}

// findSection recursively finds a section by ID
func (p *Parser) findSection(sections []types.Section, sectionID string) *types.Section {
	for _, section := range sections {
		if section.ID == sectionID {
			return &section
		}
		if found := p.findSection(section.Children, sectionID); found != nil {
			return found
		}
	}
	return nil
}

// findSectionEnd finds the end line of a section (excluding children)
func (p *Parser) findSectionEnd(sections []types.Section, target *types.Section) int {
	// Simple implementation - find the next section at the same or higher level
	found := false
	for _, section := range sections {
		if found && section.Level <= target.Level {
			return section.StartLine - 1
		}
		if section.ID == target.ID {
			found = true
		}
	}
	return target.EndLine
}
