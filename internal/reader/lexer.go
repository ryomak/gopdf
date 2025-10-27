package reader

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// TokenType はトークンの種類
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenInteger      // 123
	TokenReal         // 3.14
	TokenString       // (text) or <hex>
	TokenName         // /Name
	TokenKeyword      // obj, endobj, stream, etc.
	TokenDictStart    // <<
	TokenDictEnd      // >>
	TokenArrayStart   // [
	TokenArrayEnd     // ]
	TokenRef          // R
	TokenBoolean      // true, false
	TokenNull         // null
)

// Token はトークン
type Token struct {
	Type  TokenType
	Value interface{} // string, int, float64, bool など
}

// Lexer はPDFバイトストリームをトークン化する
type Lexer struct {
	r      *bufio.Reader
	peeked []byte // 先読みバッファ
}

// NewLexer は新しいLexerを作成する
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		r:      bufio.NewReader(r),
		peeked: make([]byte, 0),
	}
}

// NextToken は次のトークンを返す
func (l *Lexer) NextToken() (Token, error) {
	// 空白とコメントをスキップ
	if err := l.skipWhitespaceAndComments(); err != nil {
		if err == io.EOF {
			return Token{Type: TokenEOF}, nil
		}
		return Token{}, err
	}

	// 次の文字を先読み
	b, err := l.peekByte()
	if err != nil {
		if err == io.EOF {
			return Token{Type: TokenEOF}, nil
		}
		return Token{}, err
	}

	// トークンの種類を判定
	switch b {
	case '<':
		// << または <hex string>
		return l.readDictStartOrHexString()
	case '>':
		// >>
		return l.readDictEnd()
	case '[':
		l.readByte()
		return Token{Type: TokenArrayStart}, nil
	case ']':
		l.readByte()
		return Token{Type: TokenArrayEnd}, nil
	case '(':
		return l.readLiteralString()
	case '/':
		return l.readName()
	case '+', '-', '.', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return l.readNumber()
	default:
		// キーワード、true, false, null, R
		return l.readKeyword()
	}
}

// skipWhitespaceAndComments は空白文字とコメントをスキップする
func (l *Lexer) skipWhitespaceAndComments() error {
	for {
		b, err := l.peekByte()
		if err != nil {
			return err
		}

		// 空白文字
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\x00' || b == '\x0c' {
			l.readByte()
			continue
		}

		// コメント
		if b == '%' {
			// 改行まで読み飛ばす
			for {
				b, err := l.readByte()
				if err != nil {
					return err
				}
				if b == '\n' || b == '\r' {
					break
				}
			}
			continue
		}

		// 空白でもコメントでもない
		break
	}

	return nil
}

// readDictStartOrHexString は << または <hex> を読む
func (l *Lexer) readDictStartOrHexString() (Token, error) {
	l.readByte() // '<'
	b2, err := l.peekByte()
	if err != nil {
		return Token{}, err
	}

	if b2 == '<' {
		// <<
		l.readByte()
		return Token{Type: TokenDictStart}, nil
	}

	// Hex string <...>
	var buf bytes.Buffer
	for {
		b, err := l.readByte()
		if err != nil {
			return Token{}, err
		}
		if b == '>' {
			break
		}
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			buf.WriteByte(b)
		}
	}

	// 16進数文字列をデコード
	hexStr := buf.String()
	decoded, err := decodeHexString(hexStr)
	if err != nil {
		return Token{}, err
	}

	return Token{Type: TokenString, Value: decoded}, nil
}

// readDictEnd は >> を読む
func (l *Lexer) readDictEnd() (Token, error) {
	l.readByte() // '>'
	b2, err := l.readByte()
	if err != nil {
		return Token{}, err
	}
	if b2 != '>' {
		return Token{}, fmt.Errorf("expected '>', got %c", b2)
	}
	return Token{Type: TokenDictEnd}, nil
}

// readLiteralString は (text) を読む
func (l *Lexer) readLiteralString() (Token, error) {
	l.readByte() // '('

	var buf bytes.Buffer
	depth := 1 // 括弧のネストレベル

	for depth > 0 {
		b, err := l.readByte()
		if err != nil {
			return Token{}, err
		}

		if b == '\\' {
			// エスケープシーケンス
			next, err := l.readByte()
			if err != nil {
				return Token{}, err
			}
			switch next {
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case '\\':
				buf.WriteByte('\\')
			case '(':
				buf.WriteByte('(')
			case ')':
				buf.WriteByte(')')
			default:
				buf.WriteByte(next)
			}
		} else if b == '(' {
			depth++
			buf.WriteByte(b)
		} else if b == ')' {
			depth--
			if depth > 0 {
				buf.WriteByte(b)
			}
		} else {
			buf.WriteByte(b)
		}
	}

	return Token{Type: TokenString, Value: buf.String()}, nil
}

