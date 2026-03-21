package translate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// libreTranslateRequest はLibreTranslate APIのリクエスト
type libreTranslateRequest struct {
	Q      string `json:"q"`
	Source string `json:"source"`
	Target string `json:"target"`
	APIKey string `json:"api_key,omitempty"`
}

// libreTranslateResponse はLibreTranslate APIのレスポンス
type libreTranslateResponse struct {
	TranslatedText string `json:"translatedText"`
	Error          string `json:"error,omitempty"`
}

// LibreTranslateTranslator はLibreTranslate APIを使用した翻訳
type LibreTranslateTranslator struct {
	BaseURL    string       // APIエンドポイント (例: "http://localhost:5000")
	SourceLang string       // ソース言語コード (例: "en")
	TargetLang string       // ターゲット言語コード (例: "ja")
	APIKey     string       // APIキー（省略可）
	Client     *http.Client // HTTPクライアント（テスト用にカスタマイズ可能）
}

// NewLibreTranslateTranslator は新しいLibreTranslateTranslatorを作成
// baseURLはLibreTranslateのエンドポイント（例: "http://localhost:5000"）
func NewLibreTranslateTranslator(baseURL, sourceLang, targetLang string) *LibreTranslateTranslator {
	return &LibreTranslateTranslator{
		BaseURL:    baseURL,
		SourceLang: sourceLang,
		TargetLang: targetLang,
		Client:     http.DefaultClient,
	}
}

// Translate はLibreTranslate APIを使用してテキストを翻訳
func (t *LibreTranslateTranslator) Translate(text string) (string, error) {
	if text == "" {
		return "", nil
	}

	reqBody := libreTranslateRequest{
		Q:      text,
		Source: t.SourceLang,
		Target: t.TargetLang,
		APIKey: t.APIKey,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := t.BaseURL + "/translate"

	client := t.Client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Post(endpoint, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("LibreTranslate API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LibreTranslate API returned status %d", resp.StatusCode)
	}

	var result libreTranslateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode LibreTranslate API response: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("LibreTranslate API error: %s", result.Error)
	}

	return result.TranslatedText, nil
}
