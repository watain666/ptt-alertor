package jobs

import (
	"context"
	"slices"
	"testing"
	"testing/synctest"
	"time"
)

func TestParseHighBoards(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  []string
	}{
		{"", nil},
		{" , \t,\n", nil},
		{"Gossiping", []string{"Gossiping"}},
		{" Gossiping , , Stock,\tLifeismoney ,", []string{"Gossiping", "Stock", "Lifeismoney"}},
	} {
		t.Run(tt.input, func(t *testing.T) {
			var names []string
			for _, bd := range parseHighBoards(tt.input) {
				names = append(names, bd.Name)
			}
			if !slices.Equal(names, tt.want) {
				t.Errorf("board names = %v, want %v", names, tt.want)
			}
		})
	}
}

func TestCheckOffPeakCanStopWhileSending(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan struct{})
		go func() {
			defer close(done)
			Checker{}.checkOffPeak(ctx, make(chan bool))
		}()
		// Let the first update block because the polling loop isn't receiving.
		time.Sleep(10 * time.Minute)
		synctest.Wait()
		cancel()
		synctest.Wait()
		select {
		case <-done:
		default:
			t.Fatal("off-peak monitor did not stop while its update was blocked")
		}
	})
}
