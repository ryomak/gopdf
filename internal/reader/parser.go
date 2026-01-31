package reader

import (
	"fmt"
	"io"

	"github.com/ryomak/gopdf/internal/core"
)

// Parser はPDFオブジェクトをパースする
type Parser struct {
	lexer  *Lexer
	peeked []Token // 先読みトークンのバッファ
}

// NewParser は新しいParserを作成する
func NewParser(r io.Reader) *Parser {
	return &Parser{
		lexer: NewLexer(r),
	}
}

// ParseObject は次のオブジェクトをパースする
func (p *Parser) ParseObject() (core.Object, error) {
	token, err := p.nextToken()
	if err != nil {
		return nil, err
	}

	switch token.Type {
	case TokenInteger:
		// 整数、または参照の可能性
		return p.parseIntegerOrReference(token.Value.(int))

	case TokenReal:
		return core.Real(token.Value.(float64)), nil

	case TokenString:
		return core.String(token.Value.(string)), nil

	case TokenName:
		return core.Name(token.Value.(string)), nil

	case TokenBoolean:
		return core.Boolean(token.Value.(bool)), nil

	case TokenNull:
		return nil, nil

	case TokenDictStart:
		return p.ParseDictionary()

	case TokenArrayStart:
		return p.ParseArray()

	default:
		return nil, fmt.Errorf("unexpected token type: %v", token.Type)
	}
}

// parseIntegerOrReference は整数または参照をパースする
func (p *Parser) parseIntegerOrReference(firstNum int) (core.Object, error) {
	// 次のトークンを先読み
	token2, err := p.peekToken()
	if err != nil {
		// 次がない場合は整数
		return core.Integer(firstNum), nil
	}

	if token2.Type != TokenInteger {
		// 次が整数でない場合は整数
		return core.Integer(firstNum), nil
	}

	// 2番目の数値を消費
	_, _ = p.nextToken() // エラーは既にpeekTokenで検出済み
	secondNum := token2.Value.(int)

	// 3番目のトークンを先読み
	token3, err := p.peekToken()
	if err != nil {
		// 次がない場合は2つの整数を戻す（エラー）
		return nil, fmt.Errorf("unexpected end after two integers")
	}

	if token3.Type == TokenRef {
		// 参照: N M R
		_, _ = p.nextToken() // R を消費（エラーは既にpeekTokenで検出済み）
		return &core.Reference{
			ObjectNumber:     firstNum,
			GenerationNumber: secondNum,
		}, nil
	}

	// 参照でない場合は最初の整数を返し、2番目を戻す
	p.unreadToken(token2)
	return core.Integer(firstNum), nil
}

// ParseDictionary は辞書をパースする
// 呼び出し前に << は既に消費されている
func (p *Parser) ParseDictionary() (core.Dictionary, error) {
	dict := make(core.Dictionary)

	for {
		token, err := p.peekToken()
		if err != nil {
			return nil, err
		}

		if token.Type == TokenDictEnd {
			_, _ = p.nextToken() // >> を消費（エラーは既にpeekTokenで検出済み）
			break
		}

		// キーを読む（Name型）
		keyToken, err := p.nextToken()
		if err != nil {
			return nil, err
		}

		if keyToken.Type != TokenName {
			return nil, fmt.Errorf("expected name for dictionary key, got %v", keyToken.Type)
		}

		key := core.Name(keyToken.Value.(string))

		// 値を読む
		value, err := p.ParseObject()
		if err != nil {
			return nil, err
		}

		dict[key] = value
	}

	return dict, nil
}

// ParseArray は配列をパースする
// 呼び出し前に [ は既に消費されている
func (p *Parser) ParseArray() (core.Array, error) {
	var arr core.Array

	for {
		token, err := p.peekToken()
		if err != nil {
			return nil, err
		}

		if token.Type == TokenArrayEnd {
			_, _ = p.nextToken() // ] を消費（エラーは既にpeekTokenで検出済み）
			break
		}

		// 要素を読む
		obj, err := p.ParseObject()
		if err != nil {
			return nil, err
		}

		arr = append(arr, obj)
	}

	return arr, nil
}

