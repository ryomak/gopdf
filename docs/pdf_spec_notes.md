# PDF バイナリ仕様

## 概要

本ドキュメントは、PDF（Portable Document Format）のバイナリ構造と仕様について、gopdfツール開発のために必要な知識をまとめたものです。

### PDFとは

PDFは、電子文書を表現するための標準化されたファイル形式です。環境に依存せず、作成された環境や閲覧・印刷される環境に関係なく、一貫した表示を実現します。

### 標準規格

- **ISO 32000-1:2008**: PDF 1.7の仕様（2008年公開）
- **ISO 32000-2:2020**: PDF 2.0の仕様（2017年初版、2020年改訂版）

Adobe社が開発したPDF 1.7をISOに寄贈し、ISO標準として公開されました。

**仕様書の入手:**
- PDF 1.7: https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/PDF32000_2008.pdf
- PDF 2.0: PDF Association提供（無料）

### PDFの特徴

**ファイル形式の分類:**
- **単一ファイル形式**: 1つのファイルに全情報を含む
- **バイナリ形式**: バイナリとテキストの混在
- **構造化形式**: 階層的なオブジェクト構造

**PostScript言語ベース:**
PDFはPostScript言語をベースとしており、以下の情報を完全に記述します：
- テキスト
- フォント
- ベクターグラフィックス
- ラスター画像
- 表示に必要なその他の情報

## 1. PDF ファイルの全体構造

PDFファイルは、以下の4つの主要部分から構成される：

```
┌─────────────────────────┐
│  1. Header (ヘッダー)     │
├─────────────────────────┤
│                          │
│  2. Body (本体)           │
│     - Objects            │
│                          │
├─────────────────────────┤
│  3. Xref Table           │
│     (クロスリファレンス)   │
├─────────────────────────┤
│  4. Trailer (トレーラー)  │
│     - startxref          │
│     - %%EOF              │
└─────────────────────────┘
```

### 1.1. Header (ヘッダー)

ファイルの先頭に配置され、PDFのバージョンを宣言する。

**書式:**
```
%PDF-1.x
```

**例:**
```
%PDF-1.7
```

- `%PDF-` に続けてバージョン番号（例: 1.0, 1.4, 1.7 など）
- 第1行目に記述される
- ASCII文字列
- バージョン番号により、サポートする機能が決まる

**バイナリマーカー（オプション）:**
ヘッダーの直後（2行目）に、バイナリデータを含むことを示すコメント行を入れることがある：
```
%PDF-1.7
%âãÏÓ
```
この2行目は、8ビット目が立った文字（128以上のバイト値）を含み、ファイルがバイナリであることをテキストモードで誤って処理されないよう警告する。

### 1.2. Body (本体)

PDFドキュメントの実際のコンテンツを含むオブジェクトの集合。

**特徴:**
- 開始・終了マークなし
- 複数の間接オブジェクト（Indirect Objects）から構成
- オブジェクトはどの順序で配置してもよい（xrefテーブルで位置を管理）
- テキストベース（PostScript由来）だが、ストリームにバイナリデータを含む

### 1.3. Cross-Reference Table (xref)

各間接オブジェクトのファイル内バイト位置を記録したテーブル。

**書式:**
```
xref
<starting_object_id> <number_of_entries>
<10-digit offset> <5-digit generation> <n or f>
<10-digit offset> <5-digit generation> <n or f>
...
```

**例:**
```
xref
0 6
0000000000 65535 f
0000000015 00000 n
0000000074 00000 n
0000000182 00000 n
0000000251 00000 n
0000000511 00000 n
```

**解説:**
- `xref`: キーワード
- `0 6`: オブジェクト番号0から始まる6エントリ
- 各行は20バイト固定長（10桁 + 空白 + 5桁 + 空白 + 1文字 + 空白 + 改行）
- **10桁のバイトオフセット**: ファイル先頭からのオブジェクト位置
- **5桁の世代番号**: オブジェクトの世代（通常は00000）
- **n または f**:
  - `n`: 使用中（in use）
  - `f`: 空き（free）
- オブジェクト0は常に空き（free）で、世代番号は65535

**サブセクション:**
複数のサブセクションを持つことも可能：
```
xref
0 1
0000000000 65535 f
3 2
0000025325 00000 n
0000025518 00000 n
```
上記は、オブジェクト0と、オブジェクト3, 4の情報を持つ。

