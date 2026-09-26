package jobs

import (
	"context"
	"time"
)

// poll starts one asynchronous check per interval, refreshing the list after
// each pass. Waiting before each poll also paces retries when the list is empty.
// interval is evaluated between checks so callers can adjust the polling rate.
func poll[T any](ctx context.Context, interval func() time.Duration, list func() []T, check func(T)) {
	timer := time.NewTimer(interval())
	defer timer.Stop()

	var pending []T
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		if ctx.Err() != nil {
			return
		}

		if len(pending) == 0 {
			pending = list()
		}
		if ctx.Err() != nil {
			return
		}
		if len(pending) > 0 {
			item := pending[0]
			pending = pending[1:]
			go check(item)
		}
		timer.Reset(interval())
	}
}
