package handler

import (
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
