package abuse

import (
	"context"
	"regexp"
	"strings"
	"time"
)

func (s *Service) MemberRecords(ctx context.Context, userID, cursor string, limit int) (RecordPage, error) {
	return s.repo.MemberRecords(ctx, userID, cursor, limit)
}

func (s *Service) Policy(ctx context.Context) (Policy, error) { return s.repo.Policy(ctx) }

func (s *Service) UpdatePolicy(ctx context.Context, actor string, policy Policy, now time.Time) (Policy, error) {
	if policy.GlobalLimit < 0 || policy.GlobalLimit > MaxGlobalQPS || policy.WarningValidityDays < 1 || policy.WarningValidityDays > MaxWarningValidityDays || policy.WarningCooldownMinutes < 0 || policy.WarningCooldownMinutes > MaxWarningCooldownMinutes || policy.Revision < 0 {
		return Policy{}, ErrInvalid
	}
	return s.repo.UpdatePolicy(ctx, actor, policy, now.UTC())
}

func (s *Service) Rules(ctx context.Context) ([]DomainRule, error) { return s.repo.DomainRules(ctx) }

func (s *Service) SaveRule(ctx context.Context, actor string, rule DomainRule, now time.Time) (DomainRule, error) {
	rule.Name = strings.TrimSpace(rule.Name)
	rule.Expression = strings.TrimSpace(rule.Expression)
	if rule.Name == "" || len(rule.Name) > MaxRuleNameLength || rule.Expression == "" || len(rule.Expression) > MaxRuleTextLength || rule.QPSLimit < 1 || rule.QPSLimit > MaxGlobalQPS || rule.Revision < 0 {
		return DomainRule{}, ErrInvalid
	}
	if _, err := regexp.Compile(rule.Expression); err != nil {
		return DomainRule{}, ErrInvalid
	}
	return s.repo.SaveDomainRule(ctx, actor, rule, now.UTC())
}

func (s *Service) DeleteRule(ctx context.Context, actor, id string, revision int, now time.Time) error {
	if strings.TrimSpace(id) == "" || revision < 0 {
		return ErrInvalid
	}
	return s.repo.DeleteDomainRule(ctx, actor, id, revision, now.UTC())
}

func (s *Service) Whitelist(ctx context.Context) ([]string, error) { return s.repo.Whitelist(ctx) }

func (s *Service) SetWhitelist(ctx context.Context, remoteID string, enabled bool, now time.Time) error {
	remoteID = strings.TrimSpace(remoteID)
	if remoteID == "" || len(remoteID) > MaxRemoteIDLength || strings.ContainsRune(remoteID, '\x00') {
		return ErrInvalid
	}
	return s.repo.SetWhitelist(ctx, remoteID, enabled, now.UTC())
}

func (s *Service) Punishments(ctx context.Context) ([]PunishmentRule, error) {
	return s.repo.PunishmentRules(ctx)
}

func (s *Service) SavePunishment(ctx context.Context, actor string, rule PunishmentRule, now time.Time) (PunishmentRule, error) {
	if !rule.Action.Valid() || rule.IncidentThreshold < 1 || rule.IncidentThreshold > MaxGlobalQPS || rule.DurationMinutes < 1 || rule.DurationMinutes > MaxWarningCooldownMinutes || rule.Revision < 0 {
		return PunishmentRule{}, ErrInvalid
	}
	return s.repo.SavePunishmentRule(ctx, actor, rule, now.UTC())
}

func (s *Service) Statistics(ctx context.Context, now time.Time) (map[string]float64, error) {
	return s.repo.Statistics(ctx, now.UTC())
}

func (s *Service) DeleteRecord(ctx context.Context, actor, id string, now time.Time) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalid
	}
	return s.repo.DeleteRecord(ctx, actor, id, now.UTC())
}
