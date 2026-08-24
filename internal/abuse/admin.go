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
	return s.repo.UpdatePolicy(ctx, actor, policy, now.UTC())
}
func (s *Service) Rules(ctx context.Context) ([]DomainRule, error) { return s.repo.DomainRules(ctx) }
func (s *Service) SaveRule(ctx context.Context, actor string, rule DomainRule, now time.Time) (DomainRule, error) {
	if _, err := regexp.Compile(rule.Expression); err != nil {
		return DomainRule{}, ErrInvalid
	}
	return s.repo.SaveDomainRule(ctx, actor, rule, now.UTC())
}
func (s *Service) DeleteRule(ctx context.Context, actor, id string, revision int, now time.Time) error {
	return s.repo.DeleteDomainRule(ctx, actor, id, revision, now.UTC())
}
func (s *Service) Whitelist(ctx context.Context) ([]string, error) { return s.repo.Whitelist(ctx) }
func (s *Service) SetWhitelist(ctx context.Context, remoteID string, enabled bool, now time.Time) error {
	if strings.TrimSpace(remoteID) == "" {
		return ErrInvalid
	}
	return s.repo.SetWhitelist(ctx, remoteID, enabled, now.UTC())
}
func (s *Service) Punishments(ctx context.Context) ([]PunishmentRule, error) {
	return s.repo.PunishmentRules(ctx)
}
func (s *Service) SavePunishment(ctx context.Context, actor string, rule PunishmentRule, now time.Time) (PunishmentRule, error) {
	return s.repo.SavePunishmentRule(ctx, actor, rule, now.UTC())
}
func (s *Service) Statistics(ctx context.Context, now time.Time) (map[string]float64, error) {
	return s.repo.Statistics(ctx, now.UTC())
}
func (s *Service) DeleteRecord(ctx context.Context, actor, id string, now time.Time) error {
	return s.repo.DeleteRecord(ctx, actor, id, now.UTC())
}
