package rateLimiter

import "time"

type RateLimiter interface {
	Allow(ip string) (bool, time.Duration)
}

type Config struct {
	RequestPerTF int
	TimeFrame    time.Duration
	Enabled      bool
}
