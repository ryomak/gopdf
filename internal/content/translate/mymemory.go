package translate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// defaultMyMemoryURL はMyMemory APIのデフォルトエンドポイント
const defaultMyMemoryURL = "https://api.mymemory.translated.net/get"

// myMemoryResponse はMyMemory APIのレスポンス
type myMemoryResponse struct {
	ResponseData struct {
		TranslatedText string `json:"translatedText"`
	} `json:"responseData"`
	ResponseStatus int `json:"responseStatus"`
}

// MyMemoryTranslator はMyMemory APIを使用した無料翻訳
type MyMemoryTranslator struct {
	SourceLang string       // ソース言語コード (例: "en")
	TargetLang string       // ターゲット言語コード (例: "ja")
	Email      string       // メールアドレス（省略可、設定すると1日50000文字まで）
	Client     *http.Client // HTTPクライアント（テスト用にカスタマイズ可能）
	BaseURL    string       // テスト用ベースURL（空の場合はデフォルトAPI URL）
}

// NewMyMemoryTranslator は新しいMyMemoryTranslatorを作成
func NewMyMemoryTranslator(sourceLang, targetLang string) *MyMemoryTranslator {
	return &MyMemoryTranslator{
		SourceLang: sourceLang,
		TargetLang: targetLang,
		Client:     http.DefaultClient,
	}
}

// Translate はMyMemory APIを使用してテキストを翻訳
func (t *MyMemoryTranslator) Translate(text string) (string, error) {
	if text == "" {
		return "", nil
	}

	baseURL := t.BaseURL
	if baseURL == "" {
		baseURL = defaultMyMemoryURL
	}

	apiURL := fmt.Sprintf(
		"%s?q=%s&langpair=%s|%s",
		baseURL,
		url.QueryEscape(text),
		url.QueryEscape(t.SourceLang),
		url.QueryEscape(t.TargetLang),
	)

	if t.Email != "" {
		apiURL += "&de=" + url.QueryEscape(t.Email)
	}

	client := t.Client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("MyMemory API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("MyMemory API returned status %d", resp.StatusCode)
	}

	var result myMemoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode MyMemory API response: %w", err)
	}

	if result.ResponseStatus != 200 {
		return "", fmt.Errorf("MyMemory API error: status %d", result.ResponseStatus)
	}

	return result.ResponseData.TranslatedText, nil
}
