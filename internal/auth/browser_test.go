package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestGenerateState(t *testing.T) {
	state, err := generateState()
	if err != nil {
		t.Fatalf("generateState failed: %v", err)
	}
	if len(state) != 32 {
		t.Fatalf("unexpected state length: %d", len(state))
	}
}

func TestRedeemExchangeToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/cli-login-sessions/redeem" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"key": "sk_live_test123"})
	}))
	defer server.Close()

	key, err := redeemExchangeToken(server.URL, "exchange-token")
	if err != nil {
		t.Fatalf("redeemExchangeToken failed: %v", err)
	}
	if key != "sk_live_test123" {
		t.Fatalf("unexpected key: %s", key)
	}
}

func TestLogin_EndToEnd(t *testing.T) {
	dashboard := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/cli-login-sessions/redeem" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"key": "sk_live_test123"})
	}))
	defer dashboard.Close()

	mockOpener := func(authURL string) error {
		u, err := url.Parse(authURL)
		if err != nil {
			return err
		}
		callbackURL := u.Query().Get("callback_url")
		state := u.Query().Get("state")
		if callbackURL == "" || state == "" {
			return fmt.Errorf("missing callback_url or state")
		}

		go func() {
			resp, err := http.Get(callbackURL + "?exchange_token=token123&state=" + url.QueryEscape(state)) //nolint:noctx
			if err == nil {
				_ = resp.Body.Close()
			}
		}()
		return nil
	}

	key, err := login(dashboard.URL, mockOpener)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if key != "sk_live_test123" {
		t.Fatalf("unexpected key: %s", key)
	}
}

func TestLogin_StateMismatch(t *testing.T) {
	mockOpener := func(authURL string) error {
		u, _ := url.Parse(authURL)
		callbackURL := u.Query().Get("callback_url")
		go func() {
			resp, err := http.Get(callbackURL + "?exchange_token=token123&state=tampered") //nolint:noctx
			if err == nil {
				_ = resp.Body.Close()
			}
		}()
		return nil
	}

	_, err := login("https://sexyvoice.ai", mockOpener)
	if err == nil || !strings.Contains(err.Error(), "state mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case <-time.After(10 * time.Millisecond):
	}
}
