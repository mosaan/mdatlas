package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mosaan/mdatlas/internal/core"
	"github.com/mosaan/mdatlas/pkg/types"
	"github.com/spf13/cobra"
)

var (
	maxDepth int
	pretty   bool
)

// structureCmd represents the structure command
var structureCmd = &cobra.Command{
	Use:   "structure <file>",
	Short: "Extract structure information from Markdown file",
	Long: `Extract and display the hierarchical structure of a Markdown file.
This command analyzes the heading structure and provides metadata about
each section including character counts, line numbers, and nesting levels.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

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

		// Parse structure
		parser := core.NewParser()
		structure, err := parser.ParseStructure(content)
		if err != nil {
			return fmt.Errorf("failed to parse structure: %w", err)
		}

		// Set file path in structure
		structure.FilePath = absPath

		// Filter by max depth if specified
		if maxDepth > 0 {
			structure.Structure = filterByDepth(structure.Structure, maxDepth)
		}

		// Output JSON
		encoder := json.NewEncoder(os.Stdout)
		if pretty {
			encoder.SetIndent("", "  ")
		}

		return encoder.Encode(structure)
	},
}

func init() {
	structureCmd.Flags().IntVar(&maxDepth, "max-depth", 0, "Maximum heading depth to include (0 for all)")
	structureCmd.Flags().BoolVar(&pretty, "pretty", false, "Pretty print JSON output")
}

// filterByDepth filters sections by maximum depth
func filterByDepth(sections []types.Section, maxDepth int) []types.Section {
	if maxDepth <= 0 {
		return sections
	}

	var filtered []types.Section
	for _, section := range sections {
		if section.Level <= maxDepth {
			filteredSection := section
			filteredSection.Children = filterByDepth(section.Children, maxDepth)
			filtered = append(filtered, filteredSection)
		}
	}
	return filtered
}
