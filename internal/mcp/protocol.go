package mcp

import (
	"encoding/json"
	"fmt"
)

// MCPRequest represents an MCP request message
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCPResponse represents an MCP response message
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCPNotification represents an MCP notification
type MCPNotification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Tool call parameters
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Resource list parameters
type ResourceListParams struct {
	Cursor string `json:"cursor,omitempty"`
}

// Resource read parameters
type ResourceReadParams struct {
	URI string `json:"uri"`
}

// Tool definition
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// Resource definition
type Resource struct {
	URI         string      `json:"uri"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	MimeType    string      `json:"mimeType,omitempty"`
	Metadata    interface{} `json:"metadata,omitempty"`
}

// Tool call result
type ToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content block
type Content struct {
	Type     string      `json:"type"`
	Text     string      `json:"text,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	MimeType string      `json:"mimeType,omitempty"`
}

// Resource list result
type ResourceListResult struct {
	Resources []Resource `json:"resources"`
	NextCursor string    `json:"nextCursor,omitempty"`
}

// Resource read result
type ResourceReadResult struct {
	Contents []Content `json:"contents"`
}

// Server capabilities
type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
}

// Tools capability
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// Resources capability
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// Initialize request parameters
type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

// Client capabilities
type ClientCapabilities struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

// Roots capability
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// Sampling capability
type SamplingCapability struct{}

// Client info
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Initialize result
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// Server info
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Standard error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// CreateErrorResponse creates an error response
func CreateErrorResponse(id interface{}, code int, message string, data interface{}) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// CreateSuccessResponse creates a success response
func CreateSuccessResponse(id interface{}, result interface{}) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// CreateNotification creates a notification
func CreateNotification(method string, params interface{}) MCPNotification {
	var rawParams json.RawMessage
	if params != nil {
		if data, err := json.Marshal(params); err == nil {
			rawParams = data
		}
	}
	
	return MCPNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  rawParams,
	}
}

// ValidateRequest validates an MCP request
func ValidateRequest(req MCPRequest) error {
	if req.JSONRPC != "2.0" {
		return fmt.Errorf("invalid JSON-RPC version: %s", req.JSONRPC)
	}
	
	if req.Method == "" {
		return fmt.Errorf("missing method")
	}
	
	return nil
}

// ParseToolCallParams parses tool call parameters
func ParseToolCallParams(params json.RawMessage) (*ToolCallParams, error) {
	var toolParams ToolCallParams
	if err := json.Unmarshal(params, &toolParams); err != nil {
		return nil, fmt.Errorf("failed to parse tool call params: %w", err)
	}
	
	if toolParams.Name == "" {
		return nil, fmt.Errorf("missing tool name")
	}
	
	return &toolParams, nil
}

// ParseResourceListParams parses resource list parameters
func ParseResourceListParams(params json.RawMessage) (*ResourceListParams, error) {
	var listParams ResourceListParams
	if len(params) > 0 {
		if err := json.Unmarshal(params, &listParams); err != nil {
			return nil, fmt.Errorf("failed to parse resource list params: %w", err)
		}
	}
	
	return &listParams, nil
}

// ParseResourceReadParams parses resource read parameters
func ParseResourceReadParams(params json.RawMessage) (*ResourceReadParams, error) {
	var readParams ResourceReadParams
	if err := json.Unmarshal(params, &readParams); err != nil {
		return nil, fmt.Errorf("failed to parse resource read params: %w", err)
	}
	
	if readParams.URI == "" {
		return nil, fmt.Errorf("missing resource URI")
	}
	
	return &readParams, nil
}

// CreateTextContent creates a text content block
func CreateTextContent(text string) Content {
	return Content{
		Type: "text",
		Text: text,
	}
}

// CreateJSONContent creates a JSON content block
func CreateJSONContent(data interface{}) Content {
	return Content{
		Type:     "text",
		Text:     marshalJSON(data),
		MimeType: "application/json",
	}
}

// marshalJSON marshals data to JSON string
func marshalJSON(data interface{}) string {
	if jsonData, err := json.MarshalIndent(data, "", "  "); err == nil {
		return string(jsonData)
	}
	return fmt.Sprintf("%+v", data)
}

// IsNotification checks if a message is a notification (no ID)
func IsNotification(req MCPRequest) bool {
	return req.ID == nil
}

// GetRequestID safely gets the request ID
func GetRequestID(req MCPRequest) interface{} {
	if req.ID == nil {
		return nil
	}
	return req.ID
}