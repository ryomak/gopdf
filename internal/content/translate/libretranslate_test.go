package translate

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLibreTranslateTranslator(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		sourceLang string
		targetLang string
		apiKey     string
		response   libreTranslateResponse
		statusCode int
		wantText   string
		wantErr    bool
	}{
		{
			name:       "successful translation",
			text:       "Hello",
			sourceLang: "en",
			targetLang: "ja",
			response:   libreTranslateResponse{TranslatedText: "こんにちは"},
			statusCode: http.StatusOK,
			wantText:   "こんにちは",
		},
		{
			name:       "empty text returns empty",
			text:       "",
			sourceLang: "en",
			targetLang: "ja",
			wantText:   "",
		},
		{
			name:       "with API key",
			text:       "Hello",
			sourceLang: "en",
			targetLang: "ja",
			apiKey:     "test-key",
			response:   libreTranslateResponse{TranslatedText: "こんにちは"},
			statusCode: http.StatusOK,
			wantText:   "こんにちは",
		},
		{
			name:       "API returns error message",
			text:       "Hello",
			sourceLang: "en",
			targetLang: "ja",
			response:   libreTranslateResponse{Error: "rate limit exceeded"},
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "HTTP error status",
			text:       "Hello",
			sourceLang: "en",
			targetLang: "ja",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.text == "" {
				translator := NewLibreTranslateTranslator("http://localhost:5000", tt.sourceLang, tt.targetLang)
				got, err := translator.Translate(tt.text)
				if (err != nil) != tt.wantErr {
					t.Errorf("Translate() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.wantText {
					t.Errorf("Translate() = %q, want %q", got, tt.wantText)
				}
				return
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/translate" {
					t.Errorf("expected /translate, got %s", r.URL.Path)
				}

				var req libreTranslateRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("failed to decode request: %v", err)
				}
				if req.Q != tt.text {
					t.Errorf("q = %q, want %q", req.Q, tt.text)
				}
				if req.Source != tt.sourceLang {
					t.Errorf("source = %q, want %q", req.Source, tt.sourceLang)
				}
				if req.Target != tt.targetLang {
					t.Errorf("target = %q, want %q", req.Target, tt.targetLang)
				}
				if tt.apiKey != "" && req.APIKey != tt.apiKey {
					t.Errorf("api_key = %q, want %q", req.APIKey, tt.apiKey)
				}

				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			translator := &LibreTranslateTranslator{
				BaseURL:    server.URL,
				SourceLang: tt.sourceLang,
				TargetLang: tt.targetLang,
				APIKey:     tt.apiKey,
				Client:     server.Client(),
			}

			got, err := translator.Translate(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("Translate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantText {
				t.Errorf("Translate() = %q, want %q", got, tt.wantText)
			}
		})
	}
}

func TestNewLibreTranslateTranslator(t *testing.T) {
	translator := NewLibreTranslateTranslator("http://localhost:5000", "en", "ja")
	if translator.BaseURL != "http://localhost:5000" {
		t.Errorf("BaseURL = %q, want %q", translator.BaseURL, "http://localhost:5000")
	}
	if translator.SourceLang != "en" {
		t.Errorf("SourceLang = %q, want %q", translator.SourceLang, "en")
	}
	if translator.TargetLang != "ja" {
		t.Errorf("TargetLang = %q, want %q", translator.TargetLang, "ja")
	}
}

func TestLibreTranslateTranslatorImplementsInterface(t *testing.T) {
	var _ Translator = &LibreTranslateTranslator{}
}
