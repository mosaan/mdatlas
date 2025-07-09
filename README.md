# mdatlas

A Model Context Protocol (MCP) server for Markdown document structure analysis and navigation.

## Overview

mdatlas provides efficient access to Markdown document structure information, allowing AI models to selectively retrieve specific sections without loading entire files. This solves the problem of context window limitations when working with large Markdown documents.

## Features

- **Document Structure Analysis**: Extract hierarchical structure from Markdown files (H1-H6 headings)
- **Section-based Access**: Retrieve specific sections by unique ID
- **Metadata Extraction**: Get character counts, line numbers, and nesting information
- **Multiple Output Formats**: JSON, Markdown, and Plain text
- **CLI Interface**: Standalone command-line tool for direct usage
- **MCP Server**: STDIO-based server for AI model integration (planned)

## Installation

### Building from Source

```bash
# Clone the repository
git clone https://github.com/mosaan/mdatlas.git
cd mdatlas

# Build the binary
make build

# Or install to GOPATH/bin
make install
```

### Cross-platform Builds

```bash
# Build for all supported platforms
make release
```

This creates binaries for:
- Linux (amd64)
- macOS (amd64, arm64)
- Windows (amd64)

## Usage

### CLI Commands

#### Extract Document Structure

```bash
# Basic structure extraction
mdatlas structure document.md

# Pretty-printed JSON output
mdatlas structure document.md --pretty

# Limit heading depth
mdatlas structure document.md --max-depth 3
```

**Example output:**
```json
{
  "file_path": "/path/to/document.md",
  "total_chars": 15000,
  "total_lines": 500,
  "structure": [
    {
      "id": "section_48c2fc6ee5f4af76",
      "level": 1,
      "title": "Introduction",
      "char_count": 800,
      "line_count": 25,
      "start_line": 1,
      "end_line": 25,
      "children": [
        {
          "id": "section_a1b2c3d4e5f6g7h8",
          "level": 2,
          "title": "Background",
          "char_count": 400,
          "line_count": 12,
          "start_line": 5,
          "end_line": 16,
          "children": []
        }
      ]
    }
  ],
  "last_modified": "2025-07-09T14:00:00Z"
}
```

#### Extract Section Content

```bash
# Extract specific section by ID
mdatlas section document.md --section-id section_48c2fc6ee5f4af76

# Include child sections
mdatlas section document.md --section-id section_48c2fc6ee5f4af76 --include-children

# Different output formats
mdatlas section document.md --section-id section_48c2fc6ee5f4af76 --format json
mdatlas section document.md --section-id section_48c2fc6ee5f4af76 --format plain
```

#### Other Commands

```bash
# Show version information
mdatlas version

# Run as MCP server (not yet implemented)
mdatlas --mcp-server --base-dir /path/to/documents

# Show help
mdatlas --help
mdatlas structure --help
mdatlas section --help
```

### MCP Server Mode

*Note: MCP server functionality is planned but not yet implemented.*

When complete, the MCP server will provide:

- **Tools**:
  - `get_markdown_structure`: Extract document structure
  - `get_markdown_section`: Retrieve section content
  - `search_markdown_content`: Search within documents

- **Resources**:
  - `markdown://file/{file_path}/structure`: Document structure
  - `markdown://file/{file_path}/section/{section_id}`: Section content

## Development

### Prerequisites

- Go 1.22.2 or later
- Make (for build automation)

### Building

```bash
# Download dependencies
make deps

# Build the project
make build

# Run tests
make test

# Format code
make fmt

# Run linter (if golangci-lint is installed)
make lint

# Clean build artifacts
make clean
```

### Project Structure

```
mdatlas/
├── cmd/
│   └── mdatlas/
│       └── main.go              # Entry point
├── internal/
│   ├── core/
│   │   └── parser.go            # Markdown parsing logic
│   ├── mcp/
│   │   └── server.go            # MCP server (stub)
│   └── cli/
│       ├── root.go              # Root command
│       ├── structure.go         # Structure command
│       └── section.go           # Section command
├── pkg/
│   └── types/
│       └── document.go          # Type definitions
├── docs/                        # Documentation
├── bin/                         # Built binaries
├── go.mod                       # Go module file
├── go.sum                       # Go dependencies
└── Makefile                     # Build automation
```

## Implementation Status

### ✅ Completed Features

- [x] Basic project structure
- [x] Markdown parsing with goldmark
- [x] Document structure extraction
- [x] Section content retrieval
- [x] CLI interface with cobra
- [x] JSON output formatting
- [x] Cross-platform builds
- [x] Build automation with Make

### 🚧 In Progress

- [ ] MCP server implementation
- [ ] STDIO communication protocol
- [ ] File access control and security
- [ ] Caching system
- [ ] Performance optimizations

### 📋 Planned Features

- [ ] File watching and auto-refresh
- [ ] Content search functionality
- [ ] Multi-document support
- [ ] Configuration file support
- [ ] Plugin system for custom parsers

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Format code: `make fmt`
6. Submit a pull request

## License

[License information to be added]

## Technical Details

### Dependencies

- **goldmark**: Markdown parsing
- **cobra**: CLI framework
- **Standard library**: Core functionality

### Architecture

- **Parser**: Uses goldmark AST for reliable Markdown parsing
- **Structure**: Hierarchical section representation with unique IDs
- **CLI**: Command-line interface with multiple output formats
- **MCP**: Future STDIO-based server for AI integration

### Security

- File access will be restricted to specified base directories
- Path traversal protection
- Input validation and sanitization