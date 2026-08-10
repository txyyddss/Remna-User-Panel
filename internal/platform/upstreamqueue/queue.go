// Package upstreamqueue serializes synchronous calls to rate-sensitive upstream providers.
package upstreamqueue

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrNotRunning indicates that a call was submitted outside the queue lifecycle.
	ErrNotRunning = errors.New("upstream queue is not running")
	// ErrAlreadyStarted indicates that Start was called more than once.
	ErrAlreadyStarted = errors.New("upstream queue was already started")
	// ErrTaskPanic indicates that a queued task panicked. The panic value is not exposed.
	ErrTaskPanic = errors.New("upstream queue task panicked")
)

// Config controls one provider-specific worker and its bounded backlog.
type Config struct {
	Name        string
	Capacity    int
	MinInterval time.Duration
}

type lifecycleState uint8

const (
	stateNew lifecycleState = iota
	stateRunning
	stateStopping
	stateStopped
)

type call struct {
	ctx    context.Context
	run    func(context.Context)
	reject func(error)
}

// Queue is a bounded FIFO with one context-owned worker. A Queue is started once.
type Queue struct {
	name        string
	capacity    int
	minInterval time.Duration
	calls       chan call

	mu        sync.RWMutex
	state     lifecycleState
	lifecycle context.Context
	cancel    context.CancelFunc
	done      chan struct{}
}

// New validates config and creates a stopped queue.
func New(config Config) (*Queue, error) {
	if config.Name == "" {
		return nil, errors.New("upstream queue name is required")
	}
	if config.Capacity < 1 {
		return nil, errors.New("upstream queue capacity must be positive")
	}
	if config.MinInterval < 0 {
		return nil, errors.New("upstream queue minimum interval cannot be negative")
	}
	return &Queue{
		name:        config.Name,
		capacity:    config.Capacity,
		minInterval: config.MinInterval,
		calls:       make(chan call, config.Capacity),
	}, nil
}

// Name returns the provider label used in diagnostic errors.
func (q *Queue) Name() string {
	if q == nil {
		return ""
	}
	return q.name
}

// Capacity returns the maximum number of calls waiting behind the active call.
func (q *Queue) Capacity() int {
	if q == nil {
		return 0
	}
	return q.capacity
}

// MinInterval returns the configured minimum delay between call start times.
func (q *Queue) MinInterval() time.Duration {
	if q == nil {
		return 0
	}
	return q.minInterval
}

// Start attaches the worker to ctx. Cancellation stops admission and cancels the active call.
func (q *Queue) Start(ctx context.Context) error {
	if q == nil {
		return errors.New("upstream queue is nil")
	}
	if ctx == nil {
		return errors.New("upstream queue context is nil")
	}
	q.mu.Lock()
	if q.state != stateNew {
		q.mu.Unlock()
		return fmt.Errorf("start %s queue: %w", q.name, ErrAlreadyStarted)
	}
	lifecycle, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	q.lifecycle = lifecycle
	q.cancel = cancel
	q.done = done
	q.state = stateRunning
	q.mu.Unlock()

	go q.run(lifecycle, done)
	return nil
}

// Shutdown cancels the worker and waits for its active call to observe cancellation.
// It is safe to call before Start and more than once.
func (q *Queue) Shutdown(ctx context.Context) error {
	if q == nil {
		return nil
	}
	if ctx == nil {
		return errors.New("upstream queue shutdown context is nil")
	}
	q.mu.Lock()
	switch q.state {
	case stateNew:
		q.state = stateStopped
		q.mu.Unlock()
		return nil
	case stateRunning:
		q.state = stateStopping
		q.cancel()
	}
	done := q.done
	q.mu.Unlock()
	if done == nil {
		return nil
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("stop %s queue: %w", q.name, ctx.Err())
	}
}

func (q *Queue) enqueue(ctx context.Context, item call) (context.Context, error) {
	q.mu.RLock()
	if q.state != stateRunning {
		q.mu.RUnlock()
		return nil, fmt.Errorf("submit to %s queue: %w", q.name, ErrNotRunning)
	}
	lifecycle := q.lifecycle
	q.mu.RUnlock()
	if err := lifecycle.Err(); err != nil {
		return nil, fmt.Errorf("submit to %s queue: %w", q.name, err)
	}
	select {
	case q.calls <- item:
		return lifecycle, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("submit to %s queue: %w", q.name, ctx.Err())
	case <-lifecycle.Done():
		return nil, fmt.Errorf("submit to %s queue: %w", q.name, lifecycle.Err())
	}
}

func (q *Queue) run(lifecycle context.Context, done chan struct{}) {
	defer func() {
		q.mu.Lock()
		q.state = stateStopped
		q.mu.Unlock()
		close(done)
	}()
	var nextStart time.Time
	for {
		if lifecycle.Err() != nil {
			return
		}
		select {
		case <-lifecycle.Done():
			return
		case item := <-q.calls:
			q.execute(lifecycle, item, &nextStart)
		}
	}
}

func (q *Queue) execute(lifecycle context.Context, item call, nextStart *time.Time) {
	if err := item.ctx.Err(); err != nil {
		item.reject(err)
		return
	}
	if err := waitUntil(item.ctx, lifecycle, *nextStart); err != nil {
		item.reject(err)
		return
	}
	if q.minInterval > 0 {
		*nextStart = time.Now().Add(q.minInterval)
	}
	executionCtx, cancel := mergeCancellation(item.ctx, lifecycle)
	defer cancel()
	defer func() {
		if recover() != nil {
			item.reject(fmt.Errorf("%s: %w", q.name, ErrTaskPanic))
		}
	}()
	item.run(executionCtx)
}

func waitUntil(caller, lifecycle context.Context, target time.Time) error {
	delay := time.Until(target)
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-caller.Done():
		return caller.Err()
	case <-lifecycle.Done():
		return lifecycle.Err()
	}
}

func mergeCancellation(caller, lifecycle context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(caller)
	stop := context.AfterFunc(lifecycle, cancel)
	return ctx, func() {
		stop()
		cancel()
	}
}
