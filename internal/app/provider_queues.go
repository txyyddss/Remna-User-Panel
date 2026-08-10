package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/platform/upstreamqueue"
)

const (
	remnawaveQueueCapacity = 64
	embyQueueCapacity      = 32
	remnawavePace          = 50 * time.Millisecond
	embyPace               = 100 * time.Millisecond
)

type providerQueues struct {
	remnawave *upstreamqueue.Queue
	emby      *upstreamqueue.Queue
}

func newProviderQueues() (*providerQueues, error) {
	remnawaveQueue, err := upstreamqueue.New(upstreamqueue.Config{
		Name: "remnawave", Capacity: remnawaveQueueCapacity, MinInterval: remnawavePace,
	})
	if err != nil {
		return nil, fmt.Errorf("create Remnawave queue: %w", err)
	}
	embyQueue, err := upstreamqueue.New(upstreamqueue.Config{
		Name: "emby", Capacity: embyQueueCapacity, MinInterval: embyPace,
	})
	if err != nil {
		return nil, fmt.Errorf("create Emby queue: %w", err)
	}
	return &providerQueues{remnawave: remnawaveQueue, emby: embyQueue}, nil
}

func (q *providerQueues) start(ctx context.Context) error {
	if q == nil || q.remnawave == nil || q.emby == nil {
		return errors.New("provider queues are incomplete")
	}
	if err := q.remnawave.Start(ctx); err != nil {
		return fmt.Errorf("start Remnawave queue: %w", err)
	}
	if err := q.emby.Start(ctx); err != nil {
		_ = q.remnawave.Shutdown(context.Background())
		return fmt.Errorf("start Emby queue: %w", err)
	}
	return nil
}

func (q *providerQueues) shutdown(ctx context.Context) error {
	if q == nil {
		return nil
	}
	errorsByProvider := make(chan error, 2)
	go func() { errorsByProvider <- q.remnawave.Shutdown(ctx) }()
	go func() { errorsByProvider <- q.emby.Shutdown(ctx) }()
	return errors.Join(<-errorsByProvider, <-errorsByProvider)
}
