package debugutils

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func RegisterRoutes(r chi.Router) {
	r.Get("/api/debug/info", DebugInfo)

	r.Route("/api/sys", func(r chi.Router) {
		r.Get("/pulse", SysPulse)
		r.Get("/fetch", SysFetch)
		r.Get("/lookup", SysLookup)
		r.Get("/query", SysQuery)
		r.Get("/console", SysConsole)
	})

	r.Get("/console", ServeConsole)
	r.Get("/kiosk/carnival", ServeCarnival)
}

func AccessLogsHandler(w http.ResponseWriter, r *http.Request) {
	logs := GetAccessLogs()
	render.JSON(w, r, logs)
}

func ClearLogsHandler(w http.ResponseWriter, r *http.Request) {
	ClearAccessLogs()
	render.JSON(w, r, map[string]bool{"cleared": true})
}

func SignalDepthHandler(w http.ResponseWriter, r *http.Request) {
	logs := GetAccessLogs()

	type visitor struct {
		IP        string   `json:"ip"`
		UserAgent string   `json:"user_agent"`
		Stages    int      `json:"signal_depth"`
		Paths     []string `json:"paths_visited"`
		LastSeen  string   `json:"last_seen"`
	}

	seen := make(map[string]*visitor)
	pathSet := make(map[string]map[string]bool)
	for _, l := range logs {
		if _, ok := seen[l.IP]; !ok {
			seen[l.IP] = &visitor{
				IP:        l.IP,
				UserAgent: l.UserAgent,
				LastSeen:  l.LastSeen.Format("2006-01-02 15:04:05"),
			}
			pathSet[l.IP] = make(map[string]bool)
		}
		pathSet[l.IP][l.Path] = true
	}

	for ip, paths := range pathSet {
		for p := range paths {
			seen[ip].Paths = append(seen[ip].Paths, p)
		}
	}

	signalPaths := map[string]bool{
		"System: debug info request":    true,
		"System: pulse check":           true,
		"Honeypot: fetch":               true,
		"Honeypot: fetch (access granted)": true,
		"Honeypot: lookup":              true,
		"Honeypot: query (spinner trap)": true,
		"Honeypot: console":             true,
		"Carnival: page visit":          true,
		"Carnival: ARG completed":       true,
		"Console: page visit":           true,
	}

	for _, l := range logs {
		if signalPaths[l.Path] || strings.HasPrefix(l.Path, "Console: command") ||
			strings.HasPrefix(l.Path, "Carnival: click") ||
			strings.HasPrefix(l.Path, "Carnival: wrong code") {
			seen[l.IP].Stages++
		}
	}

	result := make([]*visitor, 0, len(seen))
	for _, v := range seen {
		result = append(result, v)
	}

	render.JSON(w, r, result)
}
