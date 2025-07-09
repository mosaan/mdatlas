# API Documentation

This document provides comprehensive API documentation for the mdatlas service.

## Overview

mdatlas provides a Model Context Protocol (MCP) server that allows AI models to efficiently access and analyze Markdown document structures without loading entire files into memory.

## Authentication

Currently, mdatlas operates in a local file system mode with directory-based access control. No authentication is required for local usage.

## Base URL

When running as an MCP server, mdatlas communicates via STDIO (standard input/output) using JSON-RPC protocol.

## Available Tools

### get_markdown_structure

Extracts the hierarchical structure of a Markdown document.

**Parameters:**
- `file_path` (string, required): Path to the Markdown file
- `max_depth` (integer, optional): Maximum heading depth to include (1-6)

**Response:**
```json
{
  "file_path": "/path/to/document.md",
  "total_chars": 15000,
  "total_lines": 500,
  "structure": [
    {
      "id": "section_abc123",
      "level": 1,
      "title": "Introduction",
      "char_count": 800,
      "line_count": 25,
      "start_line": 1,
      "end_line": 25,
      "children": []
    }
  ],
  "last_modified": "2024-01-15T10:30:00Z"
}
```

### get_markdown_section

Retrieves content from a specific section of a Markdown document.

**Parameters:**
- `file_path` (string, required): Path to the Markdown file
- `section_id` (string, required): Unique identifier of the section
- `include_children` (boolean, optional): Include child sections (default: false)
- `format` (string, optional): Output format - "markdown" or "plain" (default: "markdown")

**Response:**
```json
{
  "id": "section_abc123",
  "title": "Introduction",
  "content": "# Introduction\n\nThis is the introduction section...",
  "format": "markdown",
  "include_children": false
}
```

### search_markdown_content

Searches for sections containing specific text in their titles.

**Parameters:**
- `file_path` (string, required): Path to the Markdown file
- `query` (string, required): Search query
- `case_sensitive` (boolean, optional): Case-sensitive search (default: false)

**Response:**
```json
{
  "file_path": "/path/to/document.md",
  "query": "installation",
  "results": [
    {
      "id": "section_def456",
      "level": 2,
      "title": "Installation Guide",
      "char_count": 1200,
      "line_count": 35,
      "start_line": 50,
      "end_line": 85,
      "children": []
    }
  ],
  "count": 1
}
```

### get_markdown_stats

Provides statistics about a Markdown document.

**Parameters:**
- `file_path` (string, required): Path to the Markdown file

**Response:**
```json
{
  "file_path": "/path/to/document.md",
  "total_chars": 15000,
  "total_lines": 500,
  "section_count": 25,
  "level_counts": {
    "1": 3,
    "2": 8,
    "3": 12,
    "4": 2
  },
  "last_modified": "2024-01-15T10:30:00Z"
}
```

### get_markdown_toc

Generates a table of contents for a Markdown document.

**Parameters:**
- `file_path` (string, required): Path to the Markdown file
- `max_depth` (integer, optional): Maximum heading depth to include

**Response:**
```json
{
  "file_path": "/path/to/document.md",
  "toc": [
    {
      "id": "section_abc123",
      "level": 1,
      "title": "Introduction",
      "line": 1
    },
    {
      "id": "section_def456",
      "level": 2,
      "title": "Getting Started",
      "line": 25
    }
  ],
  "count": 2
}
```

## Available Resources

### Document Structure Resource

- **URI Pattern**: `markdown://file/{file_path}/structure`
- **Description**: Provides access to the hierarchical structure of a Markdown file
- **Content-Type**: `application/json`

### Document Content Resource

- **URI Pattern**: `markdown://file/{file_path}/content`
- **Description**: Provides access to the full content of a Markdown file
- **Content-Type**: `text/markdown`

## Error Handling

All API responses follow the JSON-RPC 2.0 error format:

```json
{
  "jsonrpc": "2.0",
  "id": "request_id",
  "error": {
    "code": -32603,
    "message": "Internal error",
    "data": "Additional error details"
  }
}
```

### Common Error Codes

- `-32700`: Parse error - Invalid JSON
- `-32600`: Invalid request - Invalid request object
- `-32601`: Method not found - Method does not exist
- `-32602`: Invalid params - Invalid method parameters
- `-32603`: Internal error - Internal JSON-RPC error

## Security

### File Access Control

mdatlas implements strict file access control:

- **Base Directory Restriction**: Access is limited to files within the configured base directory
- **Path Traversal Protection**: Prevents `../` attacks and symlink traversal
- **File Extension Filtering**: Only `.md`, `.markdown`, and `.txt` files are allowed
- **File Size Limits**: Maximum file size of 50MB

### Best Practices

1. **Limit Base Directory**: Set the most restrictive base directory possible
2. **Regular Updates**: Keep mdatlas updated to the latest version
3. **Monitor Access**: Review access logs for suspicious activity
4. **File Permissions**: Ensure proper file system permissions

## Examples

### Using mdatlas CLI

```bash
# Get document structure
mdatlas structure document.md --pretty

# Get specific section
mdatlas section document.md --section-id section_abc123

# Show version
mdatlas version
```

### Using as MCP Server

```bash
# Start MCP server
mdatlas --mcp-server --base-dir /path/to/documents

# The server will communicate via STDIO using JSON-RPC
```

## Limits and Quotas

- **Maximum file size**: 50MB
- **Maximum sections per document**: 1000
- **Cache size**: 100 documents (configurable)
- **Cache TTL**: 30 minutes (configurable)

## Support

For issues and questions:
- GitHub Issues: https://github.com/mosaan/mdatlas/issues
- Documentation: https://github.com/mosaan/mdatlas/blob/main/README.md