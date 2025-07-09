# Claude Development Notes

このファイルは、mdatlasプロジェクトの開発過程で得られた知見と、今後の開発において重要な事項をまとめたものです。

## プロジェクト概要

mdatlasは、Markdown文書の構造解析を行うModel Context Protocol (MCP) サーバーです。AI モデルが大きなMarkdown文書を効率的に処理できるよう、セクション単位での選択的アクセスを提供します。

## 開発における重要な知見

### 1. 段階的実装の重要性

- **Phase 1**: 基本的なCLI機能とMarkdown解析
- **Phase 2**: MCP サーバー機能の実装
- **Phase 3**: 最適化とテスト充実

このフェーズ分けにより、各段階でのコミットが適切なサイズに保たれ、機能の完全性が確保されました。

### 2. テスト駆動開発 (TDD) の実践

```bash
# 基本的なテストコマンド
make test           # 全テスト実行
make test-coverage  # カバレッジレポート生成
go test ./internal/core  # 特定パッケージのテスト
```

- **Unit Tests**: `internal/core/parser_test.go`
- **Integration Tests**: `tests/integration/`
- **Fixtures**: `tests/fixtures/` にテスト用文書を配置

### 3. 依存関係管理

主要な依存関係：
- `github.com/yuin/goldmark`: Markdown解析
- `github.com/spf13/cobra`: CLI フレームワーク
- `github.com/stretchr/testify`: テストフレームワーク

```bash
# 依存関係の管理
go mod download
go mod tidy
```

### 3.1 テストアーキテクチャ設計

包括的なテストスイートから得られた知見：

#### テスト分類と組織化
```
tests/
├── integration/
│   ├── cli_comprehensive_test.go    # CLI全機能テスト
│   ├── mcp_server_test.go          # MCP サーバーテスト
│   ├── edge_cases_test.go          # エッジケース・異常系テスト
│   └── performance_test.go         # パフォーマンス・負荷テスト
├── fixtures/
│   ├── sample.md                   # 標準テストファイル
│   ├── complex.md                  # 複雑な構造テスト
│   └── edge_cases.md               # エッジケース専用
└── helpers/
    └── setup_test.go               # テストヘルパー関数
```

#### 統合テストのベストプラクティス
```go
// setupTest 関数 - 全テストで使用される共通初期化
func setupTest(t *testing.T) (projectRoot, binaryPath string) {
    // バイナリビルドとパス設定
    // プロジェクトルート特定
    // テスト固有の前処理
}

// テーブル駆動テストの活用
tests := []struct {
    name     string
    args     []string
    validate func(t *testing.T, output []byte, err error)
}{
    // 複数のテストケース定義
}
```

#### MCP プロトコルテスト手法
```go
// JSON-RPC 2.0 準拠テスト
type MCPRequest struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      interface{}     `json:"id"`
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params,omitempty"`
}

// 非同期通信テスト
func sendMCPRequest(t *testing.T, projectRoot, binaryPath string, request MCPRequest) MCPResponse {
    // STDIO パイプでの通信
    // タイムアウト処理
    // レスポンス検証
}
```

#### パフォーマンステスト戦略
```go
// 大規模文書生成
func generateLargeDocument(numSections, maxDepth int) string {
    // 指定された構造の文書生成
    // パフォーマンステスト用データ作成
}

// 負荷テスト
func TestStressTestMultipleFiles(t *testing.T) {
    // 複数ファイル同時処理
    // メモリ使用量監視
    // 処理時間測定
}
```

### 4. ビルドシステム

```makefile
# 重要なMakefileターゲット
build:      # 基本ビルド
test:       # テスト実行
release:    # クロスプラットフォームビルド
clean:      # 成果物削除
```

### 5. Git コミット戦略

適切なコミットサイズの維持：
- 1つの機能あたり1コミット
- 関連する複数ファイルは同時にコミット
- コミットメッセージは機能の説明を含む
- 自動生成の署名を含める

```bash
# コミットメッセージの例
git commit -m "$(cat <<'EOF'
Add core infrastructure: structure management, caching, and security

- Add structure.go: Document structure management
- Add cache.go: LRU cache with TTL
- Add security.go: File access control

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

### 6. エラーハンドリング設計

- ファイルアクセス制御による外部ファイルへの不正アクセス防止
- パストラバーサル攻撃の防止
- 適切なエラーメッセージの提供
- JSON-RPC エラーコードの標準準拠

