package ratelimiter

import "time"

//go:generate mockgen -source=./ratelimiter.go -destination=../../cmd/api/mock/ratelimiter/Mock_RateLimiter.go -package=mock_limiter Limiter
type Limiter interface {
	Allow(ip string) (bool, time.Duration)
}

type Config struct {
	RequestsPerTimeFrame int
	TimeFrame            time.Duration
	Enabled              bool
}
