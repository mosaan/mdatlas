package mcp

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mosaan/mdatlas/internal/core"
	"github.com/mosaan/mdatlas/pkg/types"
)

// ToolHandler handles MCP tool calls
type ToolHandler struct {
	structureManager *core.StructureManager
	accessControl    *core.AccessControl
}

// NewToolHandler creates a new tool handler
func NewToolHandler(structureManager *core.StructureManager, accessControl *core.AccessControl) *ToolHandler {
	return &ToolHandler{
		structureManager: structureManager,
		accessControl:    accessControl,
	}
}

// GetAvailableTools returns the list of available tools
func (th *ToolHandler) GetAvailableTools() []Tool {
	return []Tool{
		{
			Name:        "get_markdown_structure",
			Description: "Extract hierarchical structure from a Markdown file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the Markdown file (relative to base directory)",
					},
					"max_depth": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum heading depth to include (optional)",
						"minimum":     1,
						"maximum":     6,
					},
				},
				"required": []string{"file_path"},
			},
		},
		{
			Name:        "get_markdown_section",
			Description: "Retrieve content from a specific section of a Markdown file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the Markdown file (relative to base directory)",
					},
					"section_id": map[string]interface{}{
						"type":        "string",
						"description": "Unique identifier of the section to retrieve",
					},
					"include_children": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to include child sections in the content",
						"default":     false,
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "Output format for the content",
						"enum":        []string{"markdown", "plain"},
						"default":     "markdown",
					},
				},
				"required": []string{"file_path", "section_id"},
			},
		},
		{
			Name:        "search_markdown_content",
			Description: "Search for sections containing specific text in a Markdown file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the Markdown file (relative to base directory)",
					},
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query to find in section titles",
					},
					"case_sensitive": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether the search should be case sensitive",
						"default":     false,
					},
				},
				"required": []string{"file_path", "query"},
			},
		},
		{
			Name:        "get_markdown_stats",
			Description: "Get statistics about a Markdown document",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the Markdown file (relative to base directory)",
					},
				},
				"required": []string{"file_path"},
			},
		},
		{
			Name:        "get_markdown_toc",
			Description: "Generate a table of contents for a Markdown document",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the Markdown file (relative to base directory)",
					},
					"max_depth": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum heading depth to include in TOC",
						"minimum":     1,
						"maximum":     6,
					},
				},
				"required": []string{"file_path"},
			},
		},
	}
}

// HandleToolCall handles a specific tool call
func (th *ToolHandler) HandleToolCall(toolName string, arguments map[string]interface{}) ToolResult {
	switch toolName {
	case "get_markdown_structure":
		return th.handleGetMarkdownStructure(arguments)
	case "get_markdown_section":
		return th.handleGetMarkdownSection(arguments)
	case "search_markdown_content":
		return th.handleSearchMarkdownContent(arguments)
	case "get_markdown_stats":
		return th.handleGetMarkdownStats(arguments)
	case "get_markdown_toc":
		return th.handleGetMarkdownTOC(arguments)
	default:
		return ToolResult{
			Content: []Content{CreateTextContent(fmt.Sprintf("Unknown tool: %s", toolName))},
			IsError: true,
		}
	}
}

// handleGetMarkdownStructure handles the get_markdown_structure tool
func (th *ToolHandler) handleGetMarkdownStructure(args map[string]interface{}) ToolResult {
	filePath, ok := args["file_path"].(string)
	if !ok {
		return th.createErrorResult("Missing or invalid file_path parameter")
	}

	// Validate file access
	validPath, err := th.accessControl.ValidatePath(filePath)
	if err != nil {
		return th.createErrorResult(fmt.Sprintf("Access denied: %v", err))
	}

	// Get document structure
	structure, err := th.structureManager.GetDocumentStructure(validPath)
	if err != nil {
		return th.createErrorResult(fmt.Sprintf("Failed to get structure: %v", err))
	}

	// Apply max depth filter if specified
	if maxDepthRaw, exists := args["max_depth"]; exists {
		if maxDepth, ok := maxDepthRaw.(float64); ok {
			structure.Structure = th.filterByDepth(structure.Structure, int(maxDepth))
		}
	}

	return ToolResult{
		Content: []Content{CreateJSONContent(structure)},
	}
}

