# gopdf 開発進捗

## プロジェクト概要
- **プロジェクト名**: gopdf
- **目的**: Pure GoでのPDF生成・解析ライブラリ開発
- **開始日**: 2024-10-27

## 完了したタスク

### Phase 0: プロジェクト初期化 ✅
- [x] 要件定義書の確認 (docs/requirements.md)
- [x] アーキテクチャ設計書の作成 (docs/architecture.md)
- [x] プロジェクト構造設計書の作成 (docs/structure.md)
- [x] PDF仕様調査とドキュメント化 (docs/pdf_spec_notes.md)
- [x] go.mod初期化
- [x] 基本ディレクトリ構造作成
  - Root package (gopdf)
  - internal/core/
  - internal/font/
  - internal/writer/
  - internal/reader/
  - internal/content/
  - internal/util/
  - internal/testutil/
  - examples/
  - testdata/
  - scripts/

### Phase 1: 基礎構築（MVP）✅
- [x] internal/core パッケージ実装
  - [x] object.go: Object インターフェースと基本型（Null, Boolean, Integer, Real, String, Name）
  - [x] Dictionary型とArray型
  - [x] Stream型（コンテンツストリーム用）
  - [x] Reference型（間接参照）
  - [x] IndirectObject型
  - [x] 包括的なユニットテスト（14テスト、すべてパス）
- [x] internal/writer パッケージ実装
  - [x] serializer.go: PDFオブジェクトのシリアライズ
  - [x] writer.go: PDF書き込み制御
  - [x] xrefテーブル生成
  - [x] trailer出力
  - [x] ヘッダー・EOF出力
  - [x] 包括的なユニットテスト（13テスト、すべてパス）
- [x] Root package (gopdf) 実装
  - [x] document.go: Document型（ドキュメント管理）
  - [x] page.go: Page型（ページ表現）
  - [x] constants.go: ページサイズ（A4, Letter等）と向き定義
  - [x] graphics.go: Color型、図形スタイル定義
  - [x] 包括的なユニットテスト（5テスト、すべてパス）
- [x] 統合テスト
  - [x] 最小限のPDF生成（空ページ）
  - [x] PDF Readerでの開封確認（3ページ生成成功）
  - [x] examples/01_empty_page サンプルコード作成

**Phase 1 実装成果:**
- 合計32のユニットテストがすべてパス
- 有効なPDF 1.7ファイルの生成に成功
- TDD方式で実装し、高いテストカバレッジを達成

### Phase 2: テキスト描画機能 ✅
- [x] internal/font パッケージ実装
  - [x] standard.go: 標準Type1フォント14種対応
  - [x] GetStandardFont関数
  - [x] フォントメタデータ（Name, Type, Encoding）
  - [x] 包括的なユニットテスト（5テスト、すべてパス）
- [x] Root package (gopdf) テキスト描画機能
  - [x] Page.SetFont: フォント設定
  - [x] Page.DrawText: テキスト描画
  - [x] フォントリソース管理
  - [x] コンテンツストリーム生成（BT/ET, Tf, Td, Tj演算子）
  - [x] 包括的なユニットテスト（5テスト、すべてパス）
- [x] 統合テスト
  - [x] 複数フォントでのテキスト描画確認
  - [x] PDF Readerでの正常表示確認
  - [x] examples/02_hello_world サンプルコード作成

**Phase 2 実装成果:**
- 合計42のユニットテストがすべてパス（Phase 1から+10テスト）
- 14種の標準Type1フォント対応
- 複数フォント・サイズでのテキスト描画成功
- Hello WorldサンプルがmacOS Previewで正常表示

### Phase 3: 図形描画機能 ✅
- [x] internal/content パッケージ拡張
  - [x] 基本図形描画（線、矩形、円）
  - [x] 図形スタイル（線幅、色、塗りつぶし）
  - [x] ベジェ曲線対応
- [x] examples/03_graphics サンプル作成
- [x] 包括的なテストの実装

**Phase 3 実装成果:**
- 線、矩形、円、ベジェ曲線の描画機能実装
- 線幅、色、塗りつぶしスタイルのサポート
- Graphics描画サンプル作成

