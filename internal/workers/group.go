// Package workers provides cancellable worker orchestration.
package workers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Task is a long-running worker function.
type Task func(context.Context) error

// Group runs worker tasks under a shared context.
type Group struct {
	tasks map[string]Task
}

// NewGroup creates an empty worker group.
func NewGroup() *Group {
	return &Group{tasks: map[string]Task{}}
}

// Add registers a named task.
func (g *Group) Add(name string, task Task) {
	g.tasks[name] = task
}

// Run starts all tasks and returns when the context is cancelled or a task fails.
func (g *Group) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, len(g.tasks))
	var wg sync.WaitGroup
	for name, task := range g.tasks {
		name, task := name, task
		wg.Add(1)
		go func() {
			defer wg.Done()
			slog.Info("worker started", "worker", name)
			if err := task(ctx); err != nil && !errors.Is(err, context.Canceled) {
				errCh <- fmt.Errorf("%s: %w", name, err)
			}
			slog.Info("worker stopped", "worker", name)
		}()
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		cancel()
		<-done
		return nil
	case err := <-errCh:
		cancel()
		<-done
		return err
	}
}

// Interval returns a task that executes fn immediately and then on interval.
func Interval(interval time.Duration, fn func(context.Context) error) Task {
	if interval <= 0 {
		interval = time.Minute
	}
	return func(ctx context.Context) error {
		if err := fn(ctx); err != nil {
			return err
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				if err := fn(ctx); err != nil {
					return err
				}
			}
		}
	}
}