// handleGetMarkdownSection handles the get_markdown_section tool
func (th *ToolHandler) handleGetMarkdownSection(args map[string]interface{}) ToolResult {
	filePath, ok := args["file_path"].(string)
	if !ok {
		return th.createErrorResult("Missing or invalid file_path parameter")
	}

	sectionID, ok := args["section_id"].(string)
	if !ok {
		return th.createErrorResult("Missing or invalid section_id parameter")
	}

	// Validate file access
	validPath, err := th.accessControl.ValidatePath(filePath)
	if err != nil {
		return th.createErrorResult(fmt.Sprintf("Access denied: %v", err))
	}

	// Get optional parameters
	includeChildren := false
	if include, exists := args["include_children"]; exists {
		if b, ok := include.(bool); ok {
			includeChildren = b
		}
	}

	format := "markdown"
	if f, exists := args["format"]; exists {
		if s, ok := f.(string); ok {
			format = s
		}
	}

	// Get section content
	sectionContent, err := th.structureManager.GetSectionContent(validPath, sectionID, includeChildren)
	if err != nil {
		return th.createErrorResult(fmt.Sprintf("Failed to get section: %v", err))
	}

	// Set format
	sectionContent.Format = format

	// Return based on format
	switch format {
	case "json":
		return ToolResult{
			Content: []Content{CreateJSONContent(sectionContent)},
		}
	default:
		return ToolResult{
			Content: []Content{CreateTextContent(sectionContent.Content)},
		}
	}
}

// handleSearchMarkdownContent handles the search_markdown_content tool
func (th *ToolHandler) handleSearchMarkdownContent(args map[string]interface{}) ToolResult {
	filePath, ok := args["file_path"].(string)
	if !ok {
		return th.createErrorResult("Missing or invalid file_path parameter")
	}

	query, ok := args["query"].(string)
	if !ok {
		return th.createErrorResult("Missing or invalid query parameter")
	}

	// Validate file access
	validPath, err := th.accessControl.ValidatePath(filePath)
	if err != nil {
		return th.createErrorResult(fmt.Sprintf("Access denied: %v", err))
	}

	// Get case sensitivity setting
	caseSensitive := false
	if cs, exists := args["case_sensitive"]; exists {
		if b, ok := cs.(bool); ok {
			caseSensitive = b
		}
	}

	// Search sections
	sections, err := th.structureManager.SearchSections(validPath, query, caseSensitive)
	if err != nil {
		return th.createErrorResult(fmt.Sprintf("Search failed: %v", err))
	}

	searchResult := map[string]interface{}{
		"file_path": filePath,
		"query":     query,
		"results":   sections,
		"count":     len(sections),
	}

	return ToolResult{
		Content: []Content{CreateJSONContent(searchResult)},
	}
}

// handleGetMarkdownStats handles the get_markdown_stats tool
func (th *ToolHandler) handleGetMarkdownStats(args map[string]interface{}) ToolResult {
	filePath, ok := args["file_path"].(string)
	if !ok {
		return th.createErrorResult("Missing or invalid file_path parameter")
	}

	// Validate file access
	validPath, err := th.accessControl.ValidatePath(filePath)
	if err != nil {
		return th.createErrorResult(fmt.Sprintf("Access denied: %v", err))
	}

	// Get document statistics
	stats, err := th.structureManager.GetDocumentStats(validPath)
	if err != nil {
		return th.createErrorResult(fmt.Sprintf("Failed to get stats: %v", err))
	}

	return ToolResult{
		Content: []Content{CreateJSONContent(stats)},
	}
}

// handleGetMarkdownTOC handles the get_markdown_toc tool
func (th *ToolHandler) handleGetMarkdownTOC(args map[string]interface{}) ToolResult {
	filePath, ok := args["file_path"].(string)
	if !ok {
		return th.createErrorResult("Missing or invalid file_path parameter")
	}

	// Validate file access
	validPath, err := th.accessControl.ValidatePath(filePath)
	if err != nil {
		return th.createErrorResult(fmt.Sprintf("Access denied: %v", err))
	}

	// Get max depth
	maxDepth := 0
	if maxDepthRaw, exists := args["max_depth"]; exists {
		if md, ok := maxDepthRaw.(float64); ok {
			maxDepth = int(md)
		}
	}

	// Generate table of contents
	toc, err := th.structureManager.GetTableOfContents(validPath, maxDepth)
	if err != nil {
		return th.createErrorResult(fmt.Sprintf("Failed to generate TOC: %v", err))
	}

	tocResult := map[string]interface{}{
		"file_path": filePath,
		"toc":       toc,
		"count":     len(toc),
	}

	return ToolResult{
		Content: []Content{CreateJSONContent(tocResult)},
	}
}

