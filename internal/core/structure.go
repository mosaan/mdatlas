package core

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mosaan/mdatlas/pkg/types"
)

// StructureManager manages document structure information and provides
// higher-level operations for document analysis
type StructureManager struct {
	parser *Parser
	cache  *Cache
}

// NewStructureManager creates a new StructureManager instance
func NewStructureManager(cache *Cache) *StructureManager {
	return &StructureManager{
		parser: NewParser(),
		cache:  cache,
	}
}

// GetDocumentStructure retrieves the structure of a document with caching
func (sm *StructureManager) GetDocumentStructure(filePath string) (*types.DocumentStructure, error) {
	// Check cache first
	if sm.cache != nil {
		if structure, exists := sm.cache.GetStructure(filePath); exists {
			return structure, nil
		}
	}

	// Read file and parse structure
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	structure, err := sm.parser.ParseStructure(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse structure for %s: %w", filePath, err)
	}

	// Set file path and get file modification time
	structure.FilePath = filePath
	if stat, err := os.Stat(filePath); err == nil {
		structure.LastModified = stat.ModTime()
	}

	// Cache the result
	if sm.cache != nil {
		sm.cache.SetStructure(filePath, structure)
	}

	return structure, nil
}

// GetSectionContent retrieves content for a specific section
func (sm *StructureManager) GetSectionContent(filePath, sectionID string, includeChildren bool) (*types.SectionContent, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return sm.parser.GetSectionContent(content, sectionID, includeChildren)
}

// SearchSections searches for sections matching a query
func (sm *StructureManager) SearchSections(filePath, query string, caseSensitive bool) ([]types.Section, error) {
	structure, err := sm.GetDocumentStructure(filePath)
	if err != nil {
		return nil, err
	}

	var results []types.Section
	sm.searchSectionsRecursive(structure.Structure, query, caseSensitive, &results)
	return results, nil
}

// searchSectionsRecursive recursively searches through sections
func (sm *StructureManager) searchSectionsRecursive(sections []types.Section, query string, caseSensitive bool, results *[]types.Section) {
	for _, section := range sections {
		title := section.Title
		searchQuery := query

		if !caseSensitive {
			title = strings.ToLower(title)
			searchQuery = strings.ToLower(searchQuery)
		}

		if strings.Contains(title, searchQuery) {
			*results = append(*results, section)
		}

		// Search in children
		sm.searchSectionsRecursive(section.Children, query, caseSensitive, results)
	}
}

// GetSectionsByLevel returns all sections at a specific level
func (sm *StructureManager) GetSectionsByLevel(filePath string, level int) ([]types.Section, error) {
	structure, err := sm.GetDocumentStructure(filePath)
	if err != nil {
		return nil, err
	}

	var results []types.Section
	sm.collectSectionsByLevel(structure.Structure, level, &results)
	return results, nil
}

// collectSectionsByLevel recursively collects sections at a specific level
func (sm *StructureManager) collectSectionsByLevel(sections []types.Section, targetLevel int, results *[]types.Section) {
	for _, section := range sections {
		if section.Level == targetLevel {
			*results = append(*results, section)
		}

		// Continue searching in children
		sm.collectSectionsByLevel(section.Children, targetLevel, results)
	}
}

// GetDocumentStats returns statistics about the document
func (sm *StructureManager) GetDocumentStats(filePath string) (*DocumentStats, error) {
	structure, err := sm.GetDocumentStructure(filePath)
	if err != nil {
		return nil, err
	}

	stats := &DocumentStats{
		FilePath:     filePath,
		TotalChars:   structure.TotalChars,
		TotalLines:   structure.TotalLines,
		SectionCount: sm.countSections(structure.Structure),
		LevelCounts:  make(map[int]int),
	}

	// Count sections by level
	sm.countSectionsByLevel(structure.Structure, stats.LevelCounts)

	return stats, nil
}

// countSections recursively counts all sections
func (sm *StructureManager) countSections(sections []types.Section) int {
	count := len(sections)
	for _, section := range sections {
		count += sm.countSections(section.Children)
	}
	return count
}

// countSectionsByLevel counts sections by their heading level
func (sm *StructureManager) countSectionsByLevel(sections []types.Section, counts map[int]int) {
	for _, section := range sections {
		counts[section.Level]++
		sm.countSectionsByLevel(section.Children, counts)
	}
}

// ValidateStructure validates the integrity of a document structure
func (sm *StructureManager) ValidateStructure(filePath string) error {
	structure, err := sm.GetDocumentStructure(filePath)
	if err != nil {
		return err
	}

	// Check file exists and is readable
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("file validation failed: %w", err)
	}

	// Validate structure hierarchy
	return sm.validateHierarchy(structure.Structure, 0)
}

// validateHierarchy validates the section hierarchy
func (sm *StructureManager) validateHierarchy(sections []types.Section, parentLevel int) error {
	for _, section := range sections {
		// Check section has valid ID and title
		if section.ID == "" {
			return fmt.Errorf("section missing ID: %s", section.Title)
		}

		if section.Title == "" {
			return fmt.Errorf("section missing title: %s", section.ID)
		}

		// Check level is valid
		if section.Level < 1 || section.Level > 6 {
			return fmt.Errorf("invalid section level %d for section %s", section.Level, section.ID)
		}

		// Check line numbers are valid
		if section.StartLine < 1 || section.EndLine < section.StartLine {
			return fmt.Errorf("invalid line numbers for section %s: start=%d, end=%d",
				section.ID, section.StartLine, section.EndLine)
		}

		// Validate children
		if err := sm.validateHierarchy(section.Children, section.Level); err != nil {
			return err
		}
	}

	return nil
}

// DocumentStats represents statistics about a document
type DocumentStats struct {
	FilePath     string      `json:"file_path"`
	TotalChars   int         `json:"total_chars"`
	TotalLines   int         `json:"total_lines"`
	SectionCount int         `json:"section_count"`
	LevelCounts  map[int]int `json:"level_counts"`
	LastModified time.Time   `json:"last_modified"`
}

// GetTableOfContents generates a table of contents for the document
func (sm *StructureManager) GetTableOfContents(filePath string, maxDepth int) ([]TocEntry, error) {
	structure, err := sm.GetDocumentStructure(filePath)
	if err != nil {
		return nil, err
	}

	var toc []TocEntry
	sm.buildTocRecursive(structure.Structure, maxDepth, &toc)
	return toc, nil
}

// buildTocRecursive recursively builds table of contents
func (sm *StructureManager) buildTocRecursive(sections []types.Section, maxDepth int, toc *[]TocEntry) {
	for _, section := range sections {
		if maxDepth > 0 && section.Level > maxDepth {
			continue
		}

		entry := TocEntry{
			ID:    section.ID,
			Level: section.Level,
			Title: section.Title,
			Line:  section.StartLine,
		}

		*toc = append(*toc, entry)

		// Add children
		if maxDepth == 0 || section.Level < maxDepth {
			sm.buildTocRecursive(section.Children, maxDepth, toc)
		}
	}
}

// TocEntry represents a table of contents entry
type TocEntry struct {
	ID    string `json:"id"`
	Level int    `json:"level"`
	Title string `json:"title"`
	Line  int    `json:"line"`
}
