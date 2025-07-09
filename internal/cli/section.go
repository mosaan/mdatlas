package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mosaan/mdatlas/internal/core"
	"github.com/spf13/cobra"
)

var (
	sectionID       string
	includeChildren bool
	format          string
)

// sectionCmd represents the section command
var sectionCmd = &cobra.Command{
	Use:   "section <file>",
	Short: "Extract content from a specific section of a Markdown file",
	Long: `Extract and display the content of a specific section from a Markdown file.
Use the section ID obtained from the structure command to retrieve the content.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		if sectionID == "" {
			return fmt.Errorf("section ID is required (use --section-id flag)")
		}

		// Resolve absolute path
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("failed to resolve file path: %w", err)
		}

		// Check if file exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filePath)
		}

		// Read file content
		content, err := os.ReadFile(absPath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		// Get section content
		parser := core.NewParser()
		sectionContent, err := parser.GetSectionContent(content, sectionID, includeChildren)
		if err != nil {
			return fmt.Errorf("failed to get section content: %w", err)
		}

		// Set the requested format
		sectionContent.Format = format

		// Output based on format
		switch format {
		case "json":
			encoder := json.NewEncoder(os.Stdout)
			if pretty {
				encoder.SetIndent("", "  ")
			}
			return encoder.Encode(sectionContent)
		case "plain":
			fmt.Print(sectionContent.Content)
			return nil
		case "markdown":
			fmt.Print(sectionContent.Content)
			return nil
		default:
			return fmt.Errorf("unsupported format: %s", format)
		}
	},
}

func init() {
	sectionCmd.Flags().StringVar(&sectionID, "section-id", "", "Section ID to retrieve (required)")
	sectionCmd.Flags().BoolVar(&includeChildren, "include-children", false, "Include child sections in the output")
	sectionCmd.Flags().StringVar(&format, "format", "markdown", "Output format (json, markdown, plain)")
	sectionCmd.Flags().BoolVar(&pretty, "pretty", false, "Pretty print JSON output (only for json format)")

	// Mark section-id as required
	sectionCmd.MarkFlagRequired("section-id")
}
