package rateLimiter

import (
	"sync"
	"time"
)

type FWLimiter struct {
	sync.RWMutex
	clients map[string]int
	limit   int
	window  time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *FWLimiter {
	return &FWLimiter{
		clients: make(map[string]int),
		limit:   limit,
		window:  window,
	}
}

func (fw *FWLimiter) Allow(ip string) (bool, time.Duration) {
	return false, fw.window
}
