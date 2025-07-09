package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/mosaan/mdatlas/internal/mcp"
)

var (
	baseDir   string
	mcpServer bool
	version   string = "dev"
	buildDate string = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mdatlas",
	Short: "A Model Context Protocol server for Markdown document structure analysis",
	Long: `mdatlas provides efficient access to Markdown document structure information,
allowing AI models to selectively retrieve specific sections without loading entire files.

By default, mdatlas runs as an MCP server using STDIO for communication.
Use the subcommands for CLI-based operations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if mcpServer {
			return runMCPServer(baseDir)
		}
		
		// If no subcommand is provided, show help
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&baseDir, "base-dir", ".", "Base directory for file access")
	rootCmd.PersistentFlags().BoolVar(&mcpServer, "mcp-server", false, "Run as MCP server (STDIO mode)")
	
	// Add subcommands
	rootCmd.AddCommand(structureCmd)
	rootCmd.AddCommand(sectionCmd)
	rootCmd.AddCommand(versionCmd)
}

// runMCPServer starts the MCP server
func runMCPServer(baseDir string) error {
	server, err := mcp.NewServer(baseDir)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}
	
	return server.Run(context.Background())
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mdatlas version %s\n", version)
		fmt.Printf("Build date: %s\n", buildDate)
	},
}