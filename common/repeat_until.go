package common

import (
	"context"
	"time"
)

func RepeatUntil(ctx context.Context, interval time.Duration, f func(ctx context.Context) (bool, error)) error {
	done, err := f(ctx)
	if done || err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			done, err := f(ctx)
			if done || err != nil {
				return err
			}
		}
	}
}
