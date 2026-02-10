package limiter

import (
	"context"
	"time"
	"github.com/redis/go-redis/v9"
)

const luaScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

local current = redis.call("INCR", key)
if current == 1 then
    redis.call("EXPIRE", key, window)
end

if current > limit then
    return 0
end
return 1
`

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

func (r *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	// O(1) complexity di sisi Redis
	res, err := r.client.Eval(ctx, luaScript, []string{key}, limit, int(window.Seconds())).Result()
	if err != nil {
		return false, err
	}
	return res.(int64) == 1, nil
}