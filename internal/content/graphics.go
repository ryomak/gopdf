package content

import (
	"bytes"
	"fmt"

	"github.com/ryomak/gopdf/internal/core"
)

// ExtractGraphicsOperations は操作リストからグラフィックス操作（テキスト・画像以外）を抽出し、
// コンテンツストリームのバイト列として返す。
// BT...ET ブロック（テキスト）と Do（画像描画）は除外される。
// q/Q（グラフィックス状態の保存・復元）、cm（変換行列）、
// re/m/l/c/S/f/b/W 等のパス・描画操作、w/J/j/M/d（線スタイル）、
// RG/rg/K/k/SC/sc/G/g（色設定）、gs（拡張グラフィックス状態）など
// すべてのグラフィックス操作が保持される。
func ExtractGraphicsOperations(operations []Operation) []byte {
	var buf bytes.Buffer
	inText := false // BT...ET ブロック内かどうか

	for _, op := range operations {
		// テキストブロックの開始・終了を追跡
		if op.Operator == "BT" {
			inText = true
			continue
		}
		if op.Operator == "ET" {
			inText = false
			continue
		}

		// テキストブロック内の操作はスキップ
		if inText {
			continue
		}

		// Do（XObject描画 = 画像）はスキップ（画像は別途処理される）
		if op.Operator == "Do" {
			continue
		}

		// グラフィックス操作をストリームに書き出す
		writeOperation(&buf, op)
	}

	return buf.Bytes()
}

// writeOperation は単一のOperationをコンテンツストリーム形式で書き出す
func writeOperation(buf *bytes.Buffer, op Operation) {
	for _, operand := range op.Operands {
		writeOperand(buf, operand)
		buf.WriteByte(' ')
	}
	buf.WriteString(op.Operator)
	buf.WriteByte('\n')
}

// writeOperand はオペランドをPDFコンテンツストリーム形式で書き出す
func writeOperand(buf *bytes.Buffer, obj core.Object) {
	switch v := obj.(type) {
	case core.Integer:
		fmt.Fprintf(buf, "%d", int(v))
	case core.Real:
		fmt.Fprintf(buf, "%g", float64(v))
	case core.String:
		buf.WriteByte('(')
		// 特殊文字をエスケープ
		for _, b := range []byte(v) {
			switch b {
			case '(', ')', '\\':
				buf.WriteByte('\\')
				buf.WriteByte(b)
			default:
				buf.WriteByte(b)
			}
		}
		buf.WriteByte(')')
	case core.Name:
		fmt.Fprintf(buf, "/%s", string(v))
	case core.Boolean:
		if bool(v) {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case core.Array:
		buf.WriteByte('[')
		for i, item := range v {
			if i > 0 {
				buf.WriteByte(' ')
			}
			writeOperand(buf, item)
		}
		buf.WriteByte(']')
	case core.Dictionary:
		buf.WriteString("<< ")
		for key, val := range v {
			fmt.Fprintf(buf, "/%s ", string(key))
			writeOperand(buf, val)
			buf.WriteByte(' ')
		}
		buf.WriteString(">>")
	default:
		buf.WriteString("null")
	}
}