// filterByDepth filters sections by maximum depth
func (th *ToolHandler) filterByDepth(sections []types.Section, maxDepth int) []types.Section {
	if maxDepth <= 0 {
		return sections
	}

	var filtered []types.Section
	for _, section := range sections {
		if section.Level <= maxDepth {
			filteredSection := section
			filteredSection.Children = th.filterByDepth(section.Children, maxDepth)
			filtered = append(filtered, filteredSection)
		}
	}
	return filtered
}

// createErrorResult creates an error tool result
func (th *ToolHandler) createErrorResult(message string) ToolResult {
	return ToolResult{
		Content: []Content{CreateTextContent(message)},
		IsError: true,
	}
}

// ResourceHandler handles MCP resource operations
type ResourceHandler struct {
	accessControl *core.AccessControl
}

// NewResourceHandler creates a new resource handler
func NewResourceHandler(accessControl *core.AccessControl) *ResourceHandler {
	return &ResourceHandler{
		accessControl: accessControl,
	}
}

// GetAvailableResources returns the list of available resources
func (rh *ResourceHandler) GetAvailableResources() ([]Resource, error) {
	files, err := rh.accessControl.ListAllowedFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	var resources []Resource
	for _, file := range files {
		// Create structure resource
		structureURI := fmt.Sprintf("markdown://file/%s/structure", file)
		resources = append(resources, Resource{
			URI:         structureURI,
			Name:        fmt.Sprintf("Structure of %s", filepath.Base(file)),
			Description: fmt.Sprintf("Hierarchical structure of %s", file),
			MimeType:    "application/json",
		})

		// Create content resource
		contentURI := fmt.Sprintf("markdown://file/%s/content", file)
		resources = append(resources, Resource{
			URI:         contentURI,
			Name:        fmt.Sprintf("Content of %s", filepath.Base(file)),
			Description: fmt.Sprintf("Full content of %s", file),
			MimeType:    "text/markdown",
		})
	}

	return resources, nil
}

// ReadResource reads a specific resource
func (rh *ResourceHandler) ReadResource(uri string) (ResourceReadResult, error) {
	// Parse URI
	parts := strings.Split(uri, "/")
	if len(parts) < 4 || parts[0] != "markdown:" || parts[1] != "" || parts[2] != "file" {
		return ResourceReadResult{}, fmt.Errorf("invalid resource URI: %s", uri)
	}

	// Extract file path and resource type
	filePath := strings.Join(parts[3:len(parts)-1], "/")
	resourceType := parts[len(parts)-1]

	// Validate file access
	validPath, err := rh.accessControl.ValidatePath(filePath)
	if err != nil {
		return ResourceReadResult{}, fmt.Errorf("access denied: %w", err)
	}

	switch resourceType {
	case "structure":
		return rh.readStructureResource(validPath)
	case "content":
		return rh.readContentResource(validPath)
	default:
		return ResourceReadResult{}, fmt.Errorf("unknown resource type: %s", resourceType)
	}
}

// readStructureResource reads a structure resource
func (rh *ResourceHandler) readStructureResource(filePath string) (ResourceReadResult, error) {
	// Create structure manager
	cache := core.NewCache(100, 0) // No TTL for resources
	structureManager := core.NewStructureManager(cache)

	structure, err := structureManager.GetDocumentStructure(filePath)
	if err != nil {
		return ResourceReadResult{}, fmt.Errorf("failed to get structure: %w", err)
	}

	return ResourceReadResult{
		Contents: []Content{CreateJSONContent(structure)},
	}, nil
}

// readContentResource reads a content resource
func (rh *ResourceHandler) readContentResource(filePath string) (ResourceReadResult, error) {
	reader := core.NewSecureFileReader(rh.accessControl)
	content, err := reader.ReadFile(filePath)
	if err != nil {
		return ResourceReadResult{}, fmt.Errorf("failed to read file: %w", err)
	}

	return ResourceReadResult{
		Contents: []Content{CreateTextContent(string(content))},
	}, nil
}
