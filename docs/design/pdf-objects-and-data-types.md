# PDFオブジェクトとデータ型

## 概要

PDFファイルは、9種類の基本オブジェクトタイプで構成されています。これらはオブジェクト指向プログラミングのオブジェクトではなく、「PDFが立脚する構成要素（building blocks）」です。

## PDFの基本オブジェクトタイプ

PDF仕様では、以下の9種類の基本オブジェクトタイプが定義されています：

1. **Null**
2. **Boolean（ブール値）**
3. **Integer（整数）**
4. **Real（実数）**
5. **String（文字列）**
6. **Name（名前）**
7. **Array（配列）**
8. **Dictionary（辞書）**
9. **Stream（ストリーム）**

## 各オブジェクトタイプの詳細

### 1. Null

欠落した値や未定義の値を表します。

**構文:**
```
null
```

**特徴:**
- ファイルに明示的に書き込まれることは稀
- 値が存在しないことを示す

**使用例:**
```
<< /OptionalKey null >>
```

---

### 2. Boolean（ブール値）

論理値を表現します。

**構文:**
```
true
false
```

**例:**
```
<< /Visible true /Enabled false >>
```

---

### 3. Integer（整数）

整数値を表現します。

**構文:**
```
123
-456
0
+17
```

**特徴:**
- 符号付き整数
- 範囲: 通常 -2,147,483,648 〜 2,147,483,647（32ビット整数）

**例:**
```
<< /Count 10 /Width 595 /Height 842 >>
```

---

### 4. Real（実数）

浮動小数点数を表現します。

**構文:**
```
1.23
-0.456
.5
-.002
```

**特徴:**
- 小数点を含む数値
- **指数表記（1.23e4など）は使用不可**
- 小数点のみの表記も可能（`.5` = `0.5`）

**例:**
```
<< /LineWidth 0.5 /Scale 1.2 >>
```

---

### 5. String（文字列）

テキストデータを表現します。2つの形式があります。

#### リテラル文字列（Literal String）

**構文:**
```
(Hello)
(Hello world\n)
(Text with \(parentheses\))
```

**特徴:**
- 括弧 `( )` で囲む
- バックスラッシュ `\` でエスケープ

**エスケープシーケンス:**
- `\n`: 改行（Line Feed）
- `\r`: 復帰（Carriage Return）
- `\t`: タブ
- `\b`: バックスペース
- `\f`: フォームフィード
- `\\`: バックスラッシュ
- `\(`: 開き括弧
- `\)`: 閉じ括弧
- `\ddd`: 8進数表記（例: `\101` = 'A'）

#### 16進文字列（Hexadecimal String）

**構文:**
```
<48656C6C6F>
<48656C6C6F20776F726C64>
```

**特徴:**
- 山括弧 `< >` で囲む
- 各バイトを16進数2桁で表現
- `48656C6C6F` = "Hello"

**例:**
```
1 0 obj
<< /Title (PDF\040Specification)
   /Author <416C696365>
>>
endobj
```

---

### 6. Name（名前）

識別子として使用される一意の名前です。

**構文:**
```
/Name
/Type
/Font1
/Hello#20World
```

**特徴:**
- スラッシュ `/` で始まる
- 特殊文字は `#` でエスケープ（16進数2桁）
- `/Hello#20World` = "/Hello World"

**一般的な使用例:**
```
/Type /Page
/Type /Font
/Type /Catalog
```

**禁則文字のエスケープ:**
- 空白: `#20`
- `#`: `#23`
- `/`: `#2F`
- `(`: `#28`
- `)`: `#29`

**例:**
```
<< /Type /Font
   /BaseFont /Helvetica
   /Subtype /Type1
>>
```

---

### 7. Array（配列）

順序付きのオブジェクトコレクションです。

**構文:**
```
[123 456 789]
[(Hello) (World)]
[1 2 3 [4 5 6]]
[[123 (foo)] /bar true 45.6]
```

**特徴:**
- 角括弧 `[ ]` で囲む
- 要素はスペースで区切る
- 異なる型の要素を混在可能
- ネスト可能

