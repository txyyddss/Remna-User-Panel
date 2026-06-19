package workers

import (
	"context"
	"testing"
	"time"
)

func TestGroupStopsOnContextCancel(t *testing.T) {
	group := NewGroup()
	group.Add("wait", func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := group.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v, want nil on cancellation", err)
	}
}
