package writer

import (
	"fmt"
	"io"

	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/security"
)

// Writer handles PDF document writing and output.
type Writer struct {
	w            io.Writer
	serializer   *Serializer
	offsets      map[int]int64 // オブジェクト番号 -> ファイル内オフセット
	nextObjNum   int           // 次のオブジェクト番号
	bytesWritten int64         // 書き込まれた総バイト数
	encryption   *EncryptionInfo // 暗号化情報（nil = 暗号化なし）
}

// NewWriter creates a new PDF Writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:            w,
		serializer:   NewSerializer(w),
		offsets:      make(map[int]int64),
		nextObjNum:   1,
		bytesWritten: 0,
		encryption:   nil,
	}
}

// SetEncryption sets up encryption for the PDF
func (w *Writer) SetEncryption(encryptionInfo *EncryptionInfo) {
	w.encryption = encryptionInfo
}

// WriteHeader writes the PDF header (%PDF-1.7).
func (w *Writer) WriteHeader() error {
	header := "%PDF-1.7\n"
	n, err := io.WriteString(w.w, header)
	w.bytesWritten += int64(n)
	return err
}

// AddObject adds an object to the PDF and returns its object number.
func (w *Writer) AddObject(obj core.Object) (int, error) {
	objNum := w.nextObjNum
	w.nextObjNum++

	// 暗号化が有効な場合、ストリームオブジェクトを暗号化
	if w.encryption != nil {
		if stream, ok := obj.(*core.Stream); ok {
			obj = w.encryptStream(stream, objNum, 0)
		}
	}

	// 現在のオフセットを記録
	w.offsets[objNum] = w.bytesWritten

	// 間接オブジェクトとして出力
	indirectObj := &core.IndirectObject{
		ObjectNumber:     objNum,
		GenerationNumber: 0,
		Object:           obj,
	}

	// シリアライズ前のバッファで書き込みバイト数をカウント
	var buf countingWriter
	buf.w = w.w
	buf.count = &w.bytesWritten

	tempSerializer := NewSerializer(&buf)
	if err := tempSerializer.SerializeIndirectObject(indirectObj); err != nil {
		return 0, err
	}

	return objNum, nil
}

// encryptStream encrypts a stream object and returns a new stream with encrypted data
func (w *Writer) encryptStream(stream *core.Stream, objectNumber, generationNumber int) *core.Stream {
	// Get key length in bytes
	keyLengthBytes := w.encryption.KeyLength / 8

	// Encrypt the stream data
	encryptedData := security.EncryptStream(
		stream.Data,
		w.encryption.EncryptionKey,
		objectNumber,
		generationNumber,
		keyLengthBytes,
	)

	// Create a new stream with encrypted data
	newDict := make(core.Dictionary)
	for k, v := range stream.Dict {
		newDict[k] = v
	}

	// Update the Length to match encrypted data length
	newDict[core.Name("Length")] = core.Integer(len(encryptedData))

	return &core.Stream{
		Dict: newDict,
		Data: encryptedData,
	}
}

// WriteTrailer writes the xref table and trailer.
func (w *Writer) WriteTrailer(trailer core.Dictionary) error {
	// 暗号化が有効な場合、Encrypt辞書を追加
	if w.encryption != nil {
		// Encrypt辞書をオブジェクトとして追加
		encryptDict := w.encryption.CreateEncryptDictionary()
		encryptNum, err := w.AddObject(encryptDict)
		if err != nil {
			return fmt.Errorf("failed to add Encrypt dictionary: %w", err)
		}

		// TrailerにEncrypt参照を追加
		trailer[core.Name("Encrypt")] = &core.Reference{
			ObjectNumber:     encryptNum,
			GenerationNumber: 0,
		}

		// TrailerにFileID配列を追加
		trailer[core.Name("ID")] = w.encryption.CreateFileIDArray()
	}

	// xrefテーブルの開始位置を記録
	xrefOffset := w.bytesWritten

	// xrefテーブルを書く
	if err := w.writeXRefTable(); err != nil {
		return err
	}

	// trailer辞書を書く
	if err := w.writeTrailerDict(trailer); err != nil {
		return err
	}

	// startxrefを書く
	if err := w.writeStartXRef(xrefOffset); err != nil {
		return err
	}

	// %%EOFを書く
	return w.writeEOF()
}

// writeXRefTable writes the cross-reference table.
func (w *Writer) writeXRefTable() error {
	str := "xref\n"
	n, err := io.WriteString(w.w, str)
	w.bytesWritten += int64(n)
	if err != nil {
		return err
	}

	// サブセクションヘッダー: 0から始まり、nextObjNum個のエントリ
	str = fmt.Sprintf("0 %d\n", w.nextObjNum)
	n, err = io.WriteString(w.w, str)
	w.bytesWritten += int64(n)
	if err != nil {
		return err
	}

	// オブジェクト0（常にfree）
	str = "0000000000 65535 f \n"
	n, err = io.WriteString(w.w, str)
	w.bytesWritten += int64(n)
	if err != nil {
		return err
	}

	// 各オブジェクトのエントリ
	for i := 1; i < w.nextObjNum; i++ {
		offset := w.offsets[i]
		str = fmt.Sprintf("%010d 00000 n \n", offset)
		n, err = io.WriteString(w.w, str)
		w.bytesWritten += int64(n)
		if err != nil {
			return err
		}
	}

	return nil
}

// writeTrailerDict writes the trailer dictionary.
func (w *Writer) writeTrailerDict(trailer core.Dictionary) error {
	str := "trailer\n"
	n, err := io.WriteString(w.w, str)
	w.bytesWritten += int64(n)
	if err != nil {
		return err
	}

	if err := w.serializer.Serialize(trailer); err != nil {
		return err
	}

	str = "\n"
	n, err = io.WriteString(w.w, str)
	w.bytesWritten += int64(n)
	return err
}

// writeStartXRef writes the startxref keyword and offset.
func (w *Writer) writeStartXRef(offset int64) error {
	str := fmt.Sprintf("startxref\n%d\n", offset)
	n, err := io.WriteString(w.w, str)
	w.bytesWritten += int64(n)
	return err
}

// writeEOF writes the end-of-file marker.
func (w *Writer) writeEOF() error {
	str := "%%EOF\n"
	n, err := io.WriteString(w.w, str)
	w.bytesWritten += int64(n)
	return err
}

// countingWriter はバイト数をカウントするio.Writer
type countingWriter struct {
	w     io.Writer
	count *int64
}

func (cw *countingWriter) Write(p []byte) (n int, err error) {
	n, err = cw.w.Write(p)
	*cw.count += int64(n)
	return
}
