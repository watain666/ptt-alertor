package jobs

import (
	"context"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestPollEmptyListWaitsBetweenRetries(t *testing.T) {
	for _, empty := range [][]string{nil, {}} {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			start := time.Now()
			calls := 0
			poll(ctx, func() time.Duration { return time.Second }, func() []string {
				calls++
				if calls == 3 {
					cancel()
				}
				return empty
			}, func(string) {
				t.Error("empty list triggered a check")
			})
			if elapsed := time.Since(start); elapsed != 3*time.Second {
				t.Fatalf("three empty polls took %v, want 3s", elapsed)
			}
		})
	}
}

func TestPollRefreshesListsAndPacesChecks(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lists := [][]string{nil, {"a", "b"}, {}, {"c"}, nil}
		var calls atomic.Int32
		var mu sync.Mutex
		var checked []string
		go poll(ctx, func() time.Duration { return time.Second }, func() []string {
			return lists[calls.Add(1)-1]
		}, func(item string) {
			mu.Lock()
			defer mu.Unlock()
			checked = append(checked, item)
		})
		synctest.Wait()
		if calls.Load() != 0 {
			t.Fatal("list fetched before the first interval")
		}
		for _, want := range []struct {
			calls   int32
			checked []string
		}{
			{1, nil},
			{2, []string{"a"}},
			{2, []string{"a", "b"}},
			{3, []string{"a", "b"}},
			{4, []string{"a", "b", "c"}},
			{5, []string{"a", "b", "c"}},
		} {
			time.Sleep(time.Second)
			synctest.Wait()
			mu.Lock()
			got := slices.Clone(checked)
			mu.Unlock()
			if calls.Load() != want.calls || !slices.Equal(got, want.checked) {
				t.Fatalf("got %d list calls and checks %v, want %d and %v", calls.Load(), got, want.calls, want.checked)
			}
		}
	})
}

func TestPollUsesUpdatedInterval(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		start := time.Now()
		duration := time.Second
		var mu sync.Mutex
		var checkedAt []time.Duration
		go poll(ctx, func() time.Duration { return duration }, func() []int {
			// The updated rate applies between items, even within a single pass.
			duration = 2 * time.Second
			return []int{1, 2, 3}
		}, func(int) {
			mu.Lock()
			defer mu.Unlock()
			checkedAt = append(checkedAt, time.Since(start))
		})
		for range 5 {
			time.Sleep(time.Second)
			synctest.Wait()
		}
		want := []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}
		mu.Lock()
		defer mu.Unlock()
		if !slices.Equal(checkedAt, want) {
			t.Fatalf("check times = %v, want %v", checkedAt, want)
		}
	})
}

func TestPollWaitsAfterSlowList(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		start := time.Now()
		var mu sync.Mutex
		var checkedAt []time.Duration
		go poll(ctx, func() time.Duration { return time.Second }, func() []int {
			time.Sleep(3 * time.Second)
			return []int{1, 2}
		}, func(int) {
			mu.Lock()
			defer mu.Unlock()
			checkedAt = append(checkedAt, time.Since(start))
		})
		for range 5 {
			time.Sleep(time.Second)
			synctest.Wait()
		}
		want := []time.Duration{4 * time.Second, 5 * time.Second}
		mu.Lock()
		defer mu.Unlock()
		if !slices.Equal(checkedAt, want) {
			t.Fatalf("check times = %v, want %v", checkedAt, want)
		}
	})
}

func TestPollDoesNotWaitForChecks(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		release := make(chan struct{})
		defer close(release)
		started := make(chan int, 2)
		go poll(ctx, func() time.Duration { return time.Second }, func() []int {
			return []int{1, 2}
		}, func(item int) {
			started <- item
			<-release
		})
		time.Sleep(2 * time.Second)
		synctest.Wait()
		if len(started) != 2 {
			t.Fatalf("started %d checks, want 2 while the first is still running", len(started))
		}
	})
}

func TestPollCancellationStopsWaitingAndPendingChecks(t *testing.T) {
	for _, items := range [][]int{nil, {1, 2, 3}} {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan struct{})
			var calls, checks atomic.Int32
			go func() {
				defer close(done)
				poll(ctx, func() time.Duration { return time.Second }, func() []int {
					calls.Add(1)
					return items
				}, func(int) { checks.Add(1) })
			}()
			time.Sleep(time.Second)
			synctest.Wait()
			wantChecks := int32(min(1, len(items)))
			cancel()
			synctest.Wait()
			select {
			case <-done:
			default:
				t.Fatal("poll did not stop immediately on cancellation")
			}
			time.Sleep(3 * time.Second)
			synctest.Wait()
			if calls.Load() != 1 || checks.Load() != wantChecks {
				t.Fatalf("got %d list calls and %d checks, want 1 and %d", calls.Load(), checks.Load(), wantChecks)
			}
		})
	}
}

func TestPollAlreadyCancelled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		start := time.Now()
		poll(ctx, func() time.Duration { return time.Second }, func() []int {
			t.Error("cancelled poll fetched a list")
			return []int{1}
		}, func(int) { t.Error("cancelled poll started a check") })
		if time.Since(start) != 0 {
			t.Fatal("cancelled poll waited for the timer")
		}
	})
}

func TestPollCancelledDuringList(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		poll(ctx, func() time.Duration { return time.Second }, func() []int {
			cancel()
			return []int{1}
		}, func(int) { t.Error("cancelled poll started a check") })
	})
}