**使用例:**
```
% 矩形の座標 [左下x 左下y 右上x 右上y]
/MediaBox [0 0 595 842]

% RGB色値 [R G B]
/Color [1.0 0.0 0.0]

% ページ参照の配列
/Kids [3 0 R 4 0 R 5 0 R]
```

---

### 8. Dictionary（辞書）

キーと値のペアの集合です。

**構文:**
```
<< /key (value) >>
<< /Type /Page /MediaBox [0 0 612 792] >>
```

**特徴:**
- 二重山括弧 `<< >>` で囲む
- キーは必ず Name オブジェクト
- 値は任意のオブジェクトタイプ
- 順序は保証されない

**複雑な例:**
```
<< /Type /Font
   /Subtype /Type1
   /BaseFont /Helvetica
   /Encoding /WinAnsiEncoding
   /FirstChar 32
   /LastChar 126
   /Widths [278 278 355 556 556 889 667 191 333 333 389 584]
>>
```

**ネストした辞書:**
```
<< /Type /Page
   /Parent 2 0 R
   /MediaBox [0 0 612 792]
   /Contents 4 0 R
   /Resources << /Font << /F1 5 0 R >> >>
>>
```

---

### 9. Stream（ストリーム）

辞書とバイト列を組み合わせた特殊なオブジェクトです。

**構文:**
```
7 0 obj
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

**構成要素:**
1. ストリーム辞書（Stream Dictionary）
2. `stream` キーワード
3. バイトデータ
4. `endstream` キーワード

**ストリーム辞書の必須キー:**
- `/Length`: ストリームデータのバイト数

**ストリーム辞書の重要なオプションキー:**
- `/Filter`: 適用された圧縮フィルター
- `/DecodeParms`: フィルターのパラメータ

**重要な特徴:**
- **ストリームオブジェクトは必ず間接オブジェクトとして定義**
- 大量のバイナリデータを効率的に格納
- 圧縮可能

#### 圧縮フィルター

**主要な圧縮フィルター:**
- `/FlateDecode`: Deflate圧縮（zlib互換、最も一般的）
- `/LZWDecode`: LZW圧縮
- `/DCTDecode`: JPEG圧縮
- `/CCITTFaxDecode`: CCITT Fax圧縮
- `/ASCII85Decode`: ASCII85エンコーディング
- `/ASCIIHexDecode`: ASCII 16進エンコーディング

**圧縮されたストリームの例:**
```
10 0 obj
<< /Length 534
   /Filter /FlateDecode
>>
stream
[圧縮されたバイナリデータ]
endstream
endobj
```

**複数フィルターの適用:**
```
<< /Length 512
   /Filter [/ASCII85Decode /FlateDecode]
>>
```

#### ストリームの用途

1. **ページコンテンツストリーム**: ページ上のグラフィックス操作
2. **画像データストリーム**: 画像のピクセルデータ
3. **フォントデータストリーム**: 埋め込みフォント
4. **メタデータストリーム**: XMPメタデータ
5. **オブジェクトストリーム**: 複数オブジェクトの圧縮（PDF 1.5以降）

## 直接オブジェクトと間接オブジェクト

### 直接オブジェクト（Direct Object）

使用される場所にインラインで配置されます。

**例:**
```
<< /Type /Page
   /MediaBox [0 0 612 792]
   /Count 1
>>
```

### 間接オブジェクト（Indirect Object）

ポインタのような参照機能を持つオブジェクトです。

**定義の構文:**
```
オブジェクト番号 世代番号 obj
  [オブジェクトの内容]
endobj
```

**参照の構文:**
```
オブジェクト番号 世代番号 R
```

**例:**
```
% オブジェクトの定義
3 0 obj
<< /Type /Font
   /BaseFont /Helvetica
   /Subtype /Type1
>>
endobj

