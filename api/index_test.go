package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func invokeHandler(method, body, secret string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/api/index.go", strings.NewReader(body))
	if secret != "" {
		req.Header.Set("X-Telegram-Bot-Api-Secret-Token", secret)
	}
	rec := httptest.NewRecorder()
	Handler(rec, req)
	return rec
}

func TestHandlerRejectsNonPost(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	rec := invokeHandler(http.MethodGet, "", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestHandlerRequiresBotToken(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	rec := invokeHandler(http.MethodPost, `{}`, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
}

func TestHandlerRejectsWrongWebhookSecret(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "expected-secret")
	rec := invokeHandler(http.MethodPost, `{}`, "wrong-secret")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestHandlerRejectsInvalidJSON(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "")
	rec := invokeHandler(http.MethodPost, `{"update_id":`, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandlerRejectsTrailingJSON(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "")
	rec := invokeHandler(http.MethodPost, `{} {}`, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandlerIgnoresEditedBusinessMessages(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "")
	rec := invokeHandler(http.MethodPost, `{"edited_business_message":{}}`, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandlerIgnoresDeletedBusinessMessages(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "")
	rec := invokeHandler(http.MethodPost, `{"deleted_business_messages":{"business_connection_id":"x","chat":{"id":1},"message_ids":[2]}}`, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
}
func TestSplitTelegramTextRespectsRuneLimit(t *testing.T) {
	input := strings.Repeat("ع", 11)
	chunks := splitTelegramText(input, 4)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	joined := strings.Join(chunks, "")
	if joined != input {
		t.Fatalf("split/join changed input: got %q", joined)
	}
	for _, chunk := range chunks {
		if got := len([]rune(chunk)); got > 4 {
			t.Fatalf("chunk has %d runes; limit is 4", got)
		}
	}
}

func TestSplitTelegramTextEmptyInput(t *testing.T) {
	chunks := splitTelegramText("", 3900)
	if len(chunks) != 1 || chunks[0] != "" {
		t.Fatalf("expected one empty chunk, got %#v", chunks)
	}
}

func TestEscapeMarkdownV1(t *testing.T) {
	input := "a_b*c`d[e\\f"
	want := "a\\_b\\*c\\`d\\[e\\\\f"
	if got := escapeMarkdownV1(input); got != want {
		t.Fatalf("escapeMarkdownV1() = %q, want %q", got, want)
	}
}

func TestIsTelegramFormattingError(t *testing.T) {
	if !isTelegramFormattingError(errors.New("Bad Request: can't parse entities")) {
		t.Fatal("expected Markdown parsing errors to be recognized")
	}
	if isTelegramFormattingError(errors.New("network timeout")) {
		t.Fatal("network errors must not be retried as plain text")
	}
}
