package executor

import (
	"context"
	"time"

	"yapi.run/cli/internal/domain"
)

// Middleware wraps an Executor to add functionality.
type Middleware func(Executor) Executor

// WithTiming measures the duration of the request automatically.
func WithTiming(next Executor) Executor {
	return &timingExecutor{next: next}
}

type timingExecutor struct {
	next Executor
}

func (t *timingExecutor) Execute(ctx context.Context, req *domain.Request) (*domain.Response, error) {
	start := time.Now()
	resp, err := t.next.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	resp.Duration = time.Since(start)
	return resp, nil
}
