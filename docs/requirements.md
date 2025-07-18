# mdatlas - MCPサーバー要件定義書

## 1. 背景と目的

### 1.1 背景
大規模なMarkdownファイルを扱う際、AIモデルのコンテキストウィンドウの制限により、ファイル全体を一度に処理することが困難になっている。また、特定のセクションのみを参照したい場合でも、ファイル全体を読み込む必要があり、効率が悪い。

### 1.2 目的
**mdatlas**は、Markdownファイルの構造情報（目次、セクション概要、文字数など）を効率的に提供し、AIモデルが必要な部分のみを選択的に取得できるようにする。これにより：
- コンテキストウィンドウの効率的な活用
- 大規模文書の高速処理
- 構造化された文書アクセスの実現

## 2. 利用者とユースケース

### 2.1 利用者
- **AIアプリケーション開発者**: 文書処理機能の実装
- **対話型AIサービス運営者**: 大規模文書の効率的な処理
- **インフラエンジニア**: モデル間連携基盤の構築
- **文書管理ツール制作者**: Markdownベースの文書管理システム

### 2.2 主要ユースケース
1. **文書構造の把握**: 大規模文書の全体像を素早く理解
2. **選択的読み込み**: 必要なセクションのみを効率的に取得
3. **文書分析**: 各セクションの規模や構成を事前に把握
4. **インデックス作成**: 文書検索やナビゲーション機能の実装

## 3. システム構成

```
[AIモデル/クライアント] ←STDIO→ [mdatlas] ←→ [ローカルMarkdownファイル群]
```

- **mdatlas**: 本要件の対象。ローカルMarkdownファイルの解析・提供
- **クライアント**: AIモデルやアプリケーション
- **ファイルシステム**: ローカルのMarkdownファイル保存場所
- **通信方式**: STDIO（標準入出力）によるシンプルな接続

### 3.1 STDIO方式の採用理由
ローカルなMarkdownファイルを扱う特性上、以下の理由からSTDIO方式を採用：
- **シンプルな構成**: ネットワーク設定やポート管理が不要
- **高速通信**: ローカルプロセス間通信による低レイテンシ
- **セキュリティ**: 外部ネットワークアクセスが不要でセキュアな構成
- **デバッグ容易性**: 標準入出力を直接確認可能
- **リソース効率**: HTTPサーバーのオーバーヘッドなし

## 4. 機能要件

### 4.1 コア機能

#### 4.1.1 ファイルアクセス制御機能
- **ベースディレクトリ設定**: 起動時にアクセス可能なルートディレクトリを指定
- **パス正規化**: 相対パス、シンボリックリンクを解決して絶対パスに変換
- **アクセス範囲検証**: 要求されたファイルパスがベースディレクトリ配下にあることを確認
- **パストラバーサル防止**: `../` などによる上位ディレクトリへのアクセスを禁止
- **ファイル存在確認**: アクセス前にファイルの存在と読み取り権限を確認

#### 4.1.2 文書構造解析機能
- **入力**: ベースディレクトリからの相対パスまたは絶対パス
- **アクセス制御**: 4.1.1の制御を通過したファイルのみ処理
- **出力**: 構造化された目次情報（JSON形式）
- **処理内容**:
  - ヘッダ階層（H1-H6）の解析
  - 各セクションの一意識別子生成
  - 文字数・行数の計測
  - ネストした構造の表現

**出力例**:
```json
{
  "file_path": "/path/to/document.md",
  "total_chars": 15000,
  "total_lines": 500,
  "structure": [
    {
      "id": "section_1",
      "level": 1,
      "title": "はじめに",
      "char_count": 800,
      "line_count": 25,
      "children": [
        {
          "id": "section_1_1",
          "level": 2,
          "title": "背景",
          "char_count": 400,
          "line_count": 12,
          "children": []
        }
      ]
    }
  ]
}
```

#### 4.1.3 セクション取得機能
- **入力**: ファイルパス（アクセス制御対象）、セクションID、取得範囲オプション
- **アクセス制御**: 4.1.1の制御を通過したファイルのみ処理
- **出力**: 指定されたセクションの本文
- **オプション**:
  - `include_children`: 子セクションも含めるかどうか
  - `format`: 出力形式（markdown/plain/html）
  - `max_chars`: 最大文字数制限

#### 4.1.4 検索・フィルタリング機能
- **対象ファイル**: ベースディレクトリ配下のファイルのみ
- **キーワード検索**: セクション内のテキスト検索
- **レベルフィルタ**: 特定の見出しレベルのみ取得
- **範囲指定**: 特定のセクション範囲のみ取得

### 4.2 MCPプロトコル実装

#### 4.2.1 提供するリソース
- `markdown://file/{file_path}/structure`: 文書構造情報
- `markdown://file/{file_path}/section/{section_id}`: 特定セクション

#### 4.2.2 提供するツール
- `get_markdown_structure`: 文書構造の取得
- `get_markdown_section`: セクション内容の取得
- `search_markdown_content`: 文書内検索

## 5. 非機能要件

### 5.1 性能要件
- **応答時間**: 構造解析は1秒以内、セクション取得は500ms以内
- **メモリ使用量**: 1ファイルあたり最大100MB
- **同時処理**: 最大10並列リクエスト

### 5.2 信頼性要件
- **エラー処理**: JSON-RPC形式での適切なエラーレスポンス
- **異常終了対応**: ファイル読み取りエラー時の適切なエラー通知
- **リソース枯渇対応**: メモリ不足時の適切な処理停止

