package service

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"x-ui/xray"
)

// IPLimitStats represents aggregated ban/unban statistics.
type IPLimitStats struct {
	GeneratedAt  time.Time        `json:"generatedAt"`
	BanCount     int              `json:"banCount"`
	UnbanCount   int              `json:"unbanCount"`
	UniqueIPs    int              `json:"uniqueIPs"`
	UniqueEmails int              `json:"uniqueEmails"`
	TopEmails    []EmailFrequency `json:"topEmails"`
}

// EmailFrequency holds frequency of bans per email.
type EmailFrequency struct {
	Email string `json:"email"`
	Bans  int    `json:"bans"`
}

var banLineRe = regexp.MustCompile(`\bBAN\b.*\[Email\] = ([^ ]+) .* \[IP\] = ([^ ]+)`)
var unbanLineRe = regexp.MustCompile(`\bUNBAN\b.*\[Email\] = ([^ ]+) .* \[IP\] = ([^ ]+)`)

// GetIPLimitStats parses the unified IP limit log and returns aggregated statistics.
// Best-effort; if log missing or unreadable returns empty stats with timestamp.
func GetIPLimitStats(maxTop int) (*IPLimitStats, error) {
	if maxTop <= 0 {
		maxTop = 5
	}
	stats := &IPLimitStats{GeneratedAt: time.Now()}
	path := xray.GetIPLimitLogPath()
	f, err := os.Open(path)
	if err != nil {
		return stats, nil // treat as empty stats; not a fatal error
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	emailBanFreq := make(map[string]int)
	emails := make(map[string]struct{})
	ips := make(map[string]struct{})

	for scanner.Scan() {
		line := scanner.Text()
		if m := banLineRe.FindStringSubmatch(line); m != nil {
			stats.BanCount++
			email := strings.TrimSpace(m[1])
			ip := strings.TrimSpace(m[2])
			if email != "" {
				emails[email] = struct{}{}
				emailBanFreq[email]++
			}
			if ip != "" {
				ips[ip] = struct{}{}
			}
			continue
		}
		if m := unbanLineRe.FindStringSubmatch(line); m != nil {
			stats.UnbanCount++
			email := strings.TrimSpace(m[1])
			ip := strings.TrimSpace(m[2])
			if email != "" {
				emails[email] = struct{}{}
			}
			if ip != "" {
				ips[ip] = struct{}{}
			}
		}
	}

	stats.UniqueIPs = len(ips)
	stats.UniqueEmails = len(emails)

	// Build top emails slice
	top := make([]EmailFrequency, 0, len(emailBanFreq))
	for e, c := range emailBanFreq {
		top = append(top, EmailFrequency{Email: e, Bans: c})
	}
	sort.Slice(top, func(i, j int) bool {
		if top[i].Bans == top[j].Bans {
			return top[i].Email < top[j].Email
		}
		return top[i].Bans > top[j].Bans
	})
	if len(top) > maxTop {
		top = top[:maxTop]
	}
	stats.TopEmails = top
	return stats, nil
}
