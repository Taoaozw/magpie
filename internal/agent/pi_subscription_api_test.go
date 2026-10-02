package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/gateway"
	"github.com/yetone/magpie/internal/provider"
)

// A signed-in Claude Code account is discovered from inert credentials in
// a sandbox home; neither the real CLI nor the macOS Keychain is consulted.
func TestPiClaudeSubscriptionUsesMessagesAfterSync(t *testing.T) {
	home := syncHome(t)
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	writeFile(t, catalog.CachePath(), `{"zai":{"models":{"glm-4.6":{"id":"glm-4.6","limit":{"context":204800}}}},"anthropic":{"models":{"claude-opus-5-5":{"id":"claude-opus-5-5","name":"Claude Opus 5.5","limit":{"context":1000000}}}}}`)
	catalog.Reset()
	bin := t.TempDir()
	writeFile(t, filepath.Join(bin, "claude"), "#!/bin/sh\nexit 1\n")
	writeFile(t, filepath.Join(bin, "claude.cmd"), "@exit /b 1\r\n")
	if err := os.Chmod(filepath.Join(bin, "claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	writeFile(t, filepath.Join(home, ".claude", ".credentials.json"),
		`{"claudeAiOauth":{"accessToken":"fake-token","refreshToken":"fake-refresh","expiresAt":4102444800000,"subscriptionType":"max"}}`)
	writeFile(t, filepath.Join(home, ".claude.json"), `{"oauthAccount":{"emailAddress":"test@example.com"}}`)
	provider.ForgetAccounts()
	t.Cleanup(provider.ForgetAccounts)

	// The protocol preference belongs to Pi's catalog, not to Native(),
	// which still reports whether a request can be passed through.
	for _, id := range []string{"pi", "omo", "pencil"} {
		found := false
		for _, m := range magpieModels(id) {
			if m.ID != "claude/claude-opus-5-5" {
				continue
			}
			found = true
			if got := slices.Contains(m.APIs, string(provider.Anthropic)); got != (id != "pencil") {
				t.Fatalf("%s: preferred APIs %v", id, m.APIs)
			}
		}
		if !found {
			t.Fatalf("%s: Claude Code subscription missing", id)
		}
	}

	a := pi(home)
	if err := a.Field("model").Set("magpie/claude/claude-opus-5-5"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".pi", "agent", "models.json")
	check := func() {
		t.Helper()
		var file struct {
			Providers map[string]struct {
				Models []struct {
					ID      string         `json:"id"`
					API     string         `json:"api"`
					BaseURL string         `json:"baseUrl"`
					Compat  map[string]any `json:"compat"`
				} `json:"models"`
			} `json:"providers"`
		}
		if err := json.Unmarshal([]byte(readFile(path)), &file); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, m := range file.Providers["magpie"].Models {
			if m.ID == "claude/claude-opus-5-5" {
				found = true
				if m.API != "anthropic-messages" || m.BaseURL != gateway.URL() || m.Compat["forceAdaptiveThinking"] != true {
					t.Fatalf("Claude model: %+v", m)
				}
			}
			if m.ID == "relay/glm-4.6" || m.ID == "group/auto-claude-opus-5-5" {
				if m.API != "" || m.BaseURL != "" {
					t.Fatalf("Chat/group model changed protocol: %+v", m)
				}
			}
		}
		if !found {
			t.Fatal("Claude model missing from models.json")
		}
	}
	check()
	if err := provider.Save(provider.Provider{ID: "added", Key: "k", Chat: "http://127.0.0.1:1/v1", Models: []string{"glm-4.6"}}); err != nil {
		t.Fatal(err)
	}
	if err := a.Sync(); err != nil {
		t.Fatal(err)
	}
	check()
}
