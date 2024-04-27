package server

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ready int32 = iota
	stopping
	stopped
)

type Readiness struct {
	state int32
	wg    sync.WaitGroup
}

func (r *Readiness) Shutdown(n int) {
	r.wg.Add(n)
	atomic.StoreInt32(&r.state, stopping)
}

func (r *Readiness) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	ready := atomic.CompareAndSwapInt32(&r.state, ready, ready)
	if !ready {
		http.Error(w, "not ready", http.StatusServiceUnavailable)

		stopping := atomic.CompareAndSwapInt32(&r.state, stopping, stopping)
		if stopping {
			r.wg.Done()
		}

		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, http.StatusText(http.StatusOK))
}

func (r *Readiness) Wait(deadline time.Duration) {
	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		atomic.StoreInt32(&r.state, stopped)
		close(done)
	}()

	select {
	case <-done:
		return
	case <-time.After(deadline):
		return
	}
}
