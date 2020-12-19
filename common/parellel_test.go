package common_test

import (
	"sync/atomic"
	"testing"

	"github.com/rickone/athena/common"
)

func TestParellel(t *testing.T) {
	sum := int64(0)
	ch := make(chan int)

	err := common.Parellel(
		func() error {
			for i := 0; i < 1000; i++ {
				ch <- i
			}
			close(ch)
			return nil
		}, func() error {
			for n := range ch {
				if n%2 == 0 {
					atomic.AddInt64(&sum, int64(n))
				}
			}
			return nil
		}, func() error {
			for n := range ch {
				if n%2 == 0 {
					atomic.AddInt64(&sum, int64(n))
				}
			}
			return nil
		}, func() error {
			for n := range ch {
				if n%2 == 0 {
					atomic.AddInt64(&sum, int64(n))
				}
			}
			return nil
		})()
	if err != nil {
		t.Fail()
	}

	sum2 := 0
	for i := 0; i < 1000; i++ {
		if i%2 == 0 {
			sum2 += i
		}
	}
	if sum != int64(sum2) {
		t.Fail()
	}
}