### Phase 4: 画像埋め込み機能 ✅
- [x] JPEG画像サポート
  - [x] internal/image/jpeg パッケージ実装
  - [x] JPEG画像のデコードと埋め込み
  - [x] 画像リソース管理
- [x] PNG画像サポート
  - [x] internal/image/png パッケージ実装
  - [x] PNG画像のデコードと埋め込み
  - [x] 透過対応（アルファチャンネル）
- [x] examples/04_images サンプル作成（JPEG）
- [x] examples/05_png_images サンプル作成（PNG）

**Phase 4 実装成果:**
- JPEG/PNG画像の埋め込み機能実装
- 画像サイズ・位置の自由な指定
- 透過PNG対応
- 画像描画サンプル作成

### Phase 5: PDF読み込み機能 ✅
- [x] internal/reader パッケージ実装
  - [x] PDF Lexer（トークナイザ）
  - [x] PDF Parser（構文解析）
  - [x] xref テーブル解析
  - [x] オブジェクト参照解決
  - [x] ストリーム展開
- [x] PDFReader API設計
  - [x] Open関数
  - [x] ページ数取得
  - [x] メタデータ取得
- [x] examples/06_read_pdf サンプル作成

**Phase 5 実装成果:**
- PDF読み込みエンジン実装
- xref/trailer解析
- 間接参照の自動解決
- PDF情報の抽出API

### Phase 6: テキスト抽出機能 ✅
- [x] internal/content パッケージ拡張
  - [x] コンテンツストリーム解析
  - [x] テキスト抽出エンジン
  - [x] グラフィックス状態管理
  - [x] テキスト位置・スタイル情報の保持
- [x] 構造化テキスト抽出
  - [x] テキストブロック認識
  - [x] フォント情報の抽出
  - [x] 座標情報の保持
- [x] examples/07_structured_text サンプル作成

**Phase 6 実装成果:**
- コンテンツストリームパーサー実装
- 構造化テキスト抽出（位置・フォント情報付き）
- テキスト抽出サンプル作成

### Phase 7: 画像抽出機能 ✅
- [x] internal/content/image_extractor パッケージ実装
  - [x] XObject画像の検出
  - [x] 画像データの抽出
  - [x] 画像形式の判定（JPEG, PNG）
  - [x] 座標・サイズ情報の取得
- [x] ImageInfo構造体設計
  - [x] 画像データ、形式、サイズ
  - [x] PDF内での配置情報
- [x] examples/08_extract_images サンプル作成

**Phase 7 実装成果:**
- PDF内画像の抽出機能実装
- 画像形式の自動判定
- 配置情報付き画像抽出
- 画像抽出サンプル作成

### Phase 8: TTFフォント・日本語対応 ✅
- [x] internal/font パッケージ拡張
  - [x] TTFフォント解析（sfntパーサー）
  - [x] CMap生成（Unicode → GID）
  - [x] フォントサブセット化
  - [x] フォント埋め込み
- [x] internal/writer/ttf_embed パッケージ実装
  - [x] TrueTypeフォント辞書生成
  - [x] ToUnicode CMap生成
  - [x] フォントストリーム埋め込み
- [x] 日本語テキスト描画
  - [x] UTF-8文字列のサポート
  - [x] DrawTextUTF8 API
  - [x] マルチバイト文字の正しい描画
- [x] examples/09_ttf_fonts サンプル作成

**Phase 8 実装成果:**
- TTFフォント完全サポート
- 日本語を含むUnicode文字の描画
- フォントサブセット化による最適化
- 日本語描画サンプル作成

### Phase 9: PDF翻訳機能（レイアウト保持） ✅
- [x] レイアウト解析エンジン実装
  - [x] ページレイアウト構造体（PageLayout）
  - [x] テキストブロック抽出（位置・フォント情報付き）
  - [x] 画像要素の座標・サイズ取得
  - [x] ExtractPageLayout / ExtractAllLayouts API