### 1.4. Trailer (トレーラー)

PDFファイルのメタ情報と、ドキュメントのルートへの参照を含む。

**書式:**
```
trailer
<< trailer dictionary >>
startxref
<xref table byte offset>
%%EOF
```

**例:**
```
trailer
<<
  /Size 6
  /Root 1 0 R
  /Info 5 0 R
>>
startxref
531
%%EOF
```

**主要なキー:**
- `/Size`: xrefテーブルのエントリ数（最大オブジェクト番号+1）
- `/Root`: Catalogオブジェクトへの参照（必須）
- `/Info`: Infoオブジェクトへの参照（オプション、メタデータ）
- `/Prev`: 前回更新時のxrefテーブルのオフセット（増分更新時）
- `/Encrypt`: 暗号化辞書への参照（暗号化PDFの場合）
- `/ID`: ファイル識別子の配列

**startxref:**
- xrefテーブルの開始位置（ファイル先頭からのバイトオフセット）を示す数値

**%%EOF:**
- ファイルの終端マーカー

### 1.5. ファイル読み取りフロー

1. ファイル末尾から `%%EOF` を探す
2. `startxref` キーワードを見つけ、その次の数値を読む（xrefオフセット）
3. xrefオフセット位置にシークし、xrefテーブルを読む
4. trailerを読み、`/Root` からCatalogオブジェクトを取得
5. Catalogを起点に、必要なオブジェクトを順次読み込む

## 2. PDFオブジェクト型

PDFはPostScript由来のシンタックスを持ち、以下のオブジェクト型をサポートする。

### 2.1. 基本型

#### 2.1.1. Null
```
null
```
値がないことを示す。

#### 2.1.2. Boolean
```
true
false
```

#### 2.1.3. Numeric
整数または実数。

```
42          % 整数
-17         % 負の整数
3.14159     % 実数
-0.001      % 負の実数
```

#### 2.1.4. String
2つの表記法がある：

**リテラル文字列（Literal String）:**
```
(Hello, World!)
(これは文字列です)
(括弧\(エスケープ\)も使える)
```
- 括弧 `( )` で囲む
- エスケープ文字: `\n`, `\r`, `\t`, `\\`, `\(`, `\)`
- 改行を含めることも可能

**16進数文字列（Hexadecimal String）:**
```
<48656C6C6F>    % "Hello" のASCII
<>              % 空文字列
```
- 山括弧 `< >` で囲む
- 16進数で各バイトを表現

#### 2.1.5. Name
名前型。キー、識別子として使用。

```
/Type
/Font
/Helvetica
/Page
```
- スラッシュ `/` で始まる
- 空白、特殊文字は `#` でエスケープ（例: `/A#20B` は "A B"）

### 2.2. 複合型

#### 2.2.1. Array
順序付きコレクション。

```
[1 2 3 4 5]
[/Type /Page]
[(string) 123 true [nested array]]
```
- 角括弧 `[ ]` で囲む
- 異なる型を混在可能
- ネスト可能

#### 2.2.2. Dictionary
キーと値のペア。

```
<< /Type /Page
   /Parent 2 0 R
   /MediaBox [0 0 612 792]
>>
```
- 二重山括弧 `<< >>` で囲む
- キーは必ずName型
- 値は任意のオブジェクト型

#### 2.2.3. Stream
辞書とバイナリデータのペア。画像、フォント、コンテンツなどに使用。

```
<< /Length 44 >>
stream
BT
/F1 12 Tf
100 700 Td
(Hello, World!) Tj
ET
endstream
```

**重要なキー:**
- `/Length`: ストリームのバイト長（必須）
- `/Filter`: 圧縮・エンコーディングの種類（例: `/FlateDecode`）
- `/DecodeParms`: デコードパラメータ

**フィルタの例:**
- `/FlateDecode`: zlib/deflate圧縮（最も一般的）
- `/DCTDecode`: JPEG圧縮
- `/LZWDecode`: LZW圧縮
- `/ASCII85Decode`, `/ASCIIHexDecode`: ASCII エンコーディング

### 2.3. 間接オブジェクト（Indirect Object）

再利用可能なオブジェクトで、参照を介してアクセスする。

**定義:**
```
<object_number> <generation_number> obj
<object>
endobj
```

**例:**
```
1 0 obj
<< /Type /Catalog
   /Pages 2 0 R
>>
endobj
```

