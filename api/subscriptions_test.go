package handler

import "testing"

func TestStarsPlanForDays(t *testing.T) {
	tests := []struct {
		days   int
		amount int
		ok     bool
	}{
		{days: 14, amount: 100, ok: true},
		{days: 30, amount: 250, ok: true},
		{days: 180, amount: 500, ok: true},
		{days: 7, ok: false},
	}
	for _, tt := range tests {
		plan, ok := starsPlanForDays(tt.days)
		if ok != tt.ok {
			t.Fatalf("starsPlanForDays(%d) ok = %v, want %v", tt.days, ok, tt.ok)
		}
		if ok && (plan.Days != tt.days || plan.Amount != tt.amount) {
			t.Errorf("starsPlanForDays(%d) = %+v, want days=%d amount=%d", tt.days, plan, tt.days, tt.amount)
		}
	}
}

func TestParseSubscriptionPayload(t *testing.T) {
	userID, plan, ok := parseSubscriptionPayload("nerd_sub:12345:30")
	if !ok || userID != 12345 || plan.Days != 30 || plan.Amount != 250 {
		t.Fatalf("unexpected parsed payload: userID=%d plan=%+v ok=%v", userID, plan, ok)
	}
	for _, payload := range []string{"", "nerd_sub:0:14", "nerd_sub:abc:14", "nerd_sub:123:7", "other:123:14", "nerd_sub:123:14:extra"} {
		if _, _, ok := parseSubscriptionPayload(payload); ok {
			t.Errorf("parseSubscriptionPayload(%q) unexpectedly succeeded", payload)
		}
	}
}

func TestValidStarsCheckout(t *testing.T) {
	if !validStarsCheckout(12345, "XTR", 100, "nerd_sub:12345:14") {
		t.Fatal("expected matching Stars checkout to be valid")
	}
	tests := []struct {
		userID  int64
		currency string
		amount int
		payload string
	}{
		{12346, "XTR", 100, "nerd_sub:12345:14"},
		{12345, "USD", 100, "nerd_sub:12345:14"},
		{12345, "XTR", 99, "nerd_sub:12345:14"},
		{12345, "XTR", 250, "nerd_sub:12345:14"},
	}
	for _, tt := range tests {
		if validStarsCheckout(tt.userID, tt.currency, tt.amount, tt.payload) {
			t.Errorf("validStarsCheckout(%d, %q, %d, %q) unexpectedly succeeded", tt.userID, tt.currency, tt.amount, tt.payload)
		}
	}
}

func TestOwnerTelegramID(t *testing.T) {
	t.Setenv("OWNER_TELEGRAM_ID", "123456789")
	if got := ownerTelegramID(); got != 123456789 {
		t.Fatalf("ownerTelegramID() = %d, want 123456789", got)
	}
	if !isOwnerTelegramID(123456789) {
		t.Fatal("configured owner should be recognized")
	}
	if isOwnerTelegramID(987654321) {
		t.Fatal("another user must not be recognized as owner")
	}
	t.Setenv("OWNER_TELEGRAM_ID", "not-a-number")
	if got := ownerTelegramID(); got != 0 {
		t.Fatalf("invalid owner ID should be disabled, got %d", got)
	}
}
