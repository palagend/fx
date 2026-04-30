package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"gitee.com/palagend/fx/config"
)

const ITickAPIURL = "https://api-free.itick.org"

// 缓存配置
const (
	cacheTTL = 5 * time.Minute // 缓存有效期5分钟
)

// PriceCache 价格缓存项
type PriceCache struct {
	Price      float64
	UpdateTime time.Time
}

// ITickClient iTick API客户端
type ITickClient struct {
	apiKey      string
	client      *http.Client
	cache       map[string]*PriceCache
	cacheMux    sync.RWMutex
	rateLimiter *RateLimiter
}

var (
	itickInstance *ITickClient
	itickOnce     sync.Once
)

// GetITickClient 获取iTick客户端单例
func GetITickClient() *ITickClient {
	itickOnce.Do(func() {
		cfg := config.GetConfig()
		apiKey := cfg.API.ITickKey
		if apiKey == "" {
			apiKey = "your_itick_api_key_here"
		}
		itickInstance = &ITickClient{
			apiKey: apiKey,
			client: &http.Client{
				Timeout: 10 * time.Second,
			},
			cache:       make(map[string]*PriceCache),
			rateLimiter: NewRateLimiter(5.0/60.0, 5), // 每分钟5次，桶容量5
		}
	})
	return itickInstance
}

// NewITickClient 创建iTick客户端（兼容旧代码，实际返回单例）
func NewITickClient() *ITickClient {
	return GetITickClient()
}

// ITickQuoteResponse iTick实时行情响应
type ITickQuoteResponse struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data ITickQuoteData `json:"data"`
}

// ITickQuoteData iTick行情数据
type ITickQuoteData struct {
	Symbol    string  `json:"s"`  // 股票代码
	Latest    float64 `json:"ld"` // 最新价格
	Open      float64 `json:"o"`  // 开盘价
	PreClose  float64 `json:"p"`  // 昨收价
	High      float64 `json:"h"`  // 最高价
	Low       float64 `json:"l"`  // 最低价
	Timestamp int64   `json:"t"`  // 时间戳
	Volume    float64 `json:"v"`  // 成交量
}

// StockPrice 股票价格信息（兼容旧接口）
type StockPrice struct {
	Symbol    string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	PreClose  float64
	Change    float64
	PctChg    float64
	Volume    float64
	TradeDate string
}

// getCachedPrice 获取缓存价格
func (c *ITickClient) getCachedPrice(symbol string) (float64, bool) {
	c.cacheMux.RLock()
	defer c.cacheMux.RUnlock()

	if cache, ok := c.cache[symbol]; ok {
		if time.Since(cache.UpdateTime) < cacheTTL {
			return cache.Price, true
		}
	}
	return 0, false
}

// setCachedPrice 设置缓存价格
func (c *ITickClient) setCachedPrice(symbol string, price float64) {
	c.cacheMux.Lock()
	defer c.cacheMux.Unlock()

	c.cache[symbol] = &PriceCache{
		Price:      price,
		UpdateTime: time.Now(),
	}
}

// GetUSStockPrice 获取美股实时价格
// symbol: 股票代码，如 AAPL
func (c *ITickClient) GetUSStockPrice(symbol string) (*StockPrice, error) {
	// 检查缓存
	if price, ok := c.getCachedPrice(symbol); ok {
		return &StockPrice{
			Symbol:    symbol,
			Close:     price,
			TradeDate: time.Now().Format("20060102"),
		}, nil
	}

	// 限流：等待获取令牌
	c.rateLimiter.Wait()

	url := fmt.Sprintf("%s/stock/quote?region=US&code=%s", ITickAPIURL, symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("token", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result ITickQuoteResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("iTick API错误: %s", result.Msg)
	}

	data := result.Data
	price := &StockPrice{
		Symbol:    data.Symbol,
		Open:      data.Open,
		High:      data.High,
		Low:       data.Low,
		Close:     data.Latest,
		PreClose:  data.PreClose,
		Change:    data.Latest - data.PreClose,
		PctChg:    ((data.Latest - data.PreClose) / data.PreClose) * 100,
		Volume:    data.Volume,
		TradeDate: time.Now().Format("20060102"),
	}

	// 更新缓存
	c.setCachedPrice(symbol, price.Close)

	return price, nil
}

// GetUSStockPricesBatch 批量获取美股价格
// 优先使用缓存，只请求需要更新的股票
func (c *ITickClient) GetUSStockPricesBatch(symbols []string) map[string]*StockPrice {
	result := make(map[string]*StockPrice)

	for _, symbol := range symbols {
		price, err := c.GetUSStockPrice(symbol)
		if err == nil {
			result[symbol] = price
		}
	}

	return result
}
