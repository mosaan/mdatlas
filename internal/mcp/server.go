package mcp

import (
	"context"
	"fmt"
)

// Server represents the MCP server
type Server struct {
	baseDir string
}

// NewServer creates a new MCP server instance
func NewServer(baseDir string) (*Server, error) {
	return &Server{
		baseDir: baseDir,
	}, nil
}

// Run starts the MCP server
func (s *Server) Run(ctx context.Context) error {
	// TODO: Implement MCP server functionality
	// For now, return an error indicating it's not implemented
	return fmt.Errorf("MCP server functionality not yet implemented")
}