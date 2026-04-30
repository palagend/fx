package utils

import (
	"context"
	"time"
)

// RateLimiter 令牌桶限流器
type RateLimiter struct {
	tokens    chan struct{}
	ticker    *time.Ticker
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewRateLimiter 创建限流器
// rate: 每秒产生的令牌数
// burst: 桶容量（突发流量）
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	rl := &RateLimiter{
		tokens: make(chan struct{}, burst),
		ticker: time.NewTicker(time.Duration(float64(time.Second) / rate)),
		ctx:    ctx,
		cancel: cancel,
}

	// 初始化填满令牌桶
	for i := 0; i < burst; i++ {
		rl.tokens <- struct{}{}
	}

	// 启动令牌生产协程
	go rl.produceTokens()

	return rl
}

// produceTokens 持续生产令牌
func (rl *RateLimiter) produceTokens() {
	for {
		select {
		case <-rl.ctx.Done():
			return
		case <-rl.ticker.C:
			select {
			case rl.tokens <- struct{}{}:
			default:
				// 桶已满，丢弃令牌
			}
		}
	}
}

// Wait 等待获取一个令牌（阻塞）
func (rl *RateLimiter) Wait() {
	<-rl.tokens
}

// WaitWithTimeout 等待获取令牌，带超时
func (rl *RateLimiter) WaitWithTimeout(timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case <-rl.tokens:
		return true
	case <-ctx.Done():
		return false
	}
}

// TryAcquire 尝试获取令牌（非阻塞）
func (rl *RateLimiter) TryAcquire() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// Stop 停止限流器
func (rl *RateLimiter) Stop() {
	rl.cancel()
	rl.ticker.Stop()
}
