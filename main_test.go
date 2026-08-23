package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecideSMS(t *testing.T) {
	tests := []struct {
		name   string
		tenant Tenant
		to     string
		want   bool
		reason string
	}{
		{"active recipient", Tenant{Active: true, Suppressed: map[string]bool{}}, "user@example.com", true, "allowed"},
		{"admin suppression", Tenant{Active: true, Suppressed: map[string]bool{"user@example.com": true}}, "user@example.com", false, "recipient_suppressed"},
		{"inactive tenant", Tenant{Active: false, Suppressed: map[string]bool{}}, "user@example.com", false, "tenant_inactive"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, reason := decide(tt.tenant, tt.to)
			if got != tt.want || reason != tt.reason {
				t.Fatalf("got %v/%s, want %v/%s", got, reason, tt.want, tt.reason)
			}
		})
	}
}

func TestHandleSendDecodesSnakeCaseRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/send", strings.NewReader(
		`{"tenant_id":"acme","to":"blocked@example.invalid","body":"Do not send"}`,
	))
	recorder := httptest.NewRecorder()

	handleSend(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var got Decision
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Sent || got.Reason != "recipient_suppressed" {
		t.Fatalf("got sent=%v reason=%q, want false/recipient_suppressed", got.Sent, got.Reason)
	}
}