### 7. パフォーマンス最適化

- **キャッシュシステム**: LRU キャッシュによる高速化
- **階層構造構築**: 効率的なポインタ操作
- **メモリ使用量制限**: 最大ファイルサイズ制限 (50MB)
- **並行処理**: 将来的な並行アクセスを考慮した設計

### 8. セキュリティ考慮事項

```go
// 重要なセキュリティ機能
- BaseDir制限による外部ファイルアクセス防止
- 許可された拡張子のみ処理 (.md, .markdown, .txt)
- ファイルサイズ制限
- 入力検証とサニタイゼーション
```

### 9. CI/CD パイプライン

```yaml
# .github/workflows/test.yml の重要な設定
- 複数のGo バージョンでのテスト (1.22.x, 1.23.x)
- クロスプラットフォームビルド
- カバレッジレポート
- 自動リリース
```

### 10. MCP プロトコル実装

重要な MCP ツール：
- `get_markdown_structure`: 文書構造の取得
- `get_markdown_section`: セクション内容の取得
- `search_markdown_content`: コンテンツ検索
- `get_markdown_stats`: 統計情報
- `get_markdown_toc`: 目次生成

### 11. 今後の開発で注意すべき点

#### A. テストの充実
- エッジケースの追加
- パフォーマンステストの実装
- セキュリティテストの強化

#### B. 機能追加時の考慮事項
- 既存のAPI 互換性の維持
- セキュリティ影響の評価
- パフォーマンス影響の測定

#### C. ドキュメント保守
- README.md の更新
- API ドキュメントの同期
- 使用例の更新

### 11.1 テスト実装における重要な知見

#### エッジケース処理の重要性
```go
// Unicode文字、特殊文字、空タイトル、深いネストレベルの処理
func TestMarkdownEdgeCases(t *testing.T) {
    // 日本語・中国語・アラビア語のテスト
    // 空のセクションタイトル
    // 6レベルを超えるネスト
    // HTMLタグ混在の処理
}
```

#### パフォーマンスベンチマーク
```go
// 処理時間の上限設定
tests := []struct {
    name        string
    maxDuration time.Duration
}{
    {"structure analysis", 5 * time.Second},
    {"section extraction", 2 * time.Second},
    {"version command", 100 * time.Millisecond},
}
```

#### MCP プロトコルエラーハンドリング
```go
// JSON-RPC 2.0エラーコードの適切な処理
errorCodes := map[string]int{
    "invalid_request": -32600,
    "method_not_found": -32601, 
    "invalid_params": -32602,
    "internal_error": -32603,
}
```

#### テストデータ管理
```go
// 動的テストデータ生成
func generateLargeDocument(numSections, maxDepth int) string {
    // 一貫性のあるテストデータ生成
    // メモリ効率的な文書作成
    // 様々なMarkdown要素の含有
}
```

### 12. 開発環境設定

```bash
# 必須コマンド
make deps    # 依存関係ダウンロード
make build   # ビルド
make test    # テスト実行
make fmt     # コードフォーマット
```

### 13. デバッグ手法

```bash
# 基本的なデバッグコマンド
./bin/mdatlas structure tests/fixtures/sample.md --pretty
./bin/mdatlas section tests/fixtures/sample.md --section-id <id>
./bin/mdatlas --mcp-server --base-dir tests/fixtures
```

### 13.1 テストベースデバッグ手法

#### テスト実行とデバッグ
```bash
# 特定のテストケース実行
go test -v ./tests/integration -run TestCLIStructureCommandComprehensive
go test -v ./tests/integration -run TestMCPServerToolsCall
go test -v ./tests/integration -run TestPerformanceStructureAnalysis

# パフォーマンステストの実行
go test -v ./tests/integration -run TestStressTestLargeFile

# 短時間テストのスキップ
go test -v -short ./tests/integration

# テストタイムアウト調整
go test -v -timeout 30s ./tests/integration
```

#### 統合テストのデバッグ
```go
// テストでのログ出力
t.Logf("Performance: %s completed in %v", tt.name, duration)
t.Logf("Large file size: %d bytes (%.2f MB)", stat.Size(), float64(stat.Size())/1024/1024)

// エラー詳細出力
if err != nil {
    t.Errorf("Expected success but got error: %v. Output: %s", err, string(output))
}
```

