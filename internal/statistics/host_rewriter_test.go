package statistics

import (
	"context"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/providerops"
)

func TestReplaceFirstStandaloneDecimal(t *testing.T) {
	tests := []struct {
		input, replacement, want string
		changed                  bool
	}{
		{"HK 1.0 premium 2.0", "1.5", "HK 1.5 premium 2.0", true},
		{"HK 1.0x premium", "1.5", "HK 1.5 premium", true},
		{"prefix1.0 suffix", "2", "prefix1.0 suffix", false},
		{"version 1.0xbuild", "2", "version 1.0xbuild", false},
		{"node 10.1.2.3", "2", "node 10.1.2.3", false},
		{"no multiplier", "2", "no multiplier", false},
	}
	for _, test := range tests {
		got, changed := replaceFirstDecimal(test.input, test.replacement)
		if got != test.want || changed != test.changed {
			t.Fatalf("replaceFirstDecimal(%q) = %q,%v; want %q,%v", test.input, got, changed, test.want, test.changed)
		}
	}
}

func TestQueueHostMultiplierUpdatesAllowsZeroMultiplier(t *testing.T) {
	t.Parallel()
	provider := &statisticsProviderStub{
		nodes: []Node{{UUID: "node-1", Multiplier: 0}},
		hosts: []Host{
			{UUID: "host-1", Remark: "Edge 1.0", Nodes: []string{"node-1"}},
			{UUID: "host-many", Remark: "Edge 1.0", Nodes: []string{"node-1", "node-2"}},
		},
	}
	repository := &hostUpdateRepositoryStub{}
	err := QueueHostMultiplierUpdates(context.Background(), provider, repository, "admin-1", time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("QueueHostMultiplierUpdates(): %v", err)
	}
	if len(repository.inputs) != 1 {
		t.Fatalf("queued updates = %d, want one single-link host", len(repository.inputs))
	}
	target, err := decodeHostRemarkTarget(repository.inputs[0].SealedTarget)
	if err != nil || target.HostUUID != "host-1" || target.Remark != "Edge 0.0" {
		t.Fatalf("queued target = (%+v, %v)", target, err)
	}
}

func TestIntegerMultiplierFormattingRemainsReplaceable(t *testing.T) {
	t.Parallel()
	first, ok := formatMultiplierToken(2)
	if !ok || first != "2.0" {
		t.Fatalf("formatMultiplierToken(2) = %q,%v", first, ok)
	}
	remark, changed := replaceFirstDecimal("Edge 1.0", first)
	if !changed || remark != "Edge 2.0" {
		t.Fatalf("first rewrite = %q,%v", remark, changed)
	}
	second, _ := formatMultiplierToken(3)
	remark, changed = replaceFirstDecimal(remark, second)
	if !changed || remark != "Edge 3.0" {
		t.Fatalf("second rewrite = %q,%v", remark, changed)
	}
}

type hostUpdateRepositoryStub struct {
	inputs []providerops.CreateInput
}

func (r *hostUpdateRepositoryStub) CreateProviderOperation(_ context.Context, input providerops.CreateInput, _ time.Time) (providerops.Operation, bool, error) {
	r.inputs = append(r.inputs, input)
	return providerops.Operation{}, true, nil
}
