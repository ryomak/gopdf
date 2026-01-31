## ゴール
gopdfのツールを作成すること


## 制約
- PureGoで実装すること
- 設計書や調査結果はつど、docsフォルダにmarkdownで残すこと
- docsフォルダの設計書を確認してから開発や設計に進むこと
- Testableなコード意識すること
- TDDを軸に考えてください
- 適宜GitHubにcommit/pushしながら進めて
- 積極的にGenericsを使っていくとよさそう
- テーブルドリブンテストを意識して
- **必ずpush前に`make ci`を実行してテストとlintが成功することを確認すること**


## 必須
- PDFのバイナリの仕様もドキュメントにまとめること
- わからないものはWebで調査すること


## 設計書リファレンス

開発前に確認すべき設計書へのリンク:

### 基盤設計
| ドキュメント | 内容 |
|-------------|------|
| [requirements.md](docs/design/requirements.md) | 機能要件・非機能要件の定義 |
| [architecture.md](docs/design/architecture.md) | アーキテクチャ設計（レイヤー構造） |
| [structure.md](docs/design/structure.md) | プロジェクト構造設計 |
| [progress.md](docs/design/progress.md) | 開発進捗・実装状況 |

### PDF仕様
| ドキュメント | 内容 |
|-------------|------|
| [pdf_spec_notes.md](docs/design/pdf_spec_notes.md) | PDF仕様まとめ（バイナリ構造等） |
| [pdf-objects-and-data-types.md](docs/design/pdf-objects-and-data-types.md) | PDFオブジェクト型の詳細 |
| [coordinate_system_and_ctm_design.md](docs/design/coordinate_system_and_ctm_design.md) | 座標系とCTM変換 |

### 主要機能設計
| ドキュメント | 内容 |
|-------------|------|
| [reader_design.md](docs/design/reader_design.md) | PDF読み込み機能 |
| [text_extraction_design.md](docs/design/text_extraction_design.md) | テキスト抽出 |
| [image_extraction_design.md](docs/design/image_extraction_design.md) | 画像抽出 |
| [ttf_font_support_design.md](docs/design/ttf_font_support_design.md) | TTFフォント対応 |
| [pdf_translation_design.md](docs/design/pdf_translation_design.md) | PDF翻訳（レイアウト保持） |
| [ruby_annotation_design.md](docs/design/ruby_annotation_design.md) | ルビ（ふりがな）機能 |
| [markdown_conversion_design.md](docs/design/markdown_conversion_design.md) | Markdown→PDF変換 |
| [pdf_encryption_design.md](docs/design/pdf_encryption_design.md) | 暗号化・パスワード保護 |
