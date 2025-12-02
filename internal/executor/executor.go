package executor

import (
	"context"

	"github.com/j-pond/yapi/internal/domain"
)

type Executor interface {
	Execute(ctx context.Context, req *domain.Request) (*domain.Response, error)
}
