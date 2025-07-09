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

このドキュメント(`CLAUDE.md`)は継続的に更新され、改善しなければいけません。あなたが新しいコミットを作成した後、必ずこのドキュメントを確認し、必要な変更を加えてください。