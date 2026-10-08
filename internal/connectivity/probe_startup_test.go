package connectivity

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/xtls/xray-core/core"
)

type probeStartupFeature struct {
	started atomic.Int64
	closed  atomic.Int64
	failure error
}

func (*probeStartupFeature) Type() any { return (*probeStartupFeature)(nil) }
func (feature *probeStartupFeature) Start() error {
	feature.started.Add(1)
	return feature.failure
}
func (feature *probeStartupFeature) Close() error {
	feature.closed.Add(1)
	return nil
}

func probeFixtureConstructor(feature *probeStartupFeature) func(context.Context, *core.Config) (*core.Instance, error) {
	return func(ctx context.Context, config *core.Config) (*core.Instance, error) {
		instance, err := core.NewWithContext(ctx, config)
		if err != nil {
			return nil, err
		}
		if err := instance.AddFeature(feature); err != nil {
			// Return the owned partial instance so the probe closes it on failure.
			return instance, err
		}
		return instance, nil
	}
}

func TestXrayProbeFailedStartupClosesConstructedEngine(t *testing.T) {
	feature := &probeStartupFeature{failure: errors.New("fixture-sensitive-startup-value")}
	probe := &XrayProbe{newInstance: probeFixtureConstructor(feature)}
	outcome := probe.Check(context.Background(), probeResolvedTarget(t, "127.0.0.1:443", probeUserUUID), Config{ProbeURL: "https://example.com", TimeoutSeconds: 1})
	if outcome.Status != "error" || outcome.ErrorCode != "CONNECTIVITY_XRAY_START_FAILED" {
		t.Fatalf("startup failure was not sanitized: %+v", outcome)
	}
	if feature.started.Load() != 1 || feature.closed.Load() != 1 {
		t.Fatalf("startup lifecycle: starts=%d closes=%d", feature.started.Load(), feature.closed.Load())
	}
}

func TestXrayProbeCancellationAfterConstructionClosesBeforeStarting(t *testing.T) {
	feature := new(probeStartupFeature)
	constructor := probeFixtureConstructor(feature)
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	probe := &XrayProbe{newInstance: func(ctx context.Context, config *core.Config) (*core.Instance, error) {
		if _, exists := ctx.Deadline(); !exists {
			return nil, errors.New("probe constructor lost its deadline")
		}
		instance, err := constructor(ctx, config)
		cancel()
		return instance, err
	}}
	outcome := probe.Check(parent, probeResolvedTarget(t, "127.0.0.1:443", probeUserUUID), Config{ProbeURL: "https://example.com", TimeoutSeconds: 1})
	if outcome.Status != "interrupted" || outcome.ErrorCode != "CONNECTIVITY_INTERRUPTED" || feature.started.Load() != 0 || feature.closed.Load() != 1 {
		t.Fatalf("cancelled startup lifecycle: outcome=%+v starts=%d closes=%d", outcome, feature.started.Load(), feature.closed.Load())
	}
}

func TestXrayProbeInvalidConstructorResultsDoNotPanicOrLeak(t *testing.T) {
	for _, partial := range []bool{false, true} {
		name := "nil instance"
		if partial {
			name = "partial instance with error"
		}
		t.Run(name, func(t *testing.T) {
			feature := new(probeStartupFeature)
			probe := &XrayProbe{newInstance: func(ctx context.Context, config *core.Config) (*core.Instance, error) {
				if !partial {
					return nil, nil
				}
				instance, err := probeFixtureConstructor(feature)(ctx, config)
				if err != nil {
					return instance, err
				}
				return instance, errors.New("fixture-sensitive-constructor-value")
			}}
			outcome := probe.Check(context.Background(), probeResolvedTarget(t, "127.0.0.1:443", probeUserUUID), Config{ProbeURL: "https://example.com", TimeoutSeconds: 1})
			if outcome.Status != "error" || outcome.ErrorCode != "CONNECTIVITY_XRAY_START_FAILED" || feature.started.Load() != 0 {
				t.Fatalf("constructor failure = %+v", outcome)
			}
			if partial && feature.closed.Load() != 1 {
				t.Fatal("partial engine was not closed")
			}
		})
	}
}
