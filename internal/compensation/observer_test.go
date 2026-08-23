package compensation

import (
	"context"
	"errors"
	"testing"
	"time"
)

type observerRepository struct {
	config       Config
	observed     []Node
	observeCalls int
}

func (r *observerRepository) CompensationConfig(context.Context) (Config, error) {
	return r.config, nil
}
func (r *observerRepository) RecordCompensationObservation(_ context.Context, _ Config, nodes []Node, _ time.Time) error {
	r.observeCalls++
	r.observed = nodes
	return nil
}
func (*observerRepository) UpdateCompensationConfig(context.Context, string, ConfigUpdate, time.Time) (Config, error) {
	return Config{}, nil
}
func (*observerRepository) ListCompensationEvents(context.Context, string, string, int) (EventPage, error) {
	return EventPage{}, nil
}
func (*observerRepository) ApproveCompensationEvent(context.Context, ReviewInput, string, time.Time) (Event, error) {
	return Event{}, nil
}
func (*observerRepository) DismissCompensationEvent(context.Context, ReviewInput, string, time.Time) (Event, error) {
	return Event{}, nil
}

type observerProvider struct {
	nodes    []Node
	squads   []Squad
	nodeErr  error
	squadErr error
}

func (p observerProvider) CompensationNodes(context.Context) ([]Node, error) {
	return p.nodes, p.nodeErr
}
func (p observerProvider) CompensationSquads(context.Context) ([]Squad, error) {
	return p.squads, p.squadErr
}

func TestObserveSkipsIncompleteSnapshotsAndMapsInboundSquads(t *testing.T) {
	threshold, multiplier := 30, 10_000
	node := Node{UUID: "node", Name: "Node", ActiveInboundUUIDs: []string{"inbound-a"}}
	squads := []Squad{
		{UUID: "squad-a", Name: "A", Inbounds: []string{"inbound-a"}},
		{UUID: "squad-a", Name: "duplicate", Inbounds: []string{"inbound-a"}},
		{UUID: "squad-b", Name: "B", Inbounds: []string{"inbound-b"}},
	}
	tests := []struct {
		name      string
		provider  observerProvider
		wantCalls int
	}{
		{name: "node failure", provider: observerProvider{nodeErr: errors.New("nodes unavailable")}},
		{name: "squad failure", provider: observerProvider{nodes: []Node{node}, squadErr: errors.New("squads unavailable")}},
		{name: "complete", provider: observerProvider{nodes: []Node{node}, squads: squads}, wantCalls: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &observerRepository{config: Config{Enabled: true, ThresholdMinutes: &threshold, MultiplierBPS: &multiplier}}
			err := NewService(repository, test.provider).Observe(context.Background(), time.Now())
			if test.wantCalls == 0 && err == nil {
				t.Fatal("Observe() accepted an incomplete provider snapshot")
			}
			if repository.observeCalls != test.wantCalls {
				t.Fatalf("observation calls = %d, want %d", repository.observeCalls, test.wantCalls)
			}
			if test.wantCalls == 1 && (len(repository.observed[0].AffectedSquads) != 1 || repository.observed[0].AffectedSquads[0].UUID != "squad-a") {
				t.Fatalf("mapped squads = %+v", repository.observed[0].AffectedSquads)
			}
		})
	}
}
