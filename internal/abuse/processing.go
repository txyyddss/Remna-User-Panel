package abuse

import (
	"context"
	"time"
)

const processingBatchSize = 50000

func (s *Service) RecoverProcessing(ctx context.Context) error {
	return s.repo.RecoverEventClaims(ctx)
}

func (s *Service) Evaluate(ctx context.Context, now time.Time) error {
	return s.Process(ctx, now)
}

func (s *Service) Process(ctx context.Context, now time.Time) error {
	now = now.UTC()
	policy, err := s.repo.Policy(ctx)
	if err != nil {
		return err
	}
	if policy.StreakSeconds < MinStreakSeconds || policy.StreakSeconds > MaxStreakSeconds {
		return ErrInvalid
	}
	rules, err := s.repo.DomainRules(ctx)
	if err != nil {
		return err
	}
	compiled, err := compileRules(rules)
	if err != nil {
		return err
	}
	whitelist, err := s.repo.WhitelistedUsers(ctx)
	if err != nil {
		return err
	}
	for {
		claim, claimErr := s.repo.ClaimEvents(ctx, now.Add(-GracePeriod), now, processingBatchSize)
		if claimErr != nil {
			return claimErr
		}
		if len(claim.Events) == 0 && len(claim.Legacy) == 0 {
			return nil
		}
		result, evaluateErr := s.evaluateClaim(ctx, claim, policy, compiled, whitelist)
		if evaluateErr != nil {
			s.releaseClaim(claim.Token)
			return evaluateErr
		}
		if commitErr := s.repo.CommitEvaluation(ctx, claim, result, policy, now); commitErr != nil {
			s.releaseClaim(claim.Token)
			return commitErr
		}
	}
}

func (s *Service) releaseClaim(token string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.repo.ReleaseEventClaim(ctx, token)
}

func (s *Service) RestoreDue(ctx context.Context, now time.Time) error {
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
