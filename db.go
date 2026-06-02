package debugutils

import (
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

type AccessLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	IP        string    `gorm:"index" json:"ip"`
	UserAgent string    `json:"user_agent"`
	Path      string    `gorm:"index" json:"path"`
	Count     int       `gorm:"default:0" json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

func InitDB(d *gorm.DB) {
	db = d
	db.AutoMigrate(&AccessLog{})
}

var pathLabels = map[string]string{
	"/a/s":                "Console: page visit",
	"/a/c":                "Carnival: page visit",
	"/a/d":                "Carnival: ARG completed",
	"/api/debug/info":     "System: debug info request",
	"/api/sys/pulse":      "System: pulse check",
	"/api/sys/lookup":     "Honeypot: lookup",
	"/api/sys/query":      "Honeypot: query (spinner trap)",
	"/api/sys/console":    "Honeypot: console",
	"/api/sys/fetch":      "Honeypot: fetch",
	"/api/sys/fetch:granted": "Honeypot: fetch (access granted)",
}

func decodePath(path string) string {
	if label, ok := pathLabels[path]; ok {
		return label
	}
	if strings.HasPrefix(path, "/a/x/") {
		return "Carnival: click #" + path[5:]
	}
	if strings.HasPrefix(path, "/a/e/") {
		parts := strings.SplitN(path[5:], ":", 2)
		if len(parts) == 2 {
			return "Carnival: wrong code attempt #" + parts[0] + " (entered: '" + parts[1] + "')"
		}
		return "Carnival: wrong code attempt #" + path[5:]
	}
	if strings.HasPrefix(path, "/a/k/") {
		return "Console: command '" + path[5:] + "'"
	}
	if strings.HasPrefix(path, "/api/sys/console?cmd=") {
		return "Honeypot: console cmd '" + strings.TrimPrefix(path, "/api/sys/console?cmd=") + "'"
	}
	return path
}

func LogAccess(ip, ua, path string) {
	decoded := decodePath(path)
	var entry AccessLog
	result := db.Where("ip = ? AND path = ?", ip, decoded).First(&entry)
	if result.Error != nil {
		entry = AccessLog{
			IP:        ip,
			UserAgent: ua,
			Path:      decoded,
			Count:     1,
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
		}
		db.Create(&entry)
	} else {
		entry.Count++
		entry.LastSeen = time.Now()
		db.Save(&entry)
	}
}

func GetAccessLogs() []AccessLog {
	var logs []AccessLog
	db.Order("last_seen DESC").Find(&logs)
	return logs
}

func ClearAccessLogs() {
	db.Exec("DELETE FROM access_logs")
}

func GetAccessLogsByIP(ip string) []AccessLog {
	var logs []AccessLog
	db.Where("ip = ?", ip).Order("first_seen ASC").Find(&logs)
	return logs
}
