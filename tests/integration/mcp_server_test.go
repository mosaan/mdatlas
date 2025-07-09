package integration

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// MCPRequest represents an MCP request for testing
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCPResponse represents an MCP response for testing
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error for testing
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func TestMCPServerInitialization(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	// Start MCP server
	cmd := exec.Command(binaryPath, "--mcp-server", "--base-dir", filepath.Join(projectRoot, "tests", "fixtures"))
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}
	
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start MCP server: %v", err)
	}
	
	// Send initialize request
	initRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}`),
	}
	
	if err := json.NewEncoder(stdin).Encode(initRequest); err != nil {
		t.Fatalf("Failed to send initialize request: %v", err)
	}
	
	// Read response
	var response MCPResponse
	if err := json.NewDecoder(stdout).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	// Verify response
	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}
	
	if response.ID != 1.0 {
		t.Errorf("Expected ID 1, got %v", response.ID)
	}
	
	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}
	
	// Verify result structure
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be an object")
	}
	
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocol version 2024-11-05, got %v", result["protocolVersion"])
	}
	
	// Clean up
	stdin.Close()
	cmd.Process.Kill()
	cmd.Wait()
}

func TestMCPServerToolsList(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	// Start MCP server and get tools list
	response := sendMCPRequest(t, projectRoot, binaryPath, MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	})
	
	// Verify response
	if response.Error != nil {
		t.Fatalf("Expected no error, got %v", response.Error)
	}
	
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be an object")
	}
	
	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatalf("Expected tools to be an array")
	}
	
	// Check that we have the expected tools
	expectedTools := []string{
		"get_markdown_structure",
		"get_markdown_section", 
		"search_markdown_content",
		"get_markdown_stats",
		"get_markdown_toc",
	}
	
	toolNames := make([]string, len(tools))
	for i, tool := range tools {
		toolMap := tool.(map[string]interface{})
		toolNames[i] = toolMap["name"].(string)
	}
	
	for _, expectedTool := range expectedTools {
		found := false
		for _, toolName := range toolNames {
			if toolName == expectedTool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tool %s not found in tools list", expectedTool)
		}
	}
}

func TestMCPServerToolsCall(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	tests := []struct {
		name       string
		toolName   string
		args       map[string]interface{}
		expectError bool
		validate   func(t *testing.T, result interface{})
	}{
		{
			name:     "get_markdown_structure",
			toolName: "get_markdown_structure",
			args: map[string]interface{}{
				"file_path": "sample.md",
			},
			expectError: false,
			validate: func(t *testing.T, result interface{}) {
				toolResult := result.(map[string]interface{})
				content := toolResult["content"].([]interface{})
				if len(content) == 0 {
					t.Error("Expected content in tool result")
				}
				
				firstContent := content[0].(map[string]interface{})
				if firstContent["type"] != "text" {
					t.Error("Expected content type to be 'text'")
				}
				
				// Parse the JSON content
				textContent := firstContent["text"].(string)
				var structure map[string]interface{}
				if err := json.Unmarshal([]byte(textContent), &structure); err != nil {
					t.Fatalf("Failed to parse structure JSON: %v. Content: %s", err, textContent)
				}
				
				if structure["file_path"] == "" {
					t.Error("Expected file_path in structure")
				}
			},
		},
		{
			name:     "get_markdown_structure with max_depth",
			toolName: "get_markdown_structure",
			args: map[string]interface{}{
				"file_path": "sample.md",
				"max_depth": 2,
			},
			expectError: false,
			validate: func(t *testing.T, result interface{}) {
				toolResult := result.(map[string]interface{})
				content := toolResult["content"].([]interface{})
				if len(content) == 0 {
					t.Error("Expected content in tool result")
				}
			},
		},
		{
			name:     "search_markdown_content",
			toolName: "search_markdown_content",
			args: map[string]interface{}{
				"file_path": "sample.md",
				"query":     "Introduction",
			},
			expectError: false,
			validate: func(t *testing.T, result interface{}) {
				toolResult := result.(map[string]interface{})
				content := toolResult["content"].([]interface{})
				if len(content) == 0 {
					t.Error("Expected content in tool result")
				}
				
				// Parse the JSON content
				firstContent := content[0].(map[string]interface{})
				var searchResult map[string]interface{}
				if err := json.Unmarshal([]byte(firstContent["text"].(string)), &searchResult); err != nil {
					t.Fatalf("Failed to parse search result JSON: %v", err)
				}
				
				if searchResult["query"] != "Introduction" {
					t.Error("Expected query to be 'Introduction'")
				}
			},
		},
		{
			name:     "get_markdown_stats",
			toolName: "get_markdown_stats",
			args: map[string]interface{}{
				"file_path": "sample.md",
			},
			expectError: false,
			validate: func(t *testing.T, result interface{}) {
				toolResult := result.(map[string]interface{})
				content := toolResult["content"].([]interface{})
				if len(content) == 0 {
					t.Error("Expected content in tool result")
				}
				
				// Parse the JSON content
				firstContent := content[0].(map[string]interface{})
				var stats map[string]interface{}
				if err := json.Unmarshal([]byte(firstContent["text"].(string)), &stats); err != nil {
					t.Fatalf("Failed to parse stats JSON: %v", err)
				}
				
				if stats["total_chars"] == nil {
					t.Error("Expected total_chars in stats")
				}
			},
		},
		{
			name:     "get_markdown_toc",
			toolName: "get_markdown_toc",
			args: map[string]interface{}{
				"file_path": "sample.md",
			},
			expectError: false,
			validate: func(t *testing.T, result interface{}) {
				toolResult := result.(map[string]interface{})
				content := toolResult["content"].([]interface{})
				if len(content) == 0 {
					t.Error("Expected content in tool result")
				}
				
				// Parse the JSON content
				firstContent := content[0].(map[string]interface{})
				var toc map[string]interface{}
				if err := json.Unmarshal([]byte(firstContent["text"].(string)), &toc); err != nil {
					t.Fatalf("Failed to parse TOC JSON: %v", err)
				}
				
				if toc["toc"] == nil {
					t.Error("Expected toc in result")
				}
			},
		},
		{
			name:     "invalid tool",
			toolName: "invalid_tool",
			args: map[string]interface{}{
				"file_path": "sample.md",
			},
			expectError: true,
			validate: func(t *testing.T, result interface{}) {
				toolResult := result.(map[string]interface{})
				if toolResult["isError"] != true {
					t.Error("Expected isError to be true for invalid tool")
				}
			},
		},
		{
			name:     "missing file_path",
			toolName: "get_markdown_structure",
			args:     map[string]interface{}{},
			expectError: true,
			validate: func(t *testing.T, result interface{}) {
				toolResult := result.(map[string]interface{})
				if toolResult["isError"] != true {
					t.Error("Expected isError to be true for missing file_path")
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"name":      tt.toolName,
				"arguments": tt.args,
			}
			
			paramsJSON, _ := json.Marshal(params)
			response := sendMCPRequest(t, projectRoot, binaryPath, MCPRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "tools/call",
				Params:  paramsJSON,
			})
			
			if tt.expectError {
				// For tool errors, the MCP response should be successful but the tool result should indicate error
				if response.Error != nil {
					t.Errorf("Expected no MCP error, got %v", response.Error)
				}
			} else {
				if response.Error != nil {
					t.Errorf("Expected no error, got %v", response.Error)
				}
			}
			
			if tt.validate != nil {
				tt.validate(t, response.Result)
			}
		})
	}
}

func TestMCPServerResourcesList(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	response := sendMCPRequest(t, projectRoot, binaryPath, MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "resources/list",
	})
	
	if response.Error != nil {
		t.Fatalf("Expected no error, got %v", response.Error)
	}
	
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be an object")
	}
	
	resources, ok := result["resources"].([]interface{})
	if !ok {
		t.Fatalf("Expected resources to be an array")
	}
	
	// Should have resources for our test files
	if len(resources) == 0 {
		t.Error("Expected at least one resource")
	}
	
	// Check that resources have proper structure
	for _, resource := range resources {
		resourceMap := resource.(map[string]interface{})
		if resourceMap["uri"] == "" {
			t.Error("Expected resource to have URI")
		}
		if resourceMap["name"] == "" {
			t.Error("Expected resource to have name")
		}
	}
}

func TestMCPServerResourcesRead(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	tests := []struct {
		name        string
		uri         string
		expectError bool
		validate    func(t *testing.T, result interface{})
	}{
		{
			name:        "structure resource",
			uri:         "markdown://file/sample.md/structure",
			expectError: false,
			validate: func(t *testing.T, result interface{}) {
				resourceResult := result.(map[string]interface{})
				contents := resourceResult["contents"].([]interface{})
				if len(contents) == 0 {
					t.Error("Expected contents in resource result")
				}
				
				firstContent := contents[0].(map[string]interface{})
				if firstContent["type"] != "text" {
					t.Error("Expected content type to be 'text'")
				}
				
				// Parse the JSON content
				var structure map[string]interface{}
				if err := json.Unmarshal([]byte(firstContent["text"].(string)), &structure); err != nil {
					t.Fatalf("Failed to parse structure JSON: %v", err)
				}
				
				if structure["file_path"] == "" {
					t.Error("Expected file_path in structure")
				}
			},
		},
		{
			name:        "content resource",
			uri:         "markdown://file/sample.md/content",
			expectError: false,
			validate: func(t *testing.T, result interface{}) {
				resourceResult := result.(map[string]interface{})
				contents := resourceResult["contents"].([]interface{})
				if len(contents) == 0 {
					t.Error("Expected contents in resource result")
				}
				
				firstContent := contents[0].(map[string]interface{})
				if firstContent["type"] != "text" {
					t.Error("Expected content type to be 'text'")
				}
				
				text := firstContent["text"].(string)
				if !strings.Contains(text, "#") {
					t.Error("Expected markdown content to contain headers")
				}
			},
		},
		{
			name:        "invalid resource",
			uri:         "markdown://file/nonexistent.md/structure",
			expectError: true,
			validate: func(t *testing.T, result interface{}) {
				// Should be handled as an error
			},
		},
		{
			name:        "invalid URI format",
			uri:         "invalid://uri",
			expectError: true,
			validate: func(t *testing.T, result interface{}) {
				// Should be handled as an error
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"uri": tt.uri,
			}
			
			paramsJSON, _ := json.Marshal(params)
			response := sendMCPRequest(t, projectRoot, binaryPath, MCPRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "resources/read",
				Params:  paramsJSON,
			})
			
			if tt.expectError {
				if response.Error == nil {
					t.Error("Expected error for invalid resource")
				}
			} else {
				if response.Error != nil {
					t.Errorf("Expected no error, got %v", response.Error)
				}
				
				if tt.validate != nil {
					tt.validate(t, response.Result)
				}
			}
		})
	}
}

func TestMCPServerPing(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	response := sendMCPRequest(t, projectRoot, binaryPath, MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "ping",
	})
	
	if response.Error != nil {
		t.Fatalf("Expected no error, got %v", response.Error)
	}
	
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be an object")
	}
	
	if result["status"] != "pong" {
		t.Errorf("Expected status 'pong', got %v", result["status"])
	}
}

func TestMCPServerErrorHandling(t *testing.T) {
	projectRoot, binaryPath := setupTest(t)
	
	tests := []struct {
		name     string
		request  MCPRequest
		expectError bool
		errorCode   int
	}{
		{
			name: "invalid JSON-RPC version",
			request: MCPRequest{
				JSONRPC: "1.0",
				ID:      1,
				Method:  "ping",
			},
			expectError: true,
			errorCode:   -32600,
		},
		{
			name: "missing method",
			request: MCPRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "",
			},
			expectError: true,
			errorCode:   -32600,
		},
		{
			name: "unknown method",
			request: MCPRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "unknown_method",
			},
			expectError: true,
			errorCode:   -32601,
		},
		{
			name: "invalid parameters",
			request: MCPRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "tools/call",
				Params:  json.RawMessage(`{"invalid": "params"}`),
			},
			expectError: true,
			errorCode:   -32602,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := sendMCPRequest(t, projectRoot, binaryPath, tt.request)
			
			if tt.expectError {
				if response.Error == nil {
					t.Error("Expected error but got none")
				} else {
					if response.Error.Code != tt.errorCode {
						t.Errorf("Expected error code %d, got %d", tt.errorCode, response.Error.Code)
					}
				}
			} else {
				if response.Error != nil {
					t.Errorf("Expected no error, got %v", response.Error)
				}
			}
		})
	}
}

// Helper function to send MCP request and get response
func sendMCPRequest(t *testing.T, projectRoot, binaryPath string, request MCPRequest) MCPResponse {
	// Start MCP server
	cmd := exec.Command(binaryPath, "--mcp-server", "--base-dir", filepath.Join(projectRoot, "tests", "fixtures"))
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}
	
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start MCP server: %v", err)
	}
	
	// Read stderr in background to prevent blocking
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			// Log stderr for debugging if needed
			// t.Logf("MCP Server stderr: %s", scanner.Text())
		}
	}()
	
	// Send request
	if err := json.NewEncoder(stdin).Encode(request); err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	
	// Read response with timeout
	responseChan := make(chan MCPResponse, 1)
	errorChan := make(chan error, 1)
	
	go func() {
		var response MCPResponse
		if err := json.NewDecoder(stdout).Decode(&response); err != nil {
			errorChan <- err
			return
		}
		responseChan <- response
	}()
	
	select {
	case response := <-responseChan:
		// Clean up
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
		return response
	case err := <-errorChan:
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
		t.Fatalf("Failed to decode response: %v", err)
		return MCPResponse{}
	case <-time.After(5 * time.Second):
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
		t.Fatal("Timeout waiting for response")
		return MCPResponse{}
	}
}