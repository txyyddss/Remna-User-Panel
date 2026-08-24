package abuse

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"regexp"
	"strings"
	"time"
)

var (
	emailPattern  = regexp.MustCompile(`(?i)email:\s*([^\s,]+)`)
	targetPattern = regexp.MustCompile(`(?i)(?:tcp|udp):(?://)?([^\s:\[]+)(?::\d+)?`)
	directPattern = regexp.MustCompile(`(?i)(?:>>\s*direct\b|outbound:\s*direct\b)`)
	timePattern   = regexp.MustCompile(`\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}`)
)

type parsedLine struct {
	RemoteID, Domain, Fingerprint string
	BucketAt                      time.Time
}

func parseReport(raw string, fallback time.Time, limit int) ([]parsedLine, error) {
	if limit < 1 {
		return nil, ErrInvalid
	}
	lines := make([]parsedLine, 0, limit)
	scanner := bufio.NewScanner(strings.NewReader(raw))
	scanner.Buffer(make([]byte, 1024), 64<<10)
	for scanner.Scan() {
		if line, ok := parseLine(scanner.Text(), fallback); ok {
			lines = append(lines, line)
			if len(lines) == limit {
				break
			}
		}
	}
	if scanner.Err() != nil {
		return nil, ErrInvalid
	}
	return lines, nil
}

func parseLine(line string, fallback time.Time) (parsedLine, bool) {
	lower := strings.ToLower(line)
	if !directPattern.MatchString(lower) || !strings.Contains(lower, "accepted") || strings.Contains(lower, "error") {
		return parsedLine{}, false
	}
	email := emailPattern.FindStringSubmatch(line)
	target := targetPattern.FindAllStringSubmatch(line, -1)
	if len(email) != 2 || len(target) == 0 {
		return parsedLine{}, false
	}
	domain := strings.Trim(strings.TrimSpace(target[len(target)-1][1]), "[]")
	if domain == "" || net.ParseIP(domain) != nil || strings.Contains(domain, "/") {
		return parsedLine{}, false
	}
	bucket := fallback.UTC().Truncate(time.Second)
	if found := timePattern.FindString(line); found != "" {
		if parsed, err := time.ParseInLocation("2006/01/02 15:04:05", found, time.UTC); err == nil {
			bucket = parsed.UTC()
		}
	}
	fingerprint := sha256.Sum256([]byte(line))
	return parsedLine{RemoteID: strings.TrimSpace(email[1]), Domain: strings.ToLower(domain), Fingerprint: hex.EncodeToString(fingerprint[:]), BucketAt: bucket}, true
}