- [x] 画像処理機能
  - [x] 既存PDFからの画像抽出
  - [x] 画像サイズ情報の保持
  - [x] 元サイズでの画像配置機能
- [x] 可変フォントサイズ機能
  - [x] 矩形領域内でのテキストフィッティング
  - [x] 自動フォントサイズ調整
  - [x] 複数行テキスト対応（自動改行、行間調整）
  - [x] FitTextOptions構造体
- [x] レイアウト保持翻訳機能
  - [x] ページサイズ・余白の保持
  - [x] 要素配置の維持
  - [x] 翻訳テキストの配置
  - [x] TranslatePDF関数
  - [x] PDFTranslatorOptions設定
- [x] 多言語対応
  - [x] TTFフォントによる日本語埋め込み
  - [x] UTF-8文字列のサポート
  - [x] Translator インターフェース設計
- [x] examples/10_pdf_translation サンプル作成
  - [x] レイアウト抽出デモ
  - [x] 英日翻訳デモ
  - [x] READMEドキュメント

**Phase 9 実装成果:**
- レイアウト保持型PDF翻訳機能の完全実装
- 可変フォントサイズによる自動調整
- 画像の位置保持機能
- 日本語フォント対応（TTF必須）
- 翻訳サンプル・ドキュメント完備

### Phase 10: ルビ（ふりがな）機能 ✅
- [x] ルビデータ構造実装
  - [x] RubyText構造体（親文字、ルビ文字）
  - [x] RubyStyle構造体（配置、サイズ比率、オフセット、コピーモード）
  - [x] RubyAlignment定数（中央、左、右揃え）
  - [x] RubyCopyMode定数（親文字のみ、ルビのみ、両方）
- [x] ルビ描画機能
  - [x] DrawRuby: 基本的なルビ描画
  - [x] DrawRubyWithActualText: ActualText対応でコピー動作制御
  - [x] DrawRubyTexts: 複数のルビテキストを連続描画
  - [x] 配置オプション対応（中央、左、右揃え）
  - [x] サイズ比率の調整機能
- [x] ActualText対応
  - [x] PDFコピー時の動作制御
  - [x] 親文字のみコピー（デフォルト）
  - [x] ルビのみコピー
  - [x] 両方コピー（親文字(ルビ)形式）
- [x] ヘルパー関数
  - [x] NewRubyText: 単一ルビテキスト作成
  - [x] NewRubyTextPairs: 複数ルビテキスト一括作成
  - [x] DefaultRubyStyle: デフォルトスタイル取得
- [x] 包括的なテスト
  - [x] ruby_test.go: データ構造テスト
  - [x] page_ruby_test.go: 描画機能テスト（全テストパス）
- [x] examples/11_ruby_annotation サンプル作成
  - [x] 基本的なルビ描画例
  - [x] 配置オプション例
  - [x] ActualTextコピーモード例
  - [x] 複数ルビテキスト例
  - [x] READMEドキュメント

**Phase 10 実装成果:**
- ルビ（ふりがな）描画機能の完全実装（最小限実装、Mecab不使用）
- ActualText対応によるコピー動作の完全制御
- 横書きテキストのみ対応（縦書きは未対応）
- 4つの実用的なサンプルコード作成
- 包括的なテストカバレッジ（全テストパス）

### Phase 11: OCRテキストレイヤー機能 ✅
- [x] テキストレイヤーデータ構造実装
  - [x] TextLayer構造体（単語リスト、レンダリングモード、不透明度）
  - [x] TextLayerWord構造体（テキスト、位置情報）
  - [x] TextRenderMode定数（通常、輪郭、塗りつぶし、非表示）
  - [x] OCRWord/OCRResult構造体（OCR API統合用）
- [x] テキストレイヤー描画機能
  - [x] AddTextLayer: テキストレイヤー追加
  - [x] AddTextLayerWords: 個別単語追加
  - [x] AddInvisibleText: 簡易透明テキスト追加
  - [x] 透明テキストによるコピー・検索対応
- [x] 座標変換機能
  - [x] ConvertPixelToPDFCoords: ピクセル座標→PDF座標変換
  - [x] ConvertPixelToPDFRect: 矩形座標変換
  - [x] OCRResult.ToTextLayer: OCR結果を自動変換