#### MCP サーバーの手動テスト
```bash
# 手動でMCPサーバーを起動
./bin/mdatlas --mcp-server --base-dir tests/fixtures

# 別ターミナルでJSONリクエスト送信
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}' | ./bin/mdatlas --mcp-server --base-dir tests/fixtures
```

## 今後の拡張予定

1. **WebSocket サポート**: リアルタイム通知
2. **プラグインシステム**: カスタムパーサー
3. **マルチファイル対応**: 複数ファイルの横断検索
4. **パフォーマンス最適化**: 並列処理とインデックス化

## 重要なファイル構造

```
mdatlas/
├── cmd/mdatlas/main.go              # エントリーポイント
├── internal/
│   ├── core/                        # コア機能
│   │   ├── parser.go               # Markdown解析
│   │   ├── structure.go            # 構造管理
│   │   ├── cache.go                # キャッシュ
│   │   └── security.go             # セキュリティ
│   ├── mcp/                         # MCP サーバー
│   │   ├── server.go               # サーバー本体
│   │   ├── tools.go                # ツール実装
│   │   └── protocol.go             # プロトコル
│   └── cli/                         # CLI インターフェース
├── pkg/types/document.go            # 型定義
├── tests/                           # テスト
└── examples/                        # 使用例
```

このガイドラインに従うことで、一貫性のある高品質な開発を継続できます。

## 包括的テストスイートから得られた開発知見

### 1. テスト実装のベストプラクティス

#### 統合テストの構造化
- **CLI テスト**: 全機能をカバーする網羅的なテストケース
- **MCP サーバーテスト**: JSON-RPC 2.0プロトコル準拠の検証
- **エッジケーステスト**: Unicode、特殊文字、異常なデータ構造の処理
- **パフォーマンステスト**: 処理時間とメモリ使用量の監視

#### 重要な発見事項

##### CLI テストの実装
```go
// テーブル駆動テスト + 検証関数の組み合わせ
tests := []struct {
    name     string
    args     []string
    validate func(t *testing.T, output []byte, err error)
}{
    // 各テストケースに専用の検証ロジック
}
```

##### MCP サーバーテストの技術的課題
- **非同期通信**: STDIO通信での適切なタイムアウト処理
- **JSON-RPC**: エラーコード(-32600, -32601, -32602, -32603)の正確な実装
- **リソース管理**: プロセスの適切な終了処理

##### エッジケース処理の重要性
- **Unicode文字**: 日本語、中国語、アラビア語の適切な処理
- **空のセクション**: タイトルが空の場合の処理
- **深いネスト**: 6レベル以上のヘッダー階層の処理
- **特殊文字**: HTML、マークダウン記法のエスケープ処理

##### パフォーマンス最適化指針
- **処理時間上限**: 各コマンドの適切な実行時間設定
- **メモリ使用量**: 大規模文書処理時の効率的なメモリ管理
- **負荷テスト**: 複数ファイル同時処理での安定性確保

### 2. 開発プロセスの改善点

#### テストファースト開発
```go
// 機能実装前にテストケースを定義
func TestNewFeature(t *testing.T) {
    // 期待される動作を先に定義
    // 実装後に検証可能な状態にする
}
```

#### 継続的品質保証
- **自動化されたテストスイート**: 全機能の回帰テスト
- **パフォーマンスベンチマーク**: 処理速度の継続的監視
- **エラーハンドリング**: 異常系の適切な処理とメッセージ

### 3. プロジェクト成熟度の向上

#### 実装完了の確認事項
- ✅ **CLI 全機能テスト**: 200+ テストケース
- ✅ **MCP サーバーテスト**: 5つのツール + プロトコル検証
- ✅ **エッジケース処理**: Unicode、特殊文字、異常データ
- ✅ **パフォーマンステスト**: 大規模文書、負荷テスト、メモリ使用量

#### 今後の保守における重要事項
1. **テストの継続的更新**: 新機能追加時の必須テストケース追加
2. **パフォーマンス監視**: 処理速度の継続的ベンチマーク
3. **エラーハンドリング**: 異常系テストの拡充
4. **セキュリティテスト**: ファイルアクセス制御の検証

このドキュメント(`CLAUDE.md`)は継続的に更新され、改善しなければいけません。あなたが新しいコミットを作成した後、必ずこのドキュメントを確認し、必要な変更を加えてください。