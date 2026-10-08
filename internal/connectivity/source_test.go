package connectivity

import (
	"context"
	"testing"
	"time"
)

type lateSource struct {
	*testSource
	entered chan struct{}
	release chan struct{}
}

func (s *lateSource) Load(context.Context, int64) (User, []Target, error) {
	close(s.entered)
	<-s.release
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.user, cloneTargets(s.targets), nil
}

func TestInvalidatedSourceGenerationCannotRepopulateSameConfig(t *testing.T) {
	settings, original, repository, cfg := configuredDependencies(t)
	source := &lateSource{testSource: original, entered: make(chan struct{}), release: make(chan struct{})}
	t.Cleanup(func() { close(source.release) })
	service := NewService(settings, source, nil, repository, nil)
	hash := ConfigHash(cfg)
	service.mu.Lock()
	service.applyConfigLocked(cfg)
	service.mu.Unlock()
	done := make(chan error, 1)
	go func() { _, _, err := service.loadTargets(context.Background(), cfg, hash, true); done <- err }()
	waitSignal(t, source.entered)
	service.Invalidate()
	service.mu.Lock()
	service.applyConfigLocked(cfg)
	service.mu.Unlock()
	source.release <- struct{}{}
	select {
	case err := <-done:
		if ErrorCode(err) != "CONNECTIVITY_INTERRUPTED" {
			t.Fatalf("obsolete response appeared current: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("obsolete source call did not finish")
	}
	service.mu.RLock()
	defer service.mu.RUnlock()
	if service.user != nil || len(service.targets) != 0 || !service.loadedAt.IsZero() {
		t.Fatal("obsolete source response repopulated the fresh metadata cache")
	}
}