- [x] OCR API統合サポート
  - [x] Google Cloud Vision API連携例
  - [x] Tesseract OCR連携例
  - [x] 座標系の違いを自動処理
- [x] 包括的なテスト
  - [x] text_layer_test.go: データ構造と座標変換テスト
  - [x] page_text_layer_test.go: 描画機能テスト（全テストパス）
- [x] examples/12_ocr_text_layer サンプル作成
  - [x] シンプルな透明テキスト例
  - [x] OCR結果シミュレーション例
  - [x] 複数単語配置例
  - [x] READMEドキュメント（OCR API統合ガイド付き）
- [x] docs/ocr_text_layer_design.md 設計書作成

**Phase 11 実装成果:**
- 画像ベースPDFを検索・コピー可能にする機能の完全実装
- OCR処理は外部API/ライブラリ使用を想定（gopdfはインターフェースのみ提供）
- ピクセル座標とPDF座標の自動変換
- Google Vision、Tesseractなど主要OCRとの統合ガイド
- 3つの実用的なサンプルと包括的なドキュメント

## 現在の作業状況

Phase 11まで完了し、PDF生成・読み込み・翻訳・ルビ・OCRテキストレイヤーの主要機能が実装されました。
次はPhase 12としてセキュリティ・暗号化機能（パスワード保護）の実装に進みます。

## 予定タスク

### Phase 12: パスワード保護・暗号化機能
- [ ] 暗号化基盤実装
  - [ ] RC4暗号化（40-bit, 128-bit）
  - [ ] AES暗号化（128-bit, 256-bit）
  - [ ] 暗号化ヘルパー関数
- [ ] パスワード保護機能
  - [ ] ユーザーパスワード設定
  - [ ] オーナーパスワード設定
  - [ ] パスワード強度検証
  - [ ] PDF Encrypt辞書の生成
- [ ] アクセス権限制御
  - [ ] 権限フラグの定義（印刷、コピー、編集等）
  - [ ] 権限の設定・適用
  - [ ] セキュリティプリセットの実装
- [ ] パスワード保護PDFの読み込み
  - [ ] 暗号化方式の自動検出
  - [ ] パスワードによる復号化
  - [ ] エラーハンドリング（パスワード不一致等）
- [ ] 統合テスト・サンプル作成

### Phase 13: デフォルト日本語フォント埋め込み
- [ ] フォントバンドル機能
  - [ ] デフォルト日本語フォントの選定（オープンソース）
  - [ ] フォントファイルの埋め込み・管理
  - [ ] ライセンス確認と文書化
- [ ] 自動フォントフォールバック
  - [ ] 文字種判定（ASCII/日本語/その他）
  - [ ] 自動フォント切り替え機能
  - [ ] フォントキャッシュ機構
- [ ] フォントダウンロード機能
  - [ ] Google Fonts APIからの自動取得
  - [ ] フォントキャッシュディレクトリ
  - [ ] オフライン対応
- [ ] 統合テスト・サンプル作成
  - [ ] デフォルトフォントでの日本語描画サンプル
  - [ ] フォント自動切り替えサンプル

### Phase 11: ふりがな（ルビ）機能
- [ ] ルビ描画エンジン実装
  - [ ] ルビのレイアウト計算（親文字との位置関係）
  - [ ] モノルビ対応
  - [ ] グループルビ対応
  - [ ] 縦書き・横書き対応
- [ ] ActualText属性による制御
  - [ ] PDFのActualText属性実装
  - [ ] 表示テキストとコピーテキストの分離
  - [ ] コピーモードの切り替え（漢字のみ/ひらがなのみ/両方）
- [ ] ルビAPI設計
  - [ ] DrawTextWithRuby関数
  - [ ] RubyOptions構造体（フォントサイズ、オフセット等）
  - [ ] ルビペア構造体（親文字、ルビ文字）
- [ ] 既存PDFへのルビ追加機能
  - [ ] テキスト位置の検出
  - [ ] ルビの自動配置
  - [ ] レイアウト調整
