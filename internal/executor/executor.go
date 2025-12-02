package executor

import (
	"context"

	"yapi.run/cli/internal/domain"
)

type Executor interface {
	Execute(ctx context.Context, req *domain.Request) (*domain.Response, error)
}