// ParseStream はストリームをパースする
// dict: すでにパースされた辞書
// "stream" キーワードの後から読み込む
func (p *Parser) ParseStream(dict core.Dictionary) (*core.Stream, error) {
	// "stream" キーワードを確認
	token, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if token.Type != TokenKeyword || token.Value.(string) != "stream" {
		return nil, fmt.Errorf("expected 'stream' keyword, got %v", token)
	}

	// stream の後の改行をスキップ（\r\n または \n）
	// \r\n または \n のいずれかを読み飛ばす
	firstByte, err := p.lexer.ReadBytes(1)
	if err != nil {
		return nil, fmt.Errorf("failed to read newline after stream: %w", err)
	}
	if firstByte[0] == '\r' {
		// \r の後に \n が続く可能性がある
		_, _ = p.lexer.ReadBytes(1) // \nを読む（エラーは無視）
	}
	// \r\n, \n, その他いずれの場合も、改行として処理を続行
	// TODO: より厳密には、改行でない文字をunreadする必要がある

	// Lengthを取得
	lengthObj, ok := dict[core.Name("Length")]
	if !ok {
		return nil, fmt.Errorf("stream dictionary must have /Length")
	}

	var length int
	switch v := lengthObj.(type) {
	case core.Integer:
		length = int(v)
	case *core.Reference:
		// 参照の場合は解決が必要（ここでは未対応、エラーにする）
		return nil, fmt.Errorf("stream length reference not yet supported")
	default:
		return nil, fmt.Errorf("invalid stream length type: %T", lengthObj)
	}

	// ストリームデータを読む
	data, err := p.lexer.ReadBytes(length)
	if err != nil {
		return nil, fmt.Errorf("failed to read stream data: %w", err)
	}

	// "endstream" キーワードを消費
	// ストリームデータの後の改行を飛ばす可能性がある
	endstreamToken, err := p.nextToken()
	if err != nil {
		return nil, fmt.Errorf("failed to read endstream: %w", err)
	}
	if endstreamToken.Type != TokenKeyword || endstreamToken.Value.(string) != "endstream" {
		return nil, fmt.Errorf("expected 'endstream' keyword, got %v", endstreamToken)
	}

	return &core.Stream{
		Dict: dict,
		Data: data,
	}, nil
}

// ParseIndirectObject は間接オブジェクトをパースする
// 形式: N M obj <object> endobj
// 戻り値: オブジェクト番号, 世代番号, オブジェクト, エラー
func (p *Parser) ParseIndirectObject() (int, int, core.Object, error) {
	// オブジェクト番号
	token1, err := p.nextToken()
	if err != nil {
		return 0, 0, nil, err
	}
	if token1.Type != TokenInteger {
		return 0, 0, nil, fmt.Errorf("expected object number, got %v", token1.Type)
	}
	objNum := token1.Value.(int)

	// 世代番号
	token2, err := p.nextToken()
	if err != nil {
		return 0, 0, nil, err
	}
	if token2.Type != TokenInteger {
		return 0, 0, nil, fmt.Errorf("expected generation number, got %v", token2.Type)
	}
	genNum := token2.Value.(int)

	// "obj" キーワード
	token3, err := p.nextToken()
	if err != nil {
		return 0, 0, nil, err
	}
	if token3.Type != TokenKeyword || token3.Value.(string) != "obj" {
		return 0, 0, nil, fmt.Errorf("expected 'obj' keyword, got %v", token3)
	}

	// オブジェクト本体
	obj, err := p.ParseObject()
	if err != nil {
		return 0, 0, nil, err
	}

	// Dictionaryの後に "stream" がある可能性をチェック
	if dict, ok := obj.(core.Dictionary); ok {
		nextToken, err := p.peekToken()
		if err == nil && nextToken.Type == TokenKeyword && nextToken.Value.(string) == "stream" {
			// Streamをパース
			stream, err := p.ParseStream(dict)
			if err != nil {
				return 0, 0, nil, err
			}
			obj = stream
		}
	}

	// "endobj" キーワード
	token4, err := p.nextToken()
	if err != nil {
		return 0, 0, nil, err
	}
	if token4.Type != TokenKeyword || token4.Value.(string) != "endobj" {
		return 0, 0, nil, fmt.Errorf("expected 'endobj' keyword, got %v", token4)
	}

	return objNum, genNum, obj, nil
}

// nextToken は次のトークンを返す
func (p *Parser) nextToken() (Token, error) {
	if len(p.peeked) > 0 {
		token := p.peeked[0]
		p.peeked = p.peeked[1:]
		return token, nil
	}
	return p.lexer.NextToken()
}

// peekToken は次のトークンを先読みする（消費しない）
func (p *Parser) peekToken() (Token, error) {
	if len(p.peeked) > 0 {
		return p.peeked[0], nil
	}

	token, err := p.lexer.NextToken()
	if err != nil {
		return Token{}, err
	}

	p.peeked = append(p.peeked, token)
	return token, nil
}

// unreadToken はトークンを戻す
func (p *Parser) unreadToken(token Token) {
	p.peeked = append([]Token{token}, p.peeked...)
}