% オブジェクトの参照
<< /Font << /F1 3 0 R >> >>
```

**間接オブジェクトの識別子:**
- **オブジェクト番号**: 1から始まる正の整数
- **世代番号**: 通常0（インクリメンタル更新で増加）

**メリット:**
1. **データ共有**: 同じオブジェクトを複数箇所から参照
2. **木構造の作成**: ページツリーなどの階層構造
3. **ファイルサイズの削減**: 重複データの排除

**制約:**
- ストリームオブジェクトは必ず間接オブジェクトとして定義

## オブジェクトストリーム（PDF 1.5以降）

複数の間接オブジェクトを1つの圧縮ストリームにまとめる機能です。

**メリット:**
- ファイルサイズの大幅な削減
- 実運用で数十〜数百のオブジェクトを1つのストリームに格納可能

**構文:**
```
100 0 obj
<< /Type /ObjStm
   /N 10
   /First 50
   /Length 500
   /Filter /FlateDecode
>>
stream
[圧縮された複数オブジェクト]
endstream
endobj
```

**キー:**
- `/Type`: `/ObjStm`（オブジェクトストリーム）
- `/N`: 含まれるオブジェクトの数
- `/First`: 最初のオブジェクトのオフセット

## 実例: Hello Worldを表示する最小限のPDF

```pdf
%PDF-1.7

1 0 obj
<< /Type /Catalog
   /Pages 2 0 R
>>
endobj

2 0 obj
<< /Type /Pages
   /Kids [3 0 R]
   /Count 1
>>
endobj

3 0 obj
<< /Type /Page
   /Parent 2 0 R
   /MediaBox [0 0 612 792]
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
<< /Type /Font
   /Subtype /Type1
   /BaseFont /Helvetica
>>
endobj

xref
0 6
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
0000000264 00000 n
0000000367 00000 n
trailer
<< /Size 6
   /Root 1 0 R
>>
startxref
457
%%EOF
```

**オブジェクトの役割:**
- **1 0 obj**: Catalog（ルートオブジェクト）
- **2 0 obj**: Pages（ページツリーのルート）
- **3 0 obj**: Page（個別ページ）
- **4 0 obj**: Stream（ページコンテンツ）
- **5 0 obj**: Font（フォント定義）

## データ型の比較表

| タイプ | 例 | 主な用途 |
|--------|-----|----------|
| Null | `null` | 欠落値 |
| Boolean | `true` `false` | フラグ |
| Integer | `123` `-456` | カウント、座標 |
| Real | `1.23` `0.5` | 寸法、色値 |
| String | `(Hello)` `<48656C6C6F>` | テキスト、ID |
| Name | `/Type` `/Font1` | 識別子、キー |
| Array | `[1 2 3]` | リスト、座標 |
| Dictionary | `<< /key value >>` | 構造データ |
| Stream | 辞書+データ | コンテンツ、画像 |

## 実装時の注意点

### 1. パーサー実装
- 空白文字（スペース、タブ、改行）は区切り文字
- コメント: `%` から行末まで
- バイトオーダーマーク（BOM）に注意

### 2. 文字列の扱い
- PDFDocEncoding と UTF-16BE の区別
- BOMの有無で判定（`\xFE\xFF` がUTF-16BEのBOM）

### 3. ストリームの処理
- `/Length` を信頼せず、`endstream` を探索する実装も検討
- 圧縮フィルターのチェーン処理に対応

### 4. 間接参照の解決
- 循環参照の検出
- オブジェクトキャッシュの実装

### 5. エラーハンドリング
- 不正な構文への対応
- 部分的に破損したPDFの復旧

## 参考資料

- ISO 32000-1:2008 - Section 7: Syntax
- ISO 32000-2:2020 - PDF 2.0仕様書
- [Adobe PDF Reference](https://opensource.adobe.com/dc-acrobat-sdk-docs/)
- [Guillaume Endignoux - Introduction to PDF syntax](https://gendignoux.com/blog/2016/10/04/pdf-basics.html)

## 関連ドキュメント

- [PDFバイナリ仕様](./pdf_spec_notes.md) - ファイル構造と主要オブジェクト
- [次のステップ: ページ構造とコンテンツストリーム]
