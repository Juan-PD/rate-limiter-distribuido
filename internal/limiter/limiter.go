package limiter

import "context"

// Limiter interface — implementations can be local or distributed
type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}
