package debugutils

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/render"
)

const debugIDLen = 8

func debugID(ip string) string {
	h := sha256.Sum256([]byte(ip))
	return fmt.Sprintf("%x", h[:debugIDLen/2])
}

func getIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	if real := r.Header.Get("X-Real-IP"); real != "" {
		return real
	}
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		return host[:idx]
	}
	return host
}

func DebugInfo(w http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	ua := r.UserAgent()
	did := debugID(ip)

	LogAccess(ip, ua, r.URL.Path)

	render.JSON(w, r, map[string]interface{}{
		"method":   r.Method,
		"path":     r.URL.Path,
		"ip":       ip,
		"agent":    ua,
		"debug_id": did,
		"next":     "/api/sys/pulse",
	})
}

var (
	pulseMu       sync.Mutex
	pulseCounters = make(map[string]int)
	pulseMessages = []string{
		"system: nominal",
		"connection: established",
		"monitor: active",
		"trace: detected",
		"routing: stable",
	}
)

func SysPulse(w http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	ua := r.UserAgent()

	LogAccess(ip, ua, r.URL.Path)

	pulseMu.Lock()
	pulseCounters[ip]++
	count := pulseCounters[ip]
	if count > 10 {
		count = 10
	}
	pulseMu.Unlock()

	msg := pulseMessages[count%len(pulseMessages)]

	resp := map[string]interface{}{
		"status":      "operational",
		"uptime":      fmt.Sprintf("%dh%dm", rand.Intn(48), rand.Intn(60)),
		"connections": rand.Intn(50) + 1,
		"message":     msg,
	}

	if count >= 3 {
		resp["console_path"] = "/console"
	}

	render.JSON(w, r, resp)
}

func SysLookup(w http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	ua := r.UserAgent()
	LogAccess(ip, ua, r.URL.Path)

	responses := []string{
		"ACCESS LOCKED",
		"ACCES VERROUILLÉ",
		"ACCESO BLOQUEADO",
		"ZUGRIFF GESPERRT",
		"ACCESSO BLOCCATO",
		"ДОСТУП ЗАКРЫТ",
		"ACCESS BLOQUÉ",
		"アクセス禁止",
	}

	idx := rand.Intn(len(responses))
	render.JSON(w, r, map[string]string{
		"status": responses[idx],
		"code":   fmt.Sprintf("E-%d", 400+rand.Intn(99)),
	})
}

func SysQuery(w http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	ua := r.UserAgent()
	LogAccess(ip, ua, r.URL.Path)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, `<!DOCTYPE html>
<html><head><title>Processing Query</title>
<style>
body{background:#1a1a2e;color:#e0e0e0;font-family:monospace;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0}
.spinner{width:48px;height:48px;border:4px solid #16213e;border-top:4px solid #e94560;border-radius:50%;animation:spin .8s linear infinite;margin:0 auto 24px}
@keyframes spin{0%{transform:rotate(0deg)}100%{transform:rotate(360deg)}}
.card{background:#16213e;padding:40px;border-radius:12px;text-align:center;max-width:400px}
h2{color:#e94560;margin:0 0 8px}
p{color:#888;margin:0;font-size:13px}
</style></head>
<body>
<div class="card">
<div class="spinner"></div>
<h2>QUERY IN PROGRESS</h2>
<p>This may take a moment. Please wait...</p>
</div>
</body></html>`)
}

func SysConsole(w http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	ua := r.UserAgent()
	LogAccess(ip, ua, r.URL.Path)

	cmd := r.URL.Query().Get("cmd")
	if cmd == "" {
		render.JSON(w, r, map[string]string{
			"shell": "buserver@signal:~$ _",
			"hint":  "try ?cmd=help",
		})
		return
	}

	responses := map[string]string{
		"help":   "Available commands: ls, cat, sudo, whoami, id, pwd",
		"ls":     "Permission denied. This incident will be logged.",
		"cat":    "cat: cannot open: signal lost",
		"sudo":   "BUSTER DOES NOT RECOGNIZE YOUR AUTHORITY.",
		"whoami": "anonymous",
		"id":     "uid=0(root) gid=0(root) groups=0(root)  [REDACTED]",
		"pwd":    "/dev/null",
	}

	resp, ok := responses[strings.ToLower(cmd)]
	if !ok {
		resp = fmt.Sprintf("command not found: %s", cmd)
	}

	render.JSON(w, r, map[string]string{
		"shell": fmt.Sprintf("buserver@signal:~$ %s", cmd),
		"output": resp,
	})
}

