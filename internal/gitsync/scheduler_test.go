package gitsync

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type eventLog struct {
	mu     sync.Mutex
	events []Event
}

func (l *eventLog) add(ev Event) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, ev)
}

func (l *eventLog) snapshot() []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]Event(nil), l.events...)
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met in time")
}

func TestScheduler_RunsAtStartAndPeriodically(t *testing.T) {
	var runs atomic.Int32
	s := NewScheduler(func(context.Context) (Result, error) { runs.Add(1); return Result{}, nil }, 20*time.Millisecond, nil)
	s.Start()
	waitFor(t, func() bool { return runs.Load() >= 3 })
	s.Stop()
	after := runs.Load()
	time.Sleep(60 * time.Millisecond)
	if runs.Load() != after {
		t.Errorf("runs continued after Stop: %d -> %d", after, runs.Load())
	}
}

func TestScheduler_StopIsIdempotent(t *testing.T) {
	s := NewScheduler(func(context.Context) (Result, error) { return Result{}, nil }, time.Hour, nil)
	s.Stop()
	s.Start()
	s.Stop()
	s.Stop()
}

func TestScheduler_ConflictPausesUntilManualSuccess(t *testing.T) {
	var fail atomic.Bool
	fail.Store(true)
	var runs atomic.Int32
	log := &eventLog{}
	s := NewScheduler(func(context.Context) (Result, error) {
		runs.Add(1)
		if fail.Load() {
			return Result{}, &ConflictError{Files: []string{"x.md"}}
		}
		return Result{}, nil
	}, 10*time.Millisecond, log.add)

	s.Start()
	defer s.Stop()
	waitFor(t, func() bool { return s.Paused() })
	paused := runs.Load()
	time.Sleep(50 * time.Millisecond)
	if runs.Load() != paused {
		t.Errorf("automatic runs continued while paused: %d -> %d", paused, runs.Load())
	}

	fail.Store(false)
	if _, err := s.RunNow(context.Background()); err != nil {
		t.Fatal(err)
	}
	if s.Paused() {
		t.Error("still paused after successful manual run")
	}
	waitFor(t, func() bool { return runs.Load() > paused+1 })

	evs := log.snapshot()
	if !evs[0].Paused || evs[0].Manual {
		t.Errorf("first event = %+v, want automatic and paused", evs[0])
	}
}

func TestScheduler_NoOverlap(t *testing.T) {
	var active, maxActive atomic.Int32
	s := NewScheduler(func(context.Context) (Result, error) {
		n := active.Add(1)
		for {
			m := maxActive.Load()
			if n <= m || maxActive.CompareAndSwap(m, n) {
				break
			}
		}
		time.Sleep(30 * time.Millisecond)
		active.Add(-1)
		return Result{}, nil
	}, 5*time.Millisecond, nil)
	s.Start()
	var wg sync.WaitGroup
	for range 3 {
		wg.Add(1)
		go func() { defer wg.Done(); _, _ = s.RunNow(context.Background()) }()
	}
	wg.Wait()
	s.Stop()
	if maxActive.Load() != 1 {
		t.Errorf("max concurrent runs = %d, want 1", maxActive.Load())
	}
}

func TestScheduler_CanceledRunEmitsNoEvent(t *testing.T) {
	log := &eventLog{}
	started := make(chan struct{})
	s := NewScheduler(func(ctx context.Context) (Result, error) {
		close(started)
		<-ctx.Done()
		return Result{}, &GitError{Op: "pull", Err: ctx.Err()}
	}, time.Hour, log.add)
	s.Start()
	<-started
	s.Stop()
	if evs := log.snapshot(); len(evs) != 0 {
		t.Errorf("events after cancel = %+v, want none", evs)
	}
}

func TestScheduler_ManualEventFlag(t *testing.T) {
	log := &eventLog{}
	s := NewScheduler(func(context.Context) (Result, error) { return Result{Pushed: true}, errors.New("boom") }, time.Hour, log.add)
	_, _ = s.RunNow(context.Background())
	evs := log.snapshot()
	if len(evs) != 1 || !evs[0].Manual || evs[0].Err == nil || evs[0].At.IsZero() {
		t.Errorf("events = %+v, want one manual error event with time", evs)
	}
}
