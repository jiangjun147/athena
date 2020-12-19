package common_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rickone/athena/common"
)

func TestRepeatTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()

	i := 0
	common.RepeatUntil(ctx, 200*time.Millisecond, func(ctx context.Context) (bool, error) {
		i++
		t.Log(i)
		return false, nil
	})
}

func TestRepeatDone(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()

	i := 0
	common.RepeatUntil(ctx, 200*time.Millisecond, func(ctx context.Context) (bool, error) {
		i++
		t.Log(i)
		return i > 3, nil
	})
}

func TestRepeatError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()

	i := 0
	common.RepeatUntil(ctx, 200*time.Millisecond, func(ctx context.Context) (bool, error) {
		i++
		t.Log(i)
		if i > 3 {
			return false, errors.New("i > 3")
		}
		return false, nil
	})
}
