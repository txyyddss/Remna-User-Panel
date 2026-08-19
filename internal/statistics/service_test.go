package statistics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/rollover"
)

func TestRefreshRemoteUsesCreatedUsersAndCurrentMemberUsage(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	firstID, secondID := "remote-1", "remote-2"
	repository := &statisticsRepositoryStub{
		purchases: []model.Purchase{
			{ID: "purchase-1", UserID: "user-1", ValidFrom: now.AddDate(0, 0, -10), ValidUntil: now.AddDate(0, 0, 20)},
			{ID: "purchase-2", UserID: "user-2", ValidFrom: now.AddDate(0, 0, -40), ValidUntil: now.AddDate(0, 0, 20)},
		},
		users: map[string]model.User{
			"purchase-1": {ID: "user-1", Role: "user", RemnaUserID: &firstID},
			"purchase-2": {ID: "user-2", Role: "user", RemnaUserID: &secondID},
		},
	}
	firstUsed, secondUsed := int64(250), int64(750)
	provider := &statisticsProviderStub{
		digest: Digest{CreatedUsers: 9, ExpiredUsers: 4},
		snapshots: map[string]rollover.UsageSnapshot{
			firstID:  {LimitBytes: 1_000, Strategy: "NO_RESET", CurrentUsedBytes: &firstUsed},
			secondID: {LimitBytes: 1_000, Strategy: "NO_RESET", CurrentUsedBytes: &secondUsed},
		},
	}
	remote, err := NewService(repository, provider).refreshRemote(context.Background(), now)
	if err != nil {
		t.Fatalf("refreshRemote(): %v", err)
	}
	if remote.WeeklyUserIncrease != 9 || remote.MonthlyAverageUsageBPS != 5_000 || remote.MonthlyAverageUsage != 50 {
		t.Fatalf("remote statistics = %+v", remote)
	}
	if !provider.starts[firstID].Equal(now.AddDate(0, 0, -10)) || !provider.starts[secondID].Equal(now.AddDate(0, 0, -30)) {
		t.Fatalf("usage starts = %+v", provider.starts)
	}
}

func TestRefreshRemoteSkipsUnavailableMemberUsage(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	remoteID := "remote-1"
	sentinel := errors.New("provider unavailable")
	repository := &statisticsRepositoryStub{
		purchases: []model.Purchase{{ID: "purchase-1", UserID: "user-1", ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour)}},
		users:     map[string]model.User{"purchase-1": {ID: "user-1", Role: "user", RemnaUserID: &remoteID}},
	}
	provider := &statisticsProviderStub{usageErrors: map[string]error{remoteID: sentinel}}
	remote, err := NewService(repository, provider).refreshRemote(context.Background(), now)
	if err != nil || remote.MonthlyAverageUsageBPS != 0 {
		t.Fatalf("refreshRemote() = (%+v, %v)", remote, err)
	}
}

func TestRefreshRemoteSkipsMembersWithoutUsableProviderIdentity(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	remoteID := "remote-1"
	used := int64(250)
	repository := &statisticsRepositoryStub{
		purchases: []model.Purchase{
			{ID: "missing-identity", UserID: "user-1", ValidFrom: now.AddDate(0, 0, -10), ValidUntil: now.AddDate(0, 0, 10)},
			{ID: "missing-upstream", UserID: "user-2", ValidFrom: now.AddDate(0, 0, -10), ValidUntil: now.AddDate(0, 0, 10)},
			{ID: "usable", UserID: "user-3", ValidFrom: now.AddDate(0, 0, -10), ValidUntil: now.AddDate(0, 0, 10)},
		},
		users: map[string]model.User{
			"missing-identity": {ID: "user-1", Role: "user"},
			"missing-upstream": {ID: "user-2", Role: "user", RemnaUserID: stringPointer("missing")},
			"usable":           {ID: "user-3", Role: "user", RemnaUserID: &remoteID},
		},
	}
	provider := &statisticsProviderStub{
		snapshots:   map[string]rollover.UsageSnapshot{remoteID: {LimitBytes: 1_000, Strategy: "NO_RESET", CurrentUsedBytes: &used}},
		usageErrors: map[string]error{"missing": rollover.ErrRemoteUserMissing},
	}
	remote, err := NewService(repository, provider).refreshRemote(context.Background(), now)
	if err != nil || remote.MonthlyAverageUsageBPS != 2_500 {
		t.Fatalf("refreshRemote() = (%+v, %v)", remote, err)
	}
}

func stringPointer(value string) *string { return &value }

type statisticsRepositoryStub struct {
	purchases []model.Purchase
	users     map[string]model.User
}

func (r *statisticsRepositoryStub) ProductDatabaseStatistics(context.Context, time.Time) (model.DatabaseStatistics, error) {
	return model.DatabaseStatistics{}, nil
}
func (r *statisticsRepositoryStub) ActiveMemberPurchasesForStatistics(context.Context, time.Time) ([]model.Purchase, error) {
	return append([]model.Purchase(nil), r.purchases...), nil
}
func (r *statisticsRepositoryStub) UserForPurchase(_ context.Context, id string) (model.User, error) {
	return r.users[id], nil
}
func (r *statisticsRepositoryStub) LoadStatisticsPartition(context.Context, string) ([]byte, time.Time, error) {
	return nil, time.Time{}, ErrPartitionNotFound
}
func (r *statisticsRepositoryStub) SaveStatisticsPartition(context.Context, string, []byte, time.Time) error {
	return nil
}

type statisticsProviderStub struct {
	digest      Digest
	snapshots   map[string]rollover.UsageSnapshot
	usageErrors map[string]error
	starts      map[string]time.Time
	nodes       []Node
	hosts       []Host
}

func (p *statisticsProviderStub) Digest(context.Context, time.Time, time.Time) (Digest, error) {
	return p.digest, nil
}
func (p *statisticsProviderStub) Traffic(context.Context, time.Time, time.Time) (Traffic, error) {
	return Traffic{}, nil
}
func (p *statisticsProviderStub) UsageSnapshotForRollover(_ context.Context, id string, start, _ time.Time) (rollover.UsageSnapshot, error) {
	if p.starts == nil {
		p.starts = make(map[string]time.Time)
	}
	p.starts[id] = start
	if err := p.usageErrors[id]; err != nil {
		return rollover.UsageSnapshot{}, err
	}
	return p.snapshots[id], nil
}
func (p *statisticsProviderStub) Nodes(context.Context) ([]Node, error) {
	return append([]Node(nil), p.nodes...), nil
}
func (p *statisticsProviderStub) RequestNodeGeocheck(context.Context, string) (string, error) {
	return "", errors.New("geocheck is not configured")
}
func (p *statisticsProviderStub) NodeGeocheckResult(context.Context, string) (GeocheckResult, error) {
	return GeocheckResult{}, errors.New("geocheck is not configured")
}
func (p *statisticsProviderStub) Hosts(context.Context) ([]Host, error) {
	return append([]Host(nil), p.hosts...), nil
}
func (p *statisticsProviderStub) UpdateHostRemark(context.Context, string, string) error { return nil }