**参照:**
```
<object_number> <generation_number> R
```

**例:**
```
2 0 R     % オブジェクト2への参照
```

- **オブジェクト番号**: 1から始まる整数（0は予約済み）
- **世代番号**: 通常は0（オブジェクトの更新履歴管理用）
- **R**: Referenceの意味

## 3. 主要な間接オブジェクト

### 3.1. Catalog（カタログ）

PDFドキュメントのルート。

```
1 0 obj
<< /Type /Catalog
   /Pages 2 0 R
   /Metadata 10 0 R
>>
endobj
```

**必須キー:**
- `/Type`: 必ず `/Catalog`
- `/Pages`: ページツリーのルートへの参照

**オプションキー:**
- `/Metadata`: XMPメタデータ
- `/Outlines`: ブックマーク
- `/PageMode`, `/PageLayout`: 表示モード
- `/Names`: 名前付きリソース

### 3.2. Pages（ページツリー）

ページをツリー構造で管理。

```
2 0 obj
<< /Type /Pages
   /Kids [3 0 R 4 0 R 5 0 R]
   /Count 3
>>
endobj
```

**必須キー:**
- `/Type`: 必ず `/Pages`
- `/Kids`: 子ノード（ページまたはページツリー）の配列
- `/Count`: 子孫ページの総数

**オプションキー:**
- `/Parent`: 親ノードへの参照（ルート以外）
- `/MediaBox`: デフォルトのページサイズ（子に継承）
- `/Resources`: リソース辞書（子に継承）

### 3.3. Page（ページ）

個々のページ。

```
3 0 obj
<< /Type /Page
   /Parent 2 0 R
   /MediaBox [0 0 612 792]
   /Contents 4 0 R
   /Resources << /Font << /F1 5 0 R >> >>
>>
endobj
```

**必須キー:**
- `/Type`: 必ず `/Page`
- `/Parent`: 親ページツリーへの参照
- `/MediaBox`: ページサイズ（継承されていない場合）
- `/Resources`: フォント、画像などのリソース辞書

**オプションキー:**
- `/Contents`: コンテンツストリームへの参照（または配列）
- `/CropBox`, `/BleedBox`, `/TrimBox`, `/ArtBox`: 各種ボックス
- `/Rotate`: 回転角度（0, 90, 180, 270）

### 3.4. Content Stream（コンテンツストリーム）

ページに描画する内容。

```
4 0 obj
<< /Length 44 >>
stream
BT
/F1 12 Tf
100 700 Td
(Hello, World!) Tj
ET
endstream
endobj
```

**主要オペレーター:**

| オペレーター | 説明 |
|------------|------|
| `BT` ... `ET` | テキストオブジェクトの開始・終了 |
| `/F1 12 Tf` | フォント選択（F1, サイズ12） |
| `100 700 Td` | テキスト位置移動（x=100, y=700） |
| `(text) Tj` | テキスト表示 |
| `q` ... `Q` | グラフィック状態の保存・復元 |
| `cm` | 変換マトリックス設定 |
| `m`, `l`, `c`, `h`, `re` | パスの構築（移動、線、曲線、閉じる、矩形） |
| `S`, `s`, `f`, `F`, `f*`, `B`, `B*`, `b`, `b*` | パスの描画（ストローク、塗りつぶし） |
| `w` | 線の太さ |
| `RG`, `rg` | 色設定（RGB） |
| `Do` | XObject（画像など）の描画 |

### 3.5. Resources（リソース辞書）

ページで使用するリソース。

```
<< /Font << /F1 5 0 R
            /F2 6 0 R >>
   /XObject << /Im1 7 0 R >>
   /ProcSet [/PDF /Text /ImageC]
>>
```

**主要キー:**
- `/Font`: フォント辞書
- `/XObject`: 画像などの外部オブジェクト
- `/ColorSpace`: カラースペース
- `/Pattern`: パターン
- `/ExtGState`: 拡張グラフィック状態

### 3.6. Font（フォント）

```
5 0 obj
<< /Type /Font
   /Subtype /Type1
   /BaseFont /Helvetica
>>
endobj
```

**標準Type1フォント（埋め込み不要）:**
- Helvetica, Helvetica-Bold, Helvetica-Oblique, Helvetica-BoldOblique
- Times-Roman, Times-Bold, Times-Italic, Times-BoldItalic
- Courier, Courier-Bold, Courier-Oblique, Courier-BoldOblique
- Symbol, ZapfDingbats

