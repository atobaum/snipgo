package gitsync

import (
	"context"
	"errors"
	"sync"
	"time"
)

type SyncFunc func(ctx context.Context) (Result, error)

// Scheduler runs automatic syncs at start and on an interval, and manual syncs on demand.
type Scheduler struct {
	fn       SyncFunc
	interval time.Duration
	onEvent  func(Event)

	runMu sync.Mutex

	mu     sync.Mutex
	paused bool
	stop   chan struct{}
	done   chan struct{}
}

func NewScheduler(fn SyncFunc, interval time.Duration, onEvent func(Event)) *Scheduler {
	return &Scheduler{fn: fn, interval: interval, onEvent: onEvent}
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.stop != nil {
		s.mu.Unlock()
		return
	}
	stop, done := make(chan struct{}), make(chan struct{})
	s.stop, s.done = stop, done
	s.mu.Unlock()

	go func() {
		defer close(done)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			select {
			case <-stop:
				cancel()
			case <-ctx.Done():
			}
		}()

		s.runAuto(ctx)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				s.runAuto(ctx)
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	stop, done := s.stop, s.done
	s.stop, s.done = nil, nil
	s.mu.Unlock()
	if stop == nil {
		return
	}
	close(stop)
	<-done
}

func (s *Scheduler) Paused() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.paused
}

func (s *Scheduler) RunNow(ctx context.Context) (Result, error) {
	s.runMu.Lock()
	defer s.runMu.Unlock()
	return s.run(ctx, true)
}

func (s *Scheduler) runAuto(ctx context.Context) {
	if s.Paused() || ctx.Err() != nil {
		return
	}
	if !s.runMu.TryLock() {
		return
	}
	defer s.runMu.Unlock()
	_, _ = s.run(ctx, false)
}

func (s *Scheduler) run(ctx context.Context, manual bool) (Result, error) {
	res, err := s.fn(ctx)
	if errors.Is(err, context.Canceled) {
		return res, err
	}

	var conflict *ConflictError
	s.mu.Lock()
	switch {
	case errors.As(err, &conflict):
		s.paused = true
	case err == nil && manual:
		s.paused = false
	}
	paused := s.paused
	s.mu.Unlock()

	if s.onEvent != nil {
		s.onEvent(Event{At: time.Now(), Result: res, Err: err, Manual: manual, Paused: paused})
	}
	return res, err
}
