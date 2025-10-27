package content

import (
	"bytes"
	"io"

	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/reader"
)

// Operation はコンテンツストリームのオペレーション
type Operation struct {
	Operator string        // オペレーター名（例: "Tj", "Td"）
	Operands []core.Object // オペランド
}

// StreamParser はコンテンツストリームをパースする
type StreamParser struct {
	lexer *reader.Lexer
}

// NewStreamParser は新しいStreamParserを作成する
func NewStreamParser(data []byte) *StreamParser {
	return &StreamParser{
		lexer: reader.NewLexer(bytes.NewReader(data)),
	}
}

// ParseOperations はコンテンツストリームからオペレーションを抽出する
func (p *StreamParser) ParseOperations() ([]Operation, error) {
	var operations []Operation
	var operands []core.Object

	for {
		token, err := p.lexer.NextToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// EOFトークン
		if token.Type == reader.TokenEOF {
			break
		}

		// キーワード（オペレーター）の場合
		if token.Type == reader.TokenKeyword {
			op := Operation{
				Operator: token.Value.(string),
				Operands: operands,
			}
			operations = append(operations, op)
			operands = nil
			continue
		}

		// オペランドを解析
		obj := p.tokenToObject(token)
		operands = append(operands, obj)
	}

	return operations, nil
}

// tokenToObject はトークンをcore.Objectに変換する
func (p *StreamParser) tokenToObject(token reader.Token) core.Object {
	switch token.Type {
	case reader.TokenInteger:
		return core.Integer(token.Value.(int))
	case reader.TokenReal:
		return core.Real(token.Value.(float64))
	case reader.TokenString:
		return core.String(token.Value.(string))
	case reader.TokenName:
		return core.Name(token.Value.(string))
	case reader.TokenBoolean:
		return core.Boolean(token.Value.(bool))
	case reader.TokenNull:
		return nil
	case reader.TokenArrayStart:
		// 配列をパース
		return p.parseArray()
	case reader.TokenDictStart:
		// 辞書をパース
		return p.parseDictionary()
	default:
		return nil
	}
}

// parseArray は配列をパースする
func (p *StreamParser) parseArray() core.Array {
	var arr core.Array

	for {
		token, err := p.lexer.NextToken()
		if err != nil || token.Type == reader.TokenEOF {
			break
		}

		if token.Type == reader.TokenArrayEnd {
			break
		}

		obj := p.tokenToObject(token)
		arr = append(arr, obj)
	}

	return arr
}

// parseDictionary は辞書をパースする
func (p *StreamParser) parseDictionary() core.Dictionary {
	dict := make(core.Dictionary)

	for {
		keyToken, err := p.lexer.NextToken()
		if err != nil || keyToken.Type == reader.TokenEOF {
			break
		}

		if keyToken.Type == reader.TokenDictEnd {
			break
		}

		if keyToken.Type != reader.TokenName {
			continue
		}

		key := core.Name(keyToken.Value.(string))

		valueToken, err := p.lexer.NextToken()
		if err != nil {
			break
		}

		value := p.tokenToObject(valueToken)
		dict[key] = value
	}

	return dict
}
