# mdatlas Go実装計画

## 技術スタック

### 言語・バージョン
- **Go**: 1.21+ (generics活用、新しいログ機能など)
- **最小対応**: Go 1.19+ (幅広い環境対応)

### 主要依存関係
```go
// go.mod
module github.com/mosaan/mdatlas

go 1.21

require (
    github.com/yuin/goldmark v1.6.0           // Markdown parser
    github.com/spf13/cobra v1.8.0             // CLI framework
    github.com/spf13/viper v1.17.0            // Configuration
    github.com/fsnotify/fsnotify v1.7.0       // File watching
    github.com/stretchr/testify v1.8.4        // Testing
    // MCP SDK - 具体的なパッケージは後で調査
)
```

## プロジェクト構造

```
mdatlas/
├── cmd/
│   └── mdatlas/
│       └── main.go              # エントリーポイント
├── internal/
│   ├── core/
│   │   ├── parser.go            # Markdown解析
│   │   ├── structure.go         # 構造情報管理
│   │   ├── cache.go             # キャッシュ管理
│   │   └── security.go          # ファイルアクセス制御
│   ├── mcp/
│   │   ├── server.go            # MCPサーバー
│   │   ├── tools.go             # MCPツール実装
│   │   └── protocol.go          # プロトコル処理
│   └── cli/
│       ├── root.go              # ルートコマンド
│       ├── structure.go         # structure サブコマンド
│       └── section.go           # section サブコマンド
├── pkg/
│   └── types/
│       └── document.go          # 公開型定義
├── docs/
│   └── requirements.md
├── examples/
│   └── sample_documents/
├── tests/
│   ├── fixtures/
│   └── integration/
├── scripts/
│   ├── build.sh
│   └── release.sh
├── .github/
│   └── workflows/
│       ├── test.yml
│       └── release.yml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## 型定義

### 基本型
```go
// pkg/types/document.go
package types

import "time"

// DocumentStructure は文書の構造情報を表現
type DocumentStructure struct {
    FilePath    string    `json:"file_path"`
    TotalChars  int       `json:"total_chars"`
    TotalLines  int       `json:"total_lines"`
    Structure   []Section `json:"structure"`
    LastModified time.Time `json:"last_modified"`
}

// Section は文書のセクション情報
type Section struct {
    ID         string    `json:"id"`
    Level      int       `json:"level"`
    Title      string    `json:"title"`
    CharCount  int       `json:"char_count"`
    LineCount  int       `json:"line_count"`
    StartLine  int       `json:"start_line"`
    EndLine    int       `json:"end_line"`
    Children   []Section `json:"children"`
}

// SectionContent はセクションの内容
type SectionContent struct {
    ID             string `json:"id"`
    Title          string `json:"title"`
    Content        string `json:"content"`
    Format         string `json:"format"`
    IncludeChildren bool   `json:"include_children"`
}

// AccessConfig はファイルアクセス制御設定
type AccessConfig struct {
    BaseDir     string   `json:"base_dir"`
    AllowedExts []string `json:"allowed_extensions"`
    MaxFileSize int64    `json:"max_file_size"`
}
```

## 主要機能実装

### 1. Markdown解析エンジン
```go
// internal/core/parser.go
package core

import (
    "bytes"
    "github.com/yuin/goldmark"
    "github.com/yuin/goldmark/ast"
    "github.com/yuin/goldmark/text"
    "github.com/mosaan/mdatlas/pkg/types"
)

type Parser struct {
    md goldmark.Markdown
}

func NewParser() *Parser {
    return &Parser{
        md: goldmark.New(
            goldmark.WithExtensions(
                // 必要な拡張を追加
            ),
        ),
    }
}

func (p *Parser) ParseStructure(content []byte) (*types.DocumentStructure, error) {
    doc := p.md.Parser().Parse(text.NewReader(content))
    
    structure := &types.DocumentStructure{
        TotalChars: len(content),
        TotalLines: bytes.Count(content, []byte("\n")) + 1,
        Structure:  []types.Section{},
    }
    
    // AST を走査してセクション情報を抽出
    err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
        if entering && node.Kind() == ast.KindHeading {
            section := p.extractSection(node, content)
            structure.Structure = append(structure.Structure, section)
        }
        return ast.WalkContinue, nil
    })
    
    return structure, err
}

func (p *Parser) extractSection(node ast.Node, content []byte) types.Section {
    heading := node.(*ast.Heading)
    
    return types.Section{
        ID:        generateSectionID(heading),
        Level:     heading.Level,
        Title:     extractHeadingText(heading, content),
        StartLine: getLineNumber(node, content),
        // その他のフィールドを設定
    }
}
```

### 2. MCPサーバー実装
```go
// internal/mcp/server.go
package mcp

import (
    "context"
    "encoding/json"
    "os"
    "github.com/mosaan/mdatlas/internal/core"
    "github.com/mosaan/mdatlas/pkg/types"
)

type Server struct {
    parser     *core.Parser
    cache      *core.Cache
    accessCtrl *core.AccessControl
}

func NewServer(baseDir string) (*Server, error) {
    accessCtrl, err := core.NewAccessControl(baseDir)
    if err != nil {
        return nil, err
    }
    
    return &Server{
        parser:     core.NewParser(),
        cache:      core.NewCache(),
        accessCtrl: accessCtrl,
    }, nil
}

