package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mosaan/mdatlas/internal/core"
)

// Server represents the MCP server
type Server struct {
	baseDir          string
	accessControl    *core.AccessControl
	structureManager *core.StructureManager
	toolHandler      *ToolHandler
	resourceHandler  *ResourceHandler
	cache            *core.Cache
}

// NewServer creates a new MCP server instance
func NewServer(baseDir string) (*Server, error) {
	// Create access control
	accessControl, err := core.NewAccessControl(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create access control: %w", err)
	}

	// Create cache
	cache := core.NewCache(100, 30*time.Minute)

	// Create structure manager
	structureManager := core.NewStructureManager(cache)

	// Create handlers
	toolHandler := NewToolHandler(structureManager, accessControl)
	resourceHandler := NewResourceHandler(accessControl)

	return &Server{
		baseDir:          baseDir,
		accessControl:    accessControl,
		structureManager: structureManager,
		toolHandler:      toolHandler,
		resourceHandler:  resourceHandler,
		cache:            cache,
	}, nil
}

// Run starts the MCP server
func (s *Server) Run(ctx context.Context) error {
	// Create JSON decoder and encoder for STDIO
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	// Send server info to stderr for debugging
	fmt.Fprintf(os.Stderr, "MCP Server started with base directory: %s\n", s.baseDir)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Read request
			var request MCPRequest
			if err := decoder.Decode(&request); err != nil {
				if err == io.EOF {
					return nil // Clean shutdown
				}

				// Send error response if we can parse the ID
				response := CreateErrorResponse(nil, ParseError, "Failed to parse request", err.Error())
				if encodeErr := encoder.Encode(response); encodeErr != nil {
					fmt.Fprintf(os.Stderr, "Failed to encode error response: %v\n", encodeErr)
				}
				continue
			}

			// Handle request
			response := s.handleRequest(request)

			// Send response
			if err := encoder.Encode(response); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to encode response: %v\n", err)
			}
		}
	}
}

// handleRequest handles an MCP request
func (s *Server) handleRequest(req MCPRequest) MCPResponse {
	// Validate request
	if err := ValidateRequest(req); err != nil {
		return CreateErrorResponse(GetRequestID(req), InvalidRequest, err.Error(), nil)
	}

	// Handle different methods
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	case "resources/list":
		return s.handleResourcesList(req)
	case "resources/read":
		return s.handleResourcesRead(req)
	case "ping":
		return s.handlePing(req)
	default:
		return CreateErrorResponse(GetRequestID(req), MethodNotFound, fmt.Sprintf("Method not found: %s", req.Method), nil)
	}
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(req MCPRequest) MCPResponse {
	var params InitializeParams
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return CreateErrorResponse(GetRequestID(req), InvalidParams, "Invalid initialize parameters", err.Error())
		}
	}

	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
			Resources: &ResourcesCapability{
				Subscribe:   false,
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    "mdatlas",
			Version: "1.0.0",
		},
	}

	return CreateSuccessResponse(GetRequestID(req), result)
}

// handleToolsList handles the tools/list request
func (s *Server) handleToolsList(req MCPRequest) MCPResponse {
	tools := s.toolHandler.GetAvailableTools()

	result := map[string]interface{}{
		"tools": tools,
	}

	return CreateSuccessResponse(GetRequestID(req), result)
}

// handleToolsCall handles the tools/call request
func (s *Server) handleToolsCall(req MCPRequest) MCPResponse {
	toolParams, err := ParseToolCallParams(req.Params)
	if err != nil {
		return CreateErrorResponse(GetRequestID(req), InvalidParams, err.Error(), nil)
	}

	// Execute tool
	result := s.toolHandler.HandleToolCall(toolParams.Name, toolParams.Arguments)

	return CreateSuccessResponse(GetRequestID(req), result)
}

// handleResourcesList handles the resources/list request
func (s *Server) handleResourcesList(req MCPRequest) MCPResponse {
	resources, err := s.resourceHandler.GetAvailableResources()
	if err != nil {
		return CreateErrorResponse(GetRequestID(req), InternalError, "Failed to list resources", err.Error())
	}

	result := ResourceListResult{
		Resources: resources,
	}

	return CreateSuccessResponse(GetRequestID(req), result)
}

// handleResourcesRead handles the resources/read request
func (s *Server) handleResourcesRead(req MCPRequest) MCPResponse {
	readParams, err := ParseResourceReadParams(req.Params)
	if err != nil {
		return CreateErrorResponse(GetRequestID(req), InvalidParams, err.Error(), nil)
	}

	result, err := s.resourceHandler.ReadResource(readParams.URI)
	if err != nil {
		return CreateErrorResponse(GetRequestID(req), InternalError, "Failed to read resource", err.Error())
	}

	return CreateSuccessResponse(GetRequestID(req), result)
}

// handlePing handles the ping request
func (s *Server) handlePing(req MCPRequest) MCPResponse {
	return CreateSuccessResponse(GetRequestID(req), map[string]string{"status": "pong"})
}

// RunInteractive runs the server in interactive mode for testing
func (s *Server) RunInteractive(ctx context.Context) error {
	fmt.Println("MCP Server Interactive Mode")
	fmt.Println("Type 'help' for available commands, 'quit' to exit")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fmt.Print("> ")

			if !scanner.Scan() {
				return nil
			}

			line := scanner.Text()
			if line == "quit" || line == "exit" {
				return nil
			}

			s.handleInteractiveCommand(line)
		}
	}
}

// handleInteractiveCommand handles interactive commands
func (s *Server) handleInteractiveCommand(command string) {
	switch command {
	case "help":
		fmt.Println("Available commands:")
		fmt.Println("  help        - Show this help message")
		fmt.Println("  status      - Show server status")
		fmt.Println("  tools       - List available tools")
		fmt.Println("  resources   - List available resources")
		fmt.Println("  cache       - Show cache statistics")
		fmt.Println("  quit/exit   - Exit interactive mode")

	case "status":
		fmt.Printf("Base directory: %s\n", s.baseDir)
		fmt.Printf("Cache size: %d entries\n", s.cache.Size())

	case "tools":
		tools := s.toolHandler.GetAvailableTools()
		fmt.Printf("Available tools (%d):\n", len(tools))
		for _, tool := range tools {
			fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
		}

	case "resources":
		resources, err := s.resourceHandler.GetAvailableResources()
		if err != nil {
			fmt.Printf("Error listing resources: %v\n", err)
			return
		}
		fmt.Printf("Available resources (%d):\n", len(resources))
		for _, resource := range resources {
			fmt.Printf("  - %s: %s\n", resource.URI, resource.Name)
		}

	case "cache":
		stats := s.cache.Stats()
		fmt.Printf("Cache statistics:\n")
		fmt.Printf("  Size: %d/%d entries\n", stats.Size, stats.MaxSize)
		fmt.Printf("  TTL: %v\n", stats.TTL)
		if !stats.OldestEntry.IsZero() {
			fmt.Printf("  Oldest entry: %v\n", stats.OldestEntry)
		}
		if !stats.NewestEntry.IsZero() {
			fmt.Printf("  Newest entry: %v\n", stats.NewestEntry)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Type 'help' for available commands")
	}
}
