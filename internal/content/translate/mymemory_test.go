package translate

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMyMemoryTranslator(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		sourceLang string
		targetLang string
		email      string
		response   myMemoryResponse
		statusCode int
		wantText   string
		wantErr    bool
	}{
		{
			name:       "successful translation",
			text:       "Hello",
			sourceLang: "en",
			targetLang: "ja",
			response: myMemoryResponse{
				ResponseData: struct {
					TranslatedText string `json:"translatedText"`
				}{TranslatedText: "こんにちは"},
				ResponseStatus: 200,
			},
			statusCode: http.StatusOK,
			wantText:   "こんにちは",
			wantErr:    false,
		},
		{
			name:       "empty text returns empty",
			text:       "",
			sourceLang: "en",
			targetLang: "ja",
			wantText:   "",
			wantErr:    false,
		},
		{
			name:       "with email",
			text:       "Hello",
			sourceLang: "en",
			targetLang: "ja",
			email:      "test@example.com",
			response: myMemoryResponse{
				ResponseData: struct {
					TranslatedText string `json:"translatedText"`
				}{TranslatedText: "こんにちは"},
				ResponseStatus: 200,
			},
			statusCode: http.StatusOK,
			wantText:   "こんにちは",
			wantErr:    false,
		},
		{
			name:       "API error status",
			text:       "Hello",
			sourceLang: "en",
			targetLang: "ja",
			response: myMemoryResponse{
				ResponseStatus: 403,
			},
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "HTTP error",
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
				// Empty text should return immediately without API call
				translator := NewMyMemoryTranslator(tt.sourceLang, tt.targetLang)
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
				// Verify query parameters
				q := r.URL.Query()
				if q.Get("q") == "" {
					t.Error("missing q parameter")
				}
				if q.Get("langpair") == "" {
					t.Error("missing langpair parameter")
				}
				if tt.email != "" {
					if q.Get("de") != tt.email {
						t.Errorf("de parameter = %q, want %q", q.Get("de"), tt.email)
					}
				}

				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			translator := &MyMemoryTranslator{
				SourceLang: tt.sourceLang,
				TargetLang: tt.targetLang,
				Email:      tt.email,
				Client:     server.Client(),
				BaseURL:    server.URL,
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

func TestNewMyMemoryTranslator(t *testing.T) {
	translator := NewMyMemoryTranslator("en", "ja")
	if translator.SourceLang != "en" {
		t.Errorf("SourceLang = %q, want %q", translator.SourceLang, "en")
	}
	if translator.TargetLang != "ja" {
		t.Errorf("TargetLang = %q, want %q", translator.TargetLang, "ja")
	}
	if translator.Client == nil {
		t.Error("Client should not be nil")
	}
}

func TestMyMemoryTranslatorImplementsInterface(t *testing.T) {
	var _ Translator = &MyMemoryTranslator{}
}