- [ ] 自動ルビ振り機能
  - [ ] 形態素解析ライブラリ連携（MeCab等）
  - [ ] 読み仮名辞書の構築
  - [ ] ルビ候補の自動生成
- [ ] 一括ルビ機能
  - [ ] 辞書ファイルフォーマット設計（CSV、JSON）
  - [ ] 辞書読み込み・適用機能
  - [ ] ルビプレビュー機能
- [ ] 統合テスト・サンプル作成
  - [ ] 基本的なルビ付きテキストサンプル
  - [ ] 既存PDFへのルビ追加サンプル
  - [ ] 自動ルビ振りサンプル
  - [ ] コピー動作確認サンプル

### Phase 12: カスタムメタデータ機能
- [ ] メタデータAPI設計
  - [ ] カスタムメタデータの定義構造
  - [ ] SetMetadata/GetMetadata関数
  - [ ] 標準メタデータ対応（Title, Author, Subject等）
- [ ] カスタムプロパティ
  - [ ] ユーザー定義プロパティの追加
  - [ ] プロパティの型定義（文字列、数値、日付等）
  - [ ] プロパティの検索・フィルタリング
- [ ] 埋め込みデータ機能
  - [ ] EmbeddedFiles対応
  - [ ] 任意ファイルの埋め込み
  - [ ] ファイル抽出機能
- [ ] XMP（Extensible Metadata Platform）対応
  - [ ] XMPパケット生成
  - [ ] XMLベースのメタデータ
  - [ ] Dublin Core対応
- [ ] 統合テスト・サンプル作成
  - [ ] メタデータ設定・取得サンプル
  - [ ] ファイル埋め込み・抽出サンプル
  - [ ] XMPメタデータサンプル

### Phase 13: セキュリティ・暗号化機能
- [ ] 暗号化基盤実装
  - [ ] RC4暗号化（40-bit, 128-bit）
  - [ ] AES暗号化（128-bit, 256-bit）
  - [ ] 暗号化ヘルパー関数
- [ ] パスワード保護機能
  - [ ] ユーザーパスワード設定
  - [ ] オーナーパスワード設定
  - [ ] パスワード強度検証
  - [ ] PDF Encrypt辞書の生成
- [ ] アクセス権限制御
  - [ ] 権限フラグの定義（印刷、コピー、編集等）
  - [ ] 権限の設定・適用
  - [ ] セキュリティプリセットの実装
- [ ] パスワード保護PDFの読み込み
  - [ ] 暗号化方式の自動検出
  - [ ] パスワードによる復号化
  - [ ] エラーハンドリング（パスワード不一致等）
- [ ] パスワード解除機能
  - [ ] 完全な暗号化解除
  - [ ] 部分的な権限解除
  - [ ] 操作ログ記録
- [ ] 統合テスト・サンプル作成
  - [ ] パスワード保護PDFの作成・読み込みサンプル
  - [ ] 権限制御サンプル
  - [ ] パスワード解除サンプル

### Phase 14: Markdown変換機能
- [ ] Markdownパーサー実装
  - [ ] CommonMark準拠のパーサー実装（または既存ライブラリ活用）
  - [ ] GFM拡張構文対応（テーブル、タスクリスト等）
  - [ ] フロントマター解析（YAML、TOML）
- [ ] Markdown to PDF変換機能
  - [ ] 見出しレベル対応のレンダリング
  - [ ] 段落・リスト・引用の描画
  - [ ] コードブロック描画（等幅フォント、シンタックスハイライト）
  - [ ] テーブル描画
  - [ ] 画像埋め込み
  - [ ] リンクアノテーション
- [ ] Markdown to Slide変換機能
  - [ ] スライド区切り解析（水平線、見出しベース）
  - [ ] スライドレイアウトエンジン
  - [ ] アスペクト比対応（16:9、4:3）
  - [ ] スライドタイトル・本文配置
  - [ ] テーマシステム（配色、フォント）