func (s *Server) Run(ctx context.Context) error {
    // STDIO でのMCP通信を処理
    decoder := json.NewDecoder(os.Stdin)
    encoder := json.NewEncoder(os.Stdout)
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            var request MCPRequest
            if err := decoder.Decode(&request); err != nil {
                continue
            }
            
            response := s.handleRequest(request)
            encoder.Encode(response)
        }
    }
}

func (s *Server) handleRequest(req MCPRequest) MCPResponse {
    switch req.Method {
    case "tools/call":
        return s.handleToolCall(req)
    case "resources/list":
        return s.handleResourcesList(req)
    case "resources/read":
        return s.handleResourcesRead(req)
    default:
        return MCPResponse{
            ID:    req.ID,
            Error: &MCPError{Code: -32601, Message: "Method not found"},
        }
    }
}
```

### 3. ツール実装
```go
// internal/mcp/tools.go
package mcp

import (
    "path/filepath"
    "github.com/mosaan/mdatlas/pkg/types"
)

func (s *Server) handleToolCall(req MCPRequest) MCPResponse {
    var params ToolCallParams
    if err := json.Unmarshal(req.Params, &params); err != nil {
        return errorResponse(req.ID, -32602, "Invalid params")
    }
    
    switch params.Name {
    case "get_markdown_structure":
        return s.getMarkdownStructure(req.ID, params.Arguments)
    case "get_markdown_section":
        return s.getMarkdownSection(req.ID, params.Arguments)
    default:
        return errorResponse(req.ID, -32601, "Unknown tool")
    }
}

func (s *Server) getMarkdownStructure(id string, args map[string]interface{}) MCPResponse {
    filePath, ok := args["file_path"].(string)
    if !ok {
        return errorResponse(id, -32602, "Missing file_path")
    }
    
    // ファイルアクセス制御
    if !s.accessCtrl.IsAllowed(filePath) {
        return errorResponse(id, -32603, "Access denied")
    }
    
    // キャッシュ確認
    if structure, exists := s.cache.GetStructure(filePath); exists {
        return MCPResponse{
            ID:     id,
            Result: structure,
        }
    }
    
    // ファイル読み取り・解析
    content, err := os.ReadFile(filePath)
    if err != nil {
        return errorResponse(id, -32603, "Failed to read file")
    }
    
    structure, err := s.parser.ParseStructure(content)
    if err != nil {
        return errorResponse(id, -32603, "Failed to parse structure")
    }
    
    // キャッシュに保存
    s.cache.SetStructure(filePath, structure)
    
    return MCPResponse{
        ID:     id,
        Result: structure,
    }
}
```

## CLI実装

### コマンド構造
```go
// internal/cli/root.go
package cli

import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
    Use:   "mdatlas",
    Short: "A Model Context Protocol server for Markdown document structure analysis",
    Long: `mdatlas provides efficient access to Markdown document structure information,
allowing AI models to selectively retrieve specific sections without loading entire files.`,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    cobra.OnInitialize(initConfig)
    
    // MCP サーバーモード (デフォルト)
    rootCmd.PersistentFlags().String("base-dir", ".", "Base directory for file access")
    rootCmd.PersistentFlags().Bool("mcp-server", false, "Run as MCP server (STDIO mode)")
    
    // CLI モード用サブコマンド
    rootCmd.AddCommand(structureCmd)
    rootCmd.AddCommand(sectionCmd)
    rootCmd.AddCommand(versionCmd)
}

// MCP サーバーモード実行
func runMCPServer(baseDir string) error {
    server, err := mcp.NewServer(baseDir)
    if err != nil {
        return err
    }
    
    return server.Run(context.Background())
}
```

### サブコマンド実装
```go
// internal/cli/structure.go
var structureCmd = &cobra.Command{
    Use:   "structure <file>",
    Short: "Extract structure information from Markdown file",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        parser := core.NewParser()
        content, err := os.ReadFile(args[0])
        if err != nil {
            return err
        }
        
        structure, err := parser.ParseStructure(content)
        if err != nil {
            return err
        }
        
        return json.NewEncoder(os.Stdout).Encode(structure)
    },
}
```

## ビルド・リリース

### Makefile
```makefile
.PHONY: build test clean install release

BINARY_NAME=mdatlas
VERSION=$(shell git describe --tags --always --dirty)
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

build:
	go build -ldflags="-X main.version=${VERSION} -X main.buildDate=${BUILD_DATE}" \
		-o bin/${BINARY_NAME} cmd/mdatlas/main.go

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean:
	rm -rf bin/
	rm -f coverage.out

install:
	go install cmd/mdatlas/main.go

# クロスコンパイル
release:
	GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=${VERSION}" -o bin/mdatlas-linux-amd64 cmd/mdatlas/main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.version=${VERSION}" -o bin/mdatlas-darwin-amd64 cmd/mdatlas/main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.version=${VERSION}" -o bin/mdatlas-darwin-arm64 cmd/mdatlas/main.go
	GOOS=windows GOARCH=amd64 go build -ldflags="-X main.version=${VERSION}" -o bin/mdatlas-windows-amd64.exe cmd/mdatlas/main.go
```

## 開発フェーズ

### Phase 1: 基本実装
1. プロジェクト構造セットアップ
2. Markdown解析エンジン
3. 基本的な構造抽出
4. CLI版の実装

### Phase 2: MCP統合
1. MCPプロトコル実装
2. STDIO通信処理
3. ツール定義・実装

### Phase 3: 最適化
1. キャッシュ機能
2. ファイル監視
3. パフォーマンス最適化
