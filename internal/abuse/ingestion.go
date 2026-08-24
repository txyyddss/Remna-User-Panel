package abuse

import (
	"context"
	"strings"
	"time"
)

func (s *Service) Ingest(ctx context.Context, token, raw string, now time.Time) (ReportCounts, error) {
	if len(raw) == 0 || len(raw) > MaxReportBytes || strings.TrimSpace(token) == "" {
		return ReportCounts{}, ErrInvalid
	}
	credential, err := s.repo.NodeByDigest(ctx, tokenDigest(token))
	if err != nil {
		return ReportCounts{}, err
	}
	lines, err := parseReport(raw, now, MaxReportEvents)
	if err != nil {
		return ReportCounts{}, err
	}
	if err = s.repo.TouchNodeReport(ctx, credential.UUID, now.UTC()); err != nil {
		return ReportCounts{}, err
	}
	users, err := s.repo.KnownUsers(ctx, reportRemoteIDs(lines))
	if err != nil {
		return ReportCounts{}, err
	}
	events := make([]LogEvent, 0, len(lines))
	counts := ReportCounts{}
	seen := make(map[string]bool, len(lines))
	for _, line := range lines {
		if seen[line.Fingerprint] {
			counts.Duplicate++
			continue
		}
		seen[line.Fingerprint] = true
		userID, known := users[line.RemoteID]
		if !known {
			counts.Discarded++
			continue
		}
		events = append(events, LogEvent{UserID: userID, NodeUUID: credential.UUID, Domain: line.Domain, Fingerprint: line.Fingerprint, EventSecond: line.BucketAt})
	}
	return s.repo.StoreEvents(ctx, credential.UUID, events, counts, now.UTC())
}

func reportRemoteIDs(lines []parsedLine) []string {
	seen := make(map[string]bool, len(lines))
	remoteIDs := make([]string, 0, len(lines))
	for _, line := range lines {
		if line.RemoteID != "" && !seen[line.RemoteID] {
			seen[line.RemoteID] = true
			remoteIDs = append(remoteIDs, line.RemoteID)
		}
	}
	return remoteIDs
}
