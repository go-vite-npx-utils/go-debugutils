package debugutils

import (
	_ "embed"
	"net/http"
)

//go:embed pages/console.html
var consoleHTML string

//go:embed pages/carnival.html
var carnivalHTML string

func ServeConsole(w http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	ua := r.UserAgent()
	LogAccess(ip, ua, "/a/s")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(consoleHTML))
}

func ServeCarnival(w http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	ua := r.UserAgent()
	LogAccess(ip, ua, "/a/c")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(carnivalHTML))
}
