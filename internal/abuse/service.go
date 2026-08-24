package abuse

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Service struct {
	repo  Repository
	nodes NodeProvider
}

func NewService(repo Repository, nodes NodeProvider) *Service {
	return &Service{repo: repo, nodes: nodes}
}

func (s *Service) Ingest(ctx context.Context, token, raw string, now time.Time) (ReportCounts, error) {
	if len(raw) == 0 || len(raw) > MaxReportBytes || strings.TrimSpace(token) == "" {
		return ReportCounts{}, ErrInvalid
	}
	digest := tokenDigest(token)
	credential, err := s.repo.NodeByDigest(ctx, digest)
	if err != nil {
		return ReportCounts{}, err
	}
	if err = s.repo.TouchNodeReport(ctx, credential.UUID, now.UTC()); err != nil {
		return ReportCounts{}, err
	}
	users, whitelist, err := s.repo.KnownUsers(ctx)
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
	fingerprints := make([]string, 0)
	samples := make([]Sample, 0)
	seen := make(map[string]bool)
	for _, line := range parseReport(raw, now) {
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
			samples = append(samples, Sample{UserID: userID, NodeUUID: credential.UUID, ReasonName: "global", Fingerprint: line.Fingerprint, BucketAt: line.BucketAt, QPSLimit: policy.GlobalLimit, Count: 1})
		}
		for _, rule := range matched {
			if rule.regex.MatchString(line.Domain) {
				samples = append(samples, Sample{UserID: userID, NodeUUID: credential.UUID, ReasonName: rule.Name, Fingerprint: line.Fingerprint, BucketAt: line.BucketAt, QPSLimit: rule.QPSLimit, Count: 1})
			}
		}
	}
	return s.repo.StoreSamples(ctx, credential.UUID, fingerprints, samples, now.UTC())
}

type compiledRule struct {
	DomainRule
	regex *regexp.Regexp
}

func compileRules(rules []DomainRule) ([]compiledRule, error) {
	out := make([]compiledRule, 0, len(rules))
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		re, err := regexp.Compile(rule.Expression)
		if err != nil {
			return nil, ErrInvalid
		}
		out = append(out, compiledRule{rule, re})
	}
	return out, nil
}
func tokenDigest(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func (s *Service) Evaluate(ctx context.Context, now time.Time) error {
	policy, err := s.repo.Policy(ctx)
	if err != nil {
		return err
	}
	buckets, err := s.repo.ReadyBuckets(ctx, now.UTC().Add(-GracePeriod))
	if err != nil {
		return err
	}
	grouped := make(map[string][]Sample)
	for _, sample := range buckets {
		key := sample.UserID + "\x00" + sample.ReasonName + "\x00" + sample.BucketAt.UTC().Format(time.RFC3339)
		grouped[key] = append(grouped[key], sample)
	}
	keys := make([]string, 0, len(grouped))
	for key := range grouped {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		entries := grouped[key]
		first := entries[0]
		last, streak, stateErr := s.repo.DetectorState(ctx, first.UserID, first.ReasonName)
		if stateErr != nil {
			return stateErr
		}
		if !last.IsZero() && !first.BucketAt.After(last) {
			continue
		}
		qps := 0
		nodes := make([]string, 0, len(entries))
		for _, entry := range entries {
			qps += entry.Count
			nodes = append(nodes, entry.NodeUUID)
		}
		if !last.IsZero() && first.BucketAt.Sub(last) != time.Second {
			streak = 0
		}
		if reachesLimit(qps, first.QPSLimit) {
			streak++
		} else {
			streak = 0
		}
		if err := s.repo.SaveDetectorState(ctx, first.UserID, first.ReasonName, first.BucketAt, streak); err != nil {
			return err
		}
		if streak > 0 && streak%StreakEvery == 0 {
			if _, err := s.repo.CreateIncident(ctx, first.UserID, first.BucketAt, qps, first.QPSLimit, []string{first.ReasonName}, unique(nodes), policy, now.UTC()); err != nil {
				return err
			}
		}
	}
	due, err := s.repo.DueTemporaryBans(ctx, now.UTC())
	if err != nil {
		return err
	}
	for _, userID := range due {
		if err := s.repo.QueueRestore(ctx, userID, now.UTC()); err != nil {
			return err
		}
	}
	return nil
}
func reachesLimit(qps, limit int) bool { return qps >= limit }
func unique(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}