### 5.3 セキュリティ要件
- **ファイルアクセス制御**: 指定されたディレクトリ内のみアクセス可能
- **入力検証**: パストラバーサル攻撃の防止
- **ログ出力**: アクセスログの記録

## 6. 実装方針

### 6.1 推奨実装方式: **ハイブリッド方式**

#### 6.1.1 基本戦略
1. **初回アクセス時**: オンデマンドで解析・キャッシュ保存
2. **2回目以降**: キャッシュから高速応答
3. **ファイル更新検知**: タイムスタンプ比較による自動更新

#### 6.1.2 キャッシュ戦略
- **保存形式**: JSON形式でメモリ＋ファイル永続化
- **無効化条件**: ファイル更新時刻の変更
- **メモリ管理**: LRU方式で最大100ファイル保持

#### 6.1.3 フォールバック
- キャッシュ破損時は自動でオンデマンド解析に切り替え
- 解析エラー時は部分的な構造情報を提供

### 6.2 技術選定案

#### 6.2.1 プログラミング言語
- **Python**: 豊富なMarkdownライブラリ（推奨）
- **TypeScript**: Node.js環境での高いパフォーマンス
- **Go**: 高速処理とシンプルなデプロイ

#### 6.2.2 主要ライブラリ
- **Markdown解析**: `python-markdown`, `markdown-it`
- **MCPプロトコル**: 各言語の公式SDK
- **ファイル監視**: `watchdog`, `chokidar`

### 6.3 プロジェクト構造案
```
mdatlas/
├── src/
│   ├── core/           # コア機能
│   ├── mcp/            # MCPプロトコル実装
│   └── cli/            # CLI インターフェース
├── tests/
├── docs/
└── examples/
```

## 7. API仕様

### 7.1 MCPツール仕様

#### 7.1.1 get_markdown_structure
```json
{
  "name": "get_markdown_structure",
  "description": "Markdownファイルの構造情報を取得",
  "inputSchema": {
    "type": "object",
    "properties": {
      "file_path": {
        "type": "string",
        "description": "ベースディレクトリからの相対パスまたは絶対パス"
      },
      "max_depth": {"type": "integer", "default": 6}
    },
    "required": ["file_path"]
  }
}
```
**アクセス制御**: ベースディレクトリ配下のファイルのみアクセス可能

#### 7.1.2 get_markdown_section
```json
{
  "name": "get_markdown_section",
  "description": "特定セクションの内容を取得",
  "inputSchema": {
    "type": "object",
    "properties": {
      "file_path": {
        "type": "string",
        "description": "ベースディレクトリからの相対パスまたは絶対パス"
      },
      "section_id": {"type": "string"},
      "include_children": {"type": "boolean", "default": false},
      "format": {"type": "string", "enum": ["markdown", "plain"], "default": "markdown"}
    },
    "required": ["file_path", "section_id"]
  }
}
```
**アクセス制御**: ベースディレクトリ配下のファイルのみアクセス可能

## 8. 開発・運用

### 8.1 開発環境
- **バージョン管理**: Git
- **コンテナ化**: Docker + Docker Compose（テスト用）
- **実行方式**: STDIO接続によるスタンドアロン実行
- **依存関係管理**: Poetry (Python) / npm (Node.js)
- **設定管理**: 起動時引数でベースディレクトリを指定

### 8.2 STDIO実装の考慮事項
- **標準入出力の管理**: JSON-RPC形式でのメッセージ送受信
- **エラーハンドリング**: 標準エラー出力への適切なログ出力
- **プロセス管理**: 親プロセスからの終了シグナル処理
- **バッファリング**: 大きなレスポンスに対する適切なバッファリング
- **初期化処理**: 起動時のベースディレクトリ設定と検証

### 8.3 テスト戦略
- **ユニットテスト**: 各機能の単体テスト（カバレッジ90%以上）
- **統合テスト**: MCP プロトコル（STDIO）の動作確認
- **E2Eテスト**: 実際のAIモデルとのSTDIO連携テスト
- **パフォーマンステスト**: 大規模ファイルでの性能測定
- **STDIO テスト**: 標準入出力での適切なメッセージ送受信の確認

### 8.4 CI/CD
- **GitHub Actions** による自動テスト・ビルド
- **バイナリ配布**: 各OS向けのmdatlas実行可能ファイル生成
- **セマンティックバージョニング** の採用
- **STDIO接続テスト**: CI環境での自動STDIO通信テスト

### 8.5 監視・ログ
- **ログレベル**: DEBUG, INFO, WARNING, ERROR
- **ログ出力**: 標準エラー出力への構造化ログ
- **メトリクス**: 応答時間、エラー率、キャッシュヒット率
- **プロセス監視**: 親プロセスとの接続状態監視

## 9. 制約事項

### 9.1 対応ファイル形式
- **対応**: 標準Markdown (.md, .markdown)
- **非対応**: 独自拡張、バイナリファイル

### 9.2 ファイルサイズ制限
- **最大ファイルサイズ**: 50MB
- **最大セクション数**: 1000個

### 9.3 文字エンコーディング
- **対応**: UTF-8
- **非対応**: Shift_JIS, EUC-JP等

## 10. 今後の拡張予定

### 10.1 フェーズ2機能
- **マルチファイル対応**: 複数ファイルの横断検索
- **リアルタイム更新**: WebSocketによるリアルタイム通知
- **プラグイン機能**: カスタムパーサーの追加

### 10.2 パフォーマンス改善
- **並列処理**: 大規模ファイルの並列解析
- **インデックス化**: 高速検索のためのインデックス構築
- **圧縮**: キャッシュデータの圧縮保存