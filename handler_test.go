package debugutils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	testDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	InitDB(testDB)
	os.Exit(m.Run())
}

func TestDebugInfo(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/api/debug/info", DebugInfo)

	req := httptest.NewRequest("GET", "/api/debug/info", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "192.168.1.1:12345"

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["method"] != "GET" {
		t.Errorf("expected method GET, got %v", resp["method"])
	}
	if resp["path"] != "/api/debug/info" {
		t.Errorf("expected path /api/debug/info, got %v", resp["path"])
	}
	if _, ok := resp["debug_id"]; !ok {
		t.Error("expected debug_id in response")
	}
	if _, ok := resp["next"]; !ok {
		t.Error("expected next in response")
	}
}

func TestSysPulse(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/api/sys/pulse", SysPulse)

	req := httptest.NewRequest("GET", "/api/sys/pulse", nil)
	req.RemoteAddr = "10.0.0.1:8080"

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "operational" {
		t.Errorf("expected status operational, got %v", resp["status"])
	}
}

func TestSysLookup(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/api/sys/lookup", SysLookup)

	req := httptest.NewRequest("GET", "/api/sys/lookup", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] == "" {
		t.Error("expected status field")
	}
	if resp["code"] == "" {
		t.Error("expected code field")
	}
}

func TestSysConsole(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/api/sys/console", SysConsole)

	tests := []struct {
		cmd      string
		expected string
	}{
		{"", "_"},
		{"help", "Available"},
		{"ls", "denied"},
		{"sudo", "BUSTER"},
		{"whoami", "anonymous"},
	}

	for _, tt := range tests {
		path := "/api/sys/console"
		if tt.cmd != "" {
			path += "?cmd=" + tt.cmd
		}

		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("cmd=%q: expected 200, got %d", tt.cmd, w.Code)
			continue
		}

		var resp map[string]string
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Errorf("cmd=%q: failed to decode: %v", tt.cmd, err)
			continue
		}

		if !strings.Contains(resp["output"], tt.expected) && tt.expected != "_" {
			t.Errorf("cmd=%q: expected output containing %q, got %q", tt.cmd, tt.expected, resp["output"])
		}
	}
}

func TestSysFetchNoHeader(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/api/sys/fetch", SysFetch)

	req := httptest.NewRequest("GET", "/api/sys/fetch", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestDebugID(t *testing.T) {
	id1 := debugID("192.168.1.1")
	id2 := debugID("192.168.1.1")
	id3 := debugID("10.0.0.1")

	if id1 != id2 {
		t.Error("same IP should produce same debug_id")
	}
	if id1 == id3 {
		t.Error("different IPs should produce different debug_id")
	}
	if len(id1) != debugIDLen {
		t.Errorf("expected debug_id length %d, got %d", debugIDLen, len(id1))
	}
}
