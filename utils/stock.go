package utils

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// StockPrice 股票价格信息
type StockPrice struct {
	Symbol    string  // 股票代码
	Open      float64 // 开盘价
	High      float64 // 最高价
	Low       float64 // 最低价
	Close     float64 // 收盘价
	PreClose  float64 // 昨收价
	Change    float64 // 涨跌额
	PctChg    float64 // 涨跌幅
	Volume    float64 // 成交量
	TradeDate string  // 交易日期
}

// PriceCache 价格缓存项
type PriceCache struct {
	Price      float64
	UpdateTime time.Time
}

// StockClient 股票价格客户端（统一接口）
type StockClient struct {
	cache    map[string]*PriceCache
	cacheMux sync.RWMutex
	http     *http.Client
}

var (
	stockInstance *StockClient
	stockOnce     sync.Once
)

// GetStockClient 获取股票价格客户端单例
func GetStockClient() *StockClient {
	stockOnce.Do(func() {
		stockInstance = &StockClient{
			cache: make(map[string]*PriceCache),
			http: &http.Client{
				Timeout: 10 * time.Second,
			},
		}
	})
	return stockInstance
}

const (
	cacheTTL       = 5 * time.Minute // 缓存有效期5分钟
	tencentAPIURL  = "https://qt.gtimg.cn"
)

// GetUSStockPrice 获取美股实时价格
func (c *StockClient) GetUSStockPrice(symbol string) (*StockPrice, error) {
	// 检查缓存
	if price, ok := c.getCachedPrice(symbol); ok {
		return &StockPrice{
			Symbol:    symbol,
			Close:     price,
			TradeDate: time.Now().Format("20060102"),
		}, nil
	}

	// 优先使用腾讯财经（无频率限制）
	if price, err := c.getFromTencent(symbol); err == nil && price.Close > 0 {
		c.setCachedPrice(symbol, price.Close)
		return price, nil
	}

	return nil, fmt.Errorf("获取股票价格失败: %s", symbol)
}

// GetUSStockPricesBatch 批量获取美股价格
func (c *StockClient) GetUSStockPricesBatch(symbols []string) map[string]*StockPrice {
	result := make(map[string]*StockPrice)

	// 检查缓存
	var needFetch []string
	for _, symbol := range symbols {
		if price, ok := c.getCachedPrice(symbol); ok {
			result[symbol] = &StockPrice{
				Symbol:    symbol,
				Close:     price,
				TradeDate: time.Now().Format("20060102"),
			}
		} else {
			needFetch = append(needFetch, symbol)
		}
	}

	// 从腾讯财经批量获取
	if len(needFetch) > 0 {
		tencentResults := c.getFromTencentBatch(needFetch)
		for symbol, price := range tencentResults {
			if price.Close > 0 {
				result[symbol] = price
				c.setCachedPrice(symbol, price.Close)
			}
		}
	}

	return result
}

// getCachedPrice 获取缓存价格
func (c *StockClient) getCachedPrice(symbol string) (float64, bool) {
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
func (c *StockClient) setCachedPrice(symbol string, price float64) {
	c.cacheMux.Lock()
	defer c.cacheMux.Unlock()

	c.cache[symbol] = &PriceCache{
		Price:      price,
		UpdateTime: time.Now(),
	}
}

// getFromTencent 从腾讯财经获取单个股票价格
func (c *StockClient) getFromTencent(symbol string) (*StockPrice, error) {
	url := fmt.Sprintf("%s/q=us%s", tencentAPIURL, symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://gu.qq.com/")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder())
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return c.parseTencentResponse(string(body), symbol)
}

// getFromTencentBatch 从腾讯财经批量获取股票价格
func (c *StockClient) getFromTencentBatch(symbols []string) map[string]*StockPrice {
	result := make(map[string]*StockPrice)

	var codes []string
	for _, symbol := range symbols {
		codes = append(codes, fmt.Sprintf("us%s", symbol))
	}

	url := fmt.Sprintf("%s/q=%s", tencentAPIURL, strings.Join(codes, ","))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://gu.qq.com/")

	resp, err := c.http.Do(req)
	if err != nil {
		return result
	}
	defer resp.Body.Close()

	reader := transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder())
	body, err := io.ReadAll(reader)
	if err != nil {
		return result
	}

	lines := strings.Split(string(body), ";")
	for i, line := range lines {
		if i >= len(symbols) {
			break
		}
		if line = strings.TrimSpace(line); line == "" {
			continue
		}
		if price, err := c.parseTencentResponse(line, symbols[i]); err == nil {
			result[symbols[i]] = price
		}
	}

	return result
}

// parseTencentResponse 解析腾讯财经响应
func (c *StockClient) parseTencentResponse(text, symbol string) (*StockPrice, error) {
	text = strings.TrimSpace(text)
	text = strings.TrimSuffix(text, ";")

	parts := strings.Split(text, "=\"")
	if len(parts) < 2 {
		return nil, fmt.Errorf("无效响应格式")
	}

	data := strings.TrimSuffix(parts[1], "\"")
	fields := strings.Split(data, "~")

	if len(fields) < 35 {
		return nil, fmt.Errorf("响应字段不足")
	}

	price, _ := strconv.ParseFloat(fields[3], 64)
	preClose, _ := strconv.ParseFloat(fields[4], 64)
	open, _ := strconv.ParseFloat(fields[5], 64)
	high, _ := strconv.ParseFloat(fields[33], 64)
	low, _ := strconv.ParseFloat(fields[34], 64)
	change, _ := strconv.ParseFloat(fields[31], 64)
	changePct, _ := strconv.ParseFloat(fields[32], 64)
	volume, _ := strconv.ParseFloat(fields[36], 64)

	return &StockPrice{
		Symbol:    symbol,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     price,
		PreClose:  preClose,
		Change:    change,
		PctChg:    changePct,
		Volume:    volume,
		TradeDate: time.Now().Format("20060102"),
	}, nil
}