**TrueTypeフォント（埋め込み）:**
```
<< /Type /Font
   /Subtype /TrueType
   /BaseFont /ABCDEF+CustomFont
   /FontDescriptor 8 0 R
   /Encoding /WinAnsiEncoding
>>
```

### 3.7. Info（情報辞書）

メタデータ。

```
5 0 obj
<< /Title (My Document)
   /Author (John Doe)
   /Subject (PDF Example)
   /Creator (MyApp v1.0)
   /Producer (gopdf v0.1)
   /CreationDate (D:20241027120000+09'00')
   /ModDate (D:20241027120000+09'00')
>>
endobj
```

**日付フォーマット:**
```
D:YYYYMMDDHHmmSSOHH'mm'
```
例: `D:20241027153045+09'00'`（2024年10月27日 15:30:45 JST）

## 4. 最小限のPDFファイル例

以下は、"Hello, World!" を表示する最小限のPDF：

```pdf
%PDF-1.7

1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj

2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj

3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792]
   /Contents 4 0 R
   /Resources << /Font << /F1 5 0 R >> >>
>>
endobj

4 0 obj
<< /Length 44 >>
stream
BT
/F1 12 Tf
100 700 Td
(Hello, World!) Tj
ET
endstream
endobj

5 0 obj
<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>
endobj

xref
0 6
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
0000000262 00000 n
0000000359 00000 n
trailer
<< /Size 6 /Root 1 0 R >>
startxref
441
%%EOF
```

**構造解説:**
1. オブジェクト1: Catalog（ルート）
2. オブジェクト2: Pages（ページツリー）
3. オブジェクト3: Page（1ページ目）
4. オブジェクト4: コンテンツストリーム（テキスト描画命令）
5. オブジェクト5: Font（Helveticaフォント）

## 5. 座標系

PDFの座標系は左下が原点 (0, 0)。

```
      Y
      ↑
      |
      |
      |
(0,0) └─────────→ X
```

**単位:**
- デフォルトはポイント（pt）: 1 pt = 1/72 inch
- 1 inch = 72 pt
- A4サイズ: 595 x 842 pt (210mm x 297mm)
- Letter: 612 x 792 pt (8.5" x 11")

## 6. 実装のポイント

### 6.1. Writer実装時
1. ヘッダーを出力（`%PDF-1.7`）
2. オブジェクトを順次出力し、各オブジェクトの開始位置を記録
3. すべてのオブジェクト出力後、xrefテーブルを生成
4. trailerを出力
5. startxrefとxrefテーブルのオフセットを出力
6. `%%EOF` で終了

### 6.2. Reader実装時
1. ファイル末尾から `%%EOF` を探す
2. `startxref` を読み、xrefオフセットを取得
3. xrefテーブルを解析し、オブジェクトマップを構築
4. trailerから `/Root` を取得
5. Catalogから Pages を辿る
6. 必要に応じて間接オブジェクトを解決

### 6.3. ストリーム圧縮
- `/FlateDecode` を使用（Go標準の `compress/flate` で対応可能）
- 圧縮前のデータ長を `/Length` に設定
- 圧縮フィルタを `/Filter` に指定

### 6.4. エラーハンドリング
- 不正なフォーマットに対する堅牢性
- オブジェクト参照の循環を検出
- xrefテーブルの破損に対応（リニアライズPDFも考慮）

## 7. 参考リソース

**公式仕様書:**
- ISO 32000-1:2008 - PDF 1.7仕様書
- ISO 32000-2:2020 - PDF 2.0仕様書
- [Adobe PDF Reference](https://opensource.adobe.com/dc-acrobat-sdk-docs/)
- [PDF Association](https://pdfa.org/)

**実装ガイド:**
- [PDF Syntax 101](https://www.nutrient.io/blog/pdf-syntax-101/)
- [Guillaume Endignoux - Introduction to PDF syntax](https://gendignoux.com/blog/2016/10/04/pdf-basics.html)
- [PyPDF2 Documentation](https://pypdf2.readthedocs.io/)

**オープンソースライブラリ:**
- [QPDF Library](https://github.com/qpdf/qpdf) - C++実装
- PyPDF2 - Python実装

## 関連ドキュメント

- [PDFオブジェクトとデータ型](./pdf-objects-and-data-types.md) - オブジェクトタイプの詳細説明
