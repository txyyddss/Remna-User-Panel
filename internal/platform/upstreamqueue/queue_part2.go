package upstreamqueue

import (
	"context"
	"fmt"
	"time"
)

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

