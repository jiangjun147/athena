package limiter

import (
	"fmt"
	"sync"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/rickone/athena/errcode"
	"github.com/rickone/athena/redis"
	"golang.org/x/time/rate"
	"google.golang.org/grpc/status"
)

type LimiterManager struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	option   LimiterManagerOption
}

type LimiterManagerOption struct {
	Limit  float64
	Bucket int
}

func NewLimiterManager(l float64, b int) *LimiterManager {
	return &LimiterManager{
		limiters: make(map[string]*rate.Limiter),
		option: LimiterManagerOption{
			Limit:  l,
			Bucket: b,
		},
	}
}

func (lm *LimiterManager) Allow(field, id string) error {
	key := fmt.Sprintf("%s:%d", field, id)

	limiter := lm.getLimiter(key)
	if limiter == nil {
		limiter = lm.newLimiter(key, lm.getRateLimit(field))
	}

	if !limiter.Allow() {
		return status.Error(errcode.ErrRequestLimit, "request limit")
	}

	return nil
}

func (lm *LimiterManager) getLimiter(key string) *rate.Limiter {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	return lm.limiters[key]
}

func (lm *LimiterManager) newLimiter(key string, limit float64) *rate.Limiter {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	limiter := lm.limiters[key]
	if limiter != nil {
		return limiter
	}

	limiter = rate.NewLimiter(rate.Limit(limit), lm.option.Bucket)
	lm.limiters[key] = limiter
	return limiter
}

func (lm *LimiterManager) getRateLimit(field string) float64 {
	limit, err := redigo.Float64(redis.DB("limiter").Do("GET", field))
	if err != nil {
		return lm.option.Limit
	}
	return limit
}