- [ ] スタイルカスタマイズ機能
  - [ ] スタイル設定構造体の設計
  - [ ] フォント、色、余白、行間のカスタマイズ
  - [ ] テンプレートシステム
- [ ] 統合テスト・サンプル作成
  - [ ] Markdown to PDF変換サンプル
  - [ ] Markdown to Slide変換サンプル
  - [ ] カスタムスタイル適用例

## 技術的な決定事項

### 採用技術
- **言語**: Go 1.18以上
- **依存関係**: Pure Goのみ
- **標準ライブラリ**: compress/flate, image/jpeg, image/png
- **拡張ライブラリ**: golang.org/x/image/font/sfnt (TTF解析用)

### 設計パターン
- TDD（テスト駆動開発）を採用
- レイヤードアーキテクチャ
- インターフェースによる疎結合設計

### コーディング規約
- gofmt準拠
- golangci-lintでコード品質維持
- テストカバレッジ80%以上を目標

## 課題・検討事項

### 現在の課題
- TTFフォントの手動配置要件（Phase 10で解決予定）
- 複雑なPDFレイアウトの完全再現（継続改善）
- 大容量PDF処理のパフォーマンス最適化

### 今後の検討事項
- [ ] デフォルト日本語フォントの選定とライセンス確認（Phase 10）
- [ ] ルビ機能の自動振り分けエンジン選定（Phase 11）
- [ ] 暗号化方式の優先順位（Phase 13）
- [ ] PDF/A対応の必要性
- [ ] フォーム（AcroForm）サポートの優先度
- [ ] 電子署名機能の実装時期
- [ ] パフォーマンス最適化のベンチマーク基準
- [ ] フォントサブセット化のさらなる最適化

## 参考資料

### PDF仕様
- PDF 1.7 / ISO 32000-1:2008
- Adobe PDF Reference: https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/PDF32000_2008.pdf

### 参考実装
- PyPDF2: https://github.com/py-pdf/pypdf2
- QPDF: https://github.com/qpdf/qpdf
- pdfcpu: https://github.com/pdfcpu/pdfcpu (Go実装)

### 学習リソース
- PDF Syntax 101: https://www.nutrient.io/blog/pdf-syntax-101/
- PyPDF2 Documentation: https://pypdf2.readthedocs.io/en/3.0.0/dev/pdf-format.html

## 更新履歴

### 2024-10-28

**Phase 3-9 完了: 包括的なPDF機能実装**
- **Phase 3**: 図形描画機能（線、矩形、円、ベジェ曲線）
- **Phase 4**: 画像埋め込み（JPEG, PNG）
- **Phase 5**: PDF読み込み・解析機能
- **Phase 6**: 構造化テキスト抽出
- **Phase 7**: 画像抽出機能
- **Phase 8**: TTFフォント・日本語完全対応
- **Phase 9**: PDF翻訳機能（レイアウト保持、可変フォントサイズ、画像保持）

**設計書の充実**
- 各機能の詳細設計書を作成
- CI/CD設計書の追加
- ルビ機能の設計書作成（Phase 11の準備）

**サンプルコードの充実**
- 全10個のサンプルコード作成
- 各サンプルにREADME追加（翻訳機能等）
- 実用的なユースケースの実演

### 2024-10-27

**午前: プロジェクト初期化とPhase 1実装**
- プロジェクト開始
- 設計書作成完了（architecture.md, structure.md, pdf_spec_notes.md）
- 基本的なディレクトリ構造作成完了
- **Phase 1完了**: MVP実装完了
  - Core層、Writer層、Document層の実装
  - 32のユニットテストすべてパス
  - PDFファイル生成成功（examples/01_empty_page）
  - バグ修正: PageオブジェクトにParentフィールド追加（PDF仕様準拠）

**午後: Phase 2実装**
- **Phase 2完了**: テキスト描画機能実装完了
  - Font層の実装（標準Type1フォント14種対応）
  - テキスト描画機能（SetFont, DrawText）
  - 合計42のユニットテストすべてパス
  - Hello Worldサンプル作成（examples/02_hello_world）
  - 複数フォント・サイズでのテキスト描画確認