// readName は /Name を読む
func (l *Lexer) readName() (Token, error) {
	l.readByte() // '/'

	var buf bytes.Buffer
	for {
		b, err := l.peekByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return Token{}, err
		}

		// 区切り文字または空白で終了
		if isDelimiter(b) || isWhitespace(b) {
			break
		}

		l.readByte()

		// # エスケープ処理
		if b == '#' {
			// 次の2文字を16進数として読む
			h1, _ := l.readByte()
			h2, _ := l.readByte()
			hexStr := string([]byte{h1, h2})
			val, err := strconv.ParseInt(hexStr, 16, 32)
			if err != nil {
				return Token{}, err
			}
			buf.WriteByte(byte(val))
		} else {
			buf.WriteByte(b)
		}
	}

	return Token{Type: TokenName, Value: buf.String()}, nil
}

// readNumber は数値（整数または実数）を読む
func (l *Lexer) readNumber() (Token, error) {
	var buf bytes.Buffer

	for {
		b, err := l.peekByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return Token{}, err
		}

		if (b >= '0' && b <= '9') || b == '+' || b == '-' || b == '.' {
			l.readByte()
			buf.WriteByte(b)
		} else {
			break
		}
	}

	str := buf.String()

	// 小数点が含まれていれば実数
	if bytes.Contains([]byte(str), []byte{'.'}) {
		val, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return Token{}, err
		}
		return Token{Type: TokenReal, Value: val}, nil
	}

	// 整数
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return Token{}, err
	}
	return Token{Type: TokenInteger, Value: int(val)}, nil
}

// readKeyword はキーワードを読む
func (l *Lexer) readKeyword() (Token, error) {
	var buf bytes.Buffer

	for {
		b, err := l.peekByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return Token{}, err
		}

		if isDelimiter(b) || isWhitespace(b) {
			break
		}

		l.readByte()
		buf.WriteByte(b)
	}

	keyword := buf.String()

	// 特殊なキーワード
	switch keyword {
	case "true":
		return Token{Type: TokenBoolean, Value: true}, nil
	case "false":
		return Token{Type: TokenBoolean, Value: false}, nil
	case "null":
		return Token{Type: TokenNull, Value: nil}, nil
	case "R":
		return Token{Type: TokenRef}, nil
	default:
		return Token{Type: TokenKeyword, Value: keyword}, nil
	}
}

// readByte は1バイト読む
func (l *Lexer) readByte() (byte, error) {
	if len(l.peeked) > 0 {
		b := l.peeked[0]
		l.peeked = l.peeked[1:]
		return b, nil
	}
	return l.r.ReadByte()
}

// peekByte は次のバイトを先読みする（消費しない）
func (l *Lexer) peekByte() (byte, error) {
	if len(l.peeked) > 0 {
		return l.peeked[0], nil
	}

	b, err := l.r.ReadByte()
	if err != nil {
		return 0, err
	}

	l.peeked = append(l.peeked, b)
	return b, nil
}

// ReadBytes は指定されたバイト数を読む
func (l *Lexer) ReadBytes(n int) ([]byte, error) {
	result := make([]byte, 0, n)

	// 先読みバッファから読む
	if len(l.peeked) > 0 {
		if len(l.peeked) >= n {
			result = l.peeked[:n]
			l.peeked = l.peeked[n:]
			return result, nil
		}
		result = l.peeked
		n -= len(l.peeked)
		l.peeked = nil
	}

	// 残りをreaderから読む
	buf := make([]byte, n)
	bytesRead, err := io.ReadFull(l.r, buf)
	result = append(result, buf[:bytesRead]...)
	return result, err
}

// isDelimiter はデリミタかどうかを判定
func isDelimiter(b byte) bool {
	return b == '(' || b == ')' || b == '<' || b == '>' ||
		b == '[' || b == ']' || b == '{' || b == '}' ||
		b == '/' || b == '%'
}

// isWhitespace は空白文字かどうかを判定
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\x00' || b == '\x0c'
}

// decodeHexString は16進数文字列をデコード
func decodeHexString(hexStr string) (string, error) {
	// 奇数長の場合は末尾に0を追加
	if len(hexStr)%2 != 0 {
		hexStr += "0"
	}

	var buf bytes.Buffer
	for i := 0; i < len(hexStr); i += 2 {
		val, err := strconv.ParseInt(hexStr[i:i+2], 16, 32)
		if err != nil {
			return "", err
		}
		buf.WriteByte(byte(val))
	}

	return buf.String(), nil
}
