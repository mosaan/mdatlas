# User Guide

Welcome to mdatlas! This guide will help you get started with using mdatlas for efficient Markdown document analysis.

## What is mdatlas?

mdatlas is a Model Context Protocol (MCP) server that provides AI models with efficient access to Markdown document structures. Instead of loading entire files into memory, mdatlas allows selective access to specific sections, making it ideal for working with large documents within AI context windows.

## Installation

### Prerequisites

- Go 1.22 or later
- Make (for building from source)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/mosaan/mdatlas.git
cd mdatlas

# Build the project
make build

# Install to system PATH
make install
```

### Binary Releases

Download pre-built binaries from the [GitHub releases page](https://github.com/mosaan/mdatlas/releases).

## Quick Start

### Command Line Usage

#### 1. Analyze Document Structure

```bash
# Basic structure analysis
mdatlas structure document.md

# Pretty-printed output
mdatlas structure document.md --pretty

# Limit heading depth
mdatlas structure document.md --max-depth 3
```

#### 2. Extract Section Content

```bash
# First, get the structure to find section IDs
mdatlas structure document.md --pretty

# Then extract a specific section
mdatlas section document.md --section-id section_abc123

# Include child sections
mdatlas section document.md --section-id section_abc123 --include-children

# Output as JSON
mdatlas section document.md --section-id section_abc123 --format json
```

#### 3. Get Help

```bash
# General help
mdatlas --help

# Command-specific help
mdatlas structure --help
mdatlas section --help
```

### MCP Server Mode

To use mdatlas as an MCP server for AI model integration:

```bash
# Start MCP server
mdatlas --mcp-server --base-dir /path/to/your/documents

# The server will communicate via STDIO using JSON-RPC protocol
```

## Working with Documents

### Document Structure

mdatlas analyzes Markdown documents and creates a hierarchical structure based on headings:

```markdown
# Level 1 Heading
## Level 2 Heading
### Level 3 Heading
```

Each section gets:
- **Unique ID**: For referencing the section
- **Level**: Heading level (1-6)
- **Title**: The heading text
- **Position**: Start and end line numbers
- **Metrics**: Character and line counts
- **Children**: Nested subsections

### Section IDs

Section IDs are automatically generated using a hash-based approach that ensures uniqueness even for sections with identical titles at different levels.

### File Support

mdatlas supports:
- `.md` files (Markdown)
- `.markdown` files
- `.txt` files (limited support)

### Size Limits

- Maximum file size: 50MB
- Maximum sections per document: 1000
- Files must be valid UTF-8 encoded

## Advanced Features

### Caching

mdatlas includes intelligent caching:
- **File-based caching**: Structures are cached with file modification tracking
- **Automatic invalidation**: Cache updates when files are modified
- **Configurable size**: Default cache holds 100 documents

### Security

mdatlas implements robust security features:
- **Directory restriction**: Access limited to specified base directory
- **Path traversal protection**: Prevents `../` attacks
- **File type filtering**: Only allows safe file extensions
- **Size validation**: Prevents processing of overly large files

## Configuration

### Environment Variables

- `MDATLAS_BASE_DIR`: Default base directory for file access
- `MDATLAS_CACHE_SIZE`: Maximum number of cached documents (default: 100)
- `MDATLAS_CACHE_TTL`: Cache time-to-live in minutes (default: 30)

### Command Line Flags

- `--base-dir`: Set base directory for file access
- `--mcp-server`: Run as MCP server
- `--pretty`: Pretty-print JSON output
- `--max-depth`: Limit heading depth for structure analysis

## Best Practices

### Document Organization

1. **Use clear heading hierarchy**: Maintain consistent heading levels
2. **Descriptive titles**: Use meaningful section titles
3. **Reasonable file sizes**: Keep files under 10MB for optimal performance
4. **UTF-8 encoding**: Ensure all files are UTF-8 encoded

### Performance Optimization

1. **Use caching**: Let mdatlas cache frequently accessed documents
2. **Limit depth**: Use `--max-depth` for large documents when appropriate
3. **Selective access**: Use section IDs to access only needed content
4. **Monitor file sizes**: Large files may impact performance

### Security Considerations

1. **Restrict base directory**: Use the most restrictive base directory possible
2. **Review file permissions**: Ensure proper file system permissions
3. **Monitor access**: Keep track of which files are being accessed
4. **Regular updates**: Keep mdatlas updated to the latest version

## Troubleshooting

### Common Issues

#### "File not found" errors
- Verify the file path is correct
- Check file permissions
- Ensure the file is within the base directory

#### "Access denied" errors
- Check if the file is within the allowed base directory
- Verify file has correct permissions
- Ensure file extension is allowed (.md, .markdown, .txt)

#### "File too large" errors
- Check if file exceeds 50MB limit
- Consider splitting large documents into smaller files
- Use section-based access instead of full file processing

#### Performance issues
- Enable caching for frequently accessed files
- Consider file size optimization
- Use selective section access instead of full document processing

### Debug Mode

For debugging issues:

```bash
# Run with verbose output
mdatlas --mcp-server --base-dir /path/to/docs 2>debug.log

# Check the debug log for detailed information
tail -f debug.log
```

### Getting Help

- **GitHub Issues**: Report bugs and feature requests
- **Documentation**: Check the README and API docs
- **Community**: Join discussions on GitHub

## Examples

### Example 1: Document Analysis Workflow

```bash
# 1. Analyze document structure
mdatlas structure large_document.md --pretty > structure.json

# 2. Find sections of interest
cat structure.json | jq '.structure[] | select(.title | contains("Installation"))'

# 3. Extract specific section
mdatlas section large_document.md --section-id section_abc123 > section.md

# 4. Process the extracted section
# ... further processing ...
```

### Example 2: MCP Integration

```bash
# Start MCP server for AI model integration
mdatlas --mcp-server --base-dir ./docs

# The AI model can now use MCP tools to:
# - get_markdown_structure: Analyze document structure
# - get_markdown_section: Extract specific sections
# - search_markdown_content: Find relevant sections
```

## Next Steps

- Explore the [API Documentation](api_documentation.md) for detailed technical information
- Check out the [examples](../examples/) for more usage scenarios
- Contribute to the project on [GitHub](https://github.com/mosaan/mdatlas)

## Version Information

Check your mdatlas version:

```bash
mdatlas version
```

Stay updated with the latest features and bug fixes by regularly updating your installation.