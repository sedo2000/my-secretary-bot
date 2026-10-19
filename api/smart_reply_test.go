package handler

import "testing"

func TestIsMemberStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		isMember bool
		want     bool
	}{
		{name: "creator", status: "creator", want: true},
		{name: "administrator", status: "administrator", want: true},
		{name: "member", status: "member", want: true},
		{name: "restricted member", status: "restricted", isMember: true, want: true},
		{name: "restricted non-member", status: "restricted", isMember: false, want: false},
		{name: "left", status: "left", want: false},
		{name: "kicked", status: "kicked", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMemberStatus(tt.status, tt.isMember); got != tt.want {
				t.Fatalf("isMemberStatus(%q, %t) = %t, want %t", tt.status, tt.isMember, got, tt.want)
			}
		})
	}
}

func TestForcedSubscriptionChannelDefault(t *testing.T) {
	t.Setenv("FORCE_SUBSCRIPTION_CHANNEL", "")
	if got := forcedSubscriptionChannel(); got != "@slf00" {
		t.Fatalf("forcedSubscriptionChannel() = %q, want @slf00", got)
	}
}

func TestGenerateSmartReplyRequiresAPIKey(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "")
	if _, err := generateSmartReply("مرحبا"); err == nil {
		t.Fatal("expected missing API key to be reported")
	}
}
