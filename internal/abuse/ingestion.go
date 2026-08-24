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
	users, whitelist, err := s.repo.KnownUsers(ctx, reportRemoteIDs(lines))
	if err != nil {
		return ReportCounts{}, err
	}
	policy, err := s.repo.Policy(ctx)
	if err != nil {
		return ReportCounts{}, err
	}
	rules, err := s.repo.DomainRules(ctx)
	if err != nil {
		return ReportCounts{}, err
	}
	matched, err := compileRules(rules)
	if err != nil {
		return ReportCounts{}, err
	}
	fingerprints := make([]string, 0, len(lines))
	samples := make([]Sample, 0, min(MaxSamplesPerReport, len(lines)))
	seen := make(map[string]bool, len(lines))
	for _, line := range lines {
		if len(samples) == MaxSamplesPerReport {
			break
		}
		if seen[line.Fingerprint] {
			continue
		}
		seen[line.Fingerprint] = true
		userID, known := users[line.RemoteID]
		if !known || whitelist[line.RemoteID] {
			continue
		}
		fingerprints = append(fingerprints, line.Fingerprint)
		if policy.GlobalEnabled && policy.GlobalLimit > 0 {
			samples = append(samples, sampleFor(userID, credential.UUID, "global", line, policy.GlobalLimit))
		}
		for _, rule := range matched {
			if len(samples) == MaxSamplesPerReport {
				break
			}
			if rule.regex.MatchString(line.Domain) {
				samples = append(samples, sampleFor(userID, credential.UUID, rule.Name, line, rule.QPSLimit))
			}
		}
	}
	return s.repo.StoreSamples(ctx, credential.UUID, fingerprints, samples, now.UTC())
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

func sampleFor(userID, nodeID, reason string, line parsedLine, limit int) Sample {
	return Sample{UserID: userID, NodeUUID: nodeID, ReasonName: reason, Fingerprint: line.Fingerprint, BucketAt: line.BucketAt, QPSLimit: limit, Count: 1}
}
