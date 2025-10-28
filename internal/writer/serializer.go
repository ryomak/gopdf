// Package writer provides functionality to write PDF documents.
package writer

import (
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/ryomak/gopdf/internal/core"
)

// Serializer converts PDF objects to their textual representation.
type Serializer struct {
	w io.Writer
}

// NewSerializer creates a new Serializer that writes to the given writer.
func NewSerializer(w io.Writer) *Serializer {
	return &Serializer{w: w}
}

// Serialize writes the PDF object to the output in its textual format.
func (s *Serializer) Serialize(obj core.Object) error {
	switch v := obj.(type) {
	case core.Null:
		return s.writeString("null")
	case core.Boolean:
		if v {
			return s.writeString("true")
		}
		return s.writeString("false")
	case core.Integer:
		return s.writeString(strconv.Itoa(int(v)))
	case core.Real:
		return s.serializeReal(float64(v))
	case core.String:
		return s.serializeString(string(v))
	case core.Name:
		return s.serializeName(string(v))
	case core.Array:
		return s.serializeArray(v)
	case core.Dictionary:
		return s.serializeDictionary(v)
	case *core.Reference:
		return s.serializeReference(v)
	case *core.Stream:
		return s.serializeStream(v)
	default:
		return fmt.Errorf("unsupported object type: %T", obj)
	}
}

func (s *Serializer) writeString(str string) error {
	_, err := io.WriteString(s.w, str)
	return err
}

func (s *Serializer) serializeReal(v float64) error {
	// 整数の場合は小数点を省略
	if v == float64(int64(v)) {
		return s.writeString(strconv.FormatInt(int64(v), 10))
	}
	// 実数の場合は適切な精度で表示
	str := strconv.FormatFloat(v, 'f', -1, 64)
	return s.writeString(str)
}

func (s *Serializer) serializeString(v string) error {
	// TODO: エスケープ処理の実装
	// 現在は単純に括弧で囲むのみ
	return s.writeString("(" + v + ")")
}

func (s *Serializer) serializeName(v string) error {
	// TODO: 特殊文字のエスケープ処理
	return s.writeString("/" + v)
}

func (s *Serializer) serializeArray(arr core.Array) error {
	if err := s.writeString("["); err != nil {
		return err
	}

	for i, item := range arr {
		if i > 0 {
			if err := s.writeString(" "); err != nil {
				return err
			}
		}
		if err := s.Serialize(item); err != nil {
			return err
		}
	}

	return s.writeString("]")
}

func (s *Serializer) serializeDictionary(dict core.Dictionary) error {
	if err := s.writeString("<<"); err != nil {
		return err
	}

	// キーをソートして一貫した出力を保証
	keys := make([]string, 0, len(dict))
	for k := range dict {
		keys = append(keys, string(k))
	}
	sort.Strings(keys)

	for i, k := range keys {
		if i > 0 {
			if err := s.writeString(" "); err != nil {
				return err
			}
		}
		// キーを出力
		if err := s.serializeName(k); err != nil {
			return err
		}
		if err := s.writeString(" "); err != nil {
			return err
		}
		// 値を出力
		if err := s.Serialize(dict[core.Name(k)]); err != nil {
			return err
		}
	}

	return s.writeString(">>")
}

func (s *Serializer) serializeReference(ref *core.Reference) error {
	return s.writeString(fmt.Sprintf("%d %d R", ref.ObjectNumber, ref.GenerationNumber))
}

func (s *Serializer) serializeStream(stream *core.Stream) error {
	// 辞書を出力
	if err := s.serializeDictionary(stream.Dict); err != nil {
		return err
	}
	// stream キーワード
	if err := s.writeString("\nstream\n"); err != nil {
		return err
	}
	// データを出力
	if _, err := s.w.Write(stream.Data); err != nil {
		return err
	}
	// endstream キーワード
	return s.writeString("\nendstream")
}

// SerializeIndirectObject writes an indirect object definition.
func (s *Serializer) SerializeIndirectObject(obj *core.IndirectObject) error {
	// オブジェクト定義の開始
	if err := s.writeString(fmt.Sprintf("%d %d obj\n", obj.ObjectNumber, obj.GenerationNumber)); err != nil {
		return err
	}

	// オブジェクトの内容を出力
	if err := s.Serialize(obj.Object); err != nil {
		return err
	}

	// オブジェクト定義の終了
	return s.writeString("\nendobj\n")
}