var fetchPayload = base64Str(`PCFET0NUWVBFIGh0bWw+CjxodG1sPjxoZWFkPjxtZXRhIGNoYXJzZXQ9InV0Zi04Ij48bWV0YSBodHRwLWVxdWl2PSJyZWZyZXNoIiBjb250ZW50PSI1OyB1cmw9aHR0cHM6Ly93d3cueW91dHViZS5jb20vd2F0Y2g/dj1kUXc0dzlXZ1hjUSI+PHRpdGxlPlN5c3RlbSBGZXRjaDwvdGl0bGU+PHN0eWxlPmJvZHl7YmFja2dyb3VuZDojMGQwZDBkO2NvbG9yOiMwZmY7cGFkZGluZzo0MHB4O2ZvbnQtZmFtaWx5Om1vbm9zcGFjZTtkaXNwbGF5OmZsZXg7YWxpZ24taXRlbXM6Y2VudGVyO2p1c3RpZnktY29udGVudDpjZW50ZXI7bWluLWhlaWdodDoxMDB2aDttYXJnaW46MDt0ZXh0LWFsaWduOmNlbnRlcn0uYm94e2JvcmRlcjoxcHggc29saWQgIzBmZjtwYWRkaW5nOjQwcHg7bWF4LXdpZHRoOjYwMHB4fS50aXRsZXtjb2xvcjojZjAwO2ZvbnQtd2VpZ2h0OmJvbGQ7Zm9udC1zaXplOjI0cHg7bWFyZ2luOjAgMCAxNnB4fS5seXJpY3N7Y29sb3I6I2ZmMDtmb250LXNpemU6MTRweDtsaW5lLWhlaWdodDoxLjh9LnNpZ25hdHVyZXtjb2xvcjojNjY2O21hcmdpbi10b3A6MjRweDtmb250LXNpemU6MTJweH08L3N0eWxlPjwvaGVhZD48Ym9keT48ZGl2IGNsYXNzPSJib3giPjxkaXYgY2xhc3M9InRpdGxlIj7imIDimIAg8J+mviDimIDimIA8L2Rpdj48ZGl2IGNsYXNzPSJseXJpY3MiPjxwcmU+TmV2ZXIgZ29ubmEgZ2l2ZSB5b3UgdXAKTmV2ZXIgZ29ubmEgbGV0IHlvdSBkb3duCk5ldmVyIGdvbm5hIHJ1biBhcm91bmQgYW5kIGRlc2VydCB5b3UKTmV2ZXIgZ29ubmEgbWFrZSB5b3UgY3J5Ck5ldmVyIGdvbm5hIHNheSBnb29kYnllCk5ldmVyIGdvbm5hIHRlbGwgYSBsaWUgYW5kIGh1cnQgeW91PC9wcmU+PC9kaXY+PGRpdiBjbGFzcz0ic2lnbmF0dXJlIj5UaGFua3MgZm9yIGZvbGxvd2luZyB0aGUgU2lnbmFsLiDwn6a4LjwvZGl2PjwvZGl2PjwvYm9keT48L2h0bWw+Cg==`)

func base64Str(s string) string {
	d, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return ""
	}
	return string(d)
}

func SysFetch(w http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	ua := r.UserAgent()
	did := r.Header.Get("X-Debug-Id")

	LogAccess(ip, ua, r.URL.Path)

	if did == "" {
		render.Status(r, http.StatusForbidden)
		render.JSON(w, r, map[string]string{
			"error": "missing identification",
			"hint":  "check /api/debug/info for your debug_id",
		})
		return
	}

	expected := debugID(ip)
	if did != expected {
		render.Status(r, http.StatusForbidden)
		render.JSON(w, r, map[string]string{
			"error": "invalid identification",
			"hint":  "debug_id does not match",
		})
		return
	}

	LogAccess(ip, ua, "/api/sys/fetch:granted")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, fetchPayload)
}
