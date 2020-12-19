package metrics

import (
	"runtime"
	"time"
)

const (
	watchInterval = 10 * time.Second
)

func watchNumGoroutine() {
	g := NewGauge("numgo")
	t := time.Tick(watchInterval)
	for {
		select {
		case <-t:
			g.Update(int64(runtime.NumGoroutine()))
		}
	}
}
