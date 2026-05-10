package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// ExchangeRateService 汇率服务
type ExchangeRateService struct {
	rates     map[string]float64 // 以USD为基准的汇率
	lastUpdate time.Time
	mutex     sync.RWMutex
	client    *http.Client
}

var (
	exchangeService *ExchangeRateService
	once            sync.Once
)

// GetExchangeRateService 获取汇率服务单例
func GetExchangeRateService() *ExchangeRateService {
	once.Do(func() {
		exchangeService = &ExchangeRateService{
			rates: make(map[string]float64),
			client: &http.Client{
				Timeout: 10 * time.Second,
			},
		}
		// 初始化默认汇率
		exchangeService.rates["USD"] = 1.0
		exchangeService.rates["CNY"] = 0.138  // 1 CNY = 0.138 USD (约7.25汇率)
		exchangeService.rates["HKD"] = 0.128  // 1 HKD = 0.128 USD (约7.8汇率)
	})
	return exchangeService
}

// GetRate 获取汇率（转换为USD）
func (s *ExchangeRateService) GetRate(currency string) float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if rate, ok := s.rates[currency]; ok {
		return rate
	}
	return 1.0 // 默认1:1
}

// ConvertToUSD 将金额转换为USD
func (s *ExchangeRateService) ConvertToUSD(amount float64, fromCurrency string) float64 {
	rate := s.GetRate(fromCurrency)
	return amount * rate
}

// UpdateRates 更新汇率（从外部API）
func (s *ExchangeRateService) UpdateRates() error {
	// 使用 exchangerate-api.com (免费版)
	url := "https://api.exchangerate-api.com/v4/latest/USD"

	resp, err := s.client.Get(url)
	if err != nil {
		return fmt.Errorf("获取汇率失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result struct {
		Base  string             `json:"base"`
		Date  string             `json:"date"`
		Rates map[string]float64 `json:"rates"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 更新汇率（转换为以USD为基准）
	// API返回的是 1 USD = X CNY，我们需要 1 CNY = X USD
	if cnyRate, ok := result.Rates["CNY"]; ok && cnyRate > 0 {
		s.rates["CNY"] = 1 / cnyRate
	}
	if hkdRate, ok := result.Rates["HKD"]; ok && hkdRate > 0 {
		s.rates["HKD"] = 1 / hkdRate
	}

	s.lastUpdate = time.Now()
	return nil
}

// GetLastUpdateTime 获取最后更新时间
func (s *ExchangeRateService) GetLastUpdateTime() time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.lastUpdate
}

// GetAllRates 获取所有汇率
func (s *ExchangeRateService) GetAllRates() map[string]float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	rates := make(map[string]float64)
	for k, v := range s.rates {
		rates[k] = v
	}
	return rates
}

// StartAutoUpdate 启动自动更新（每小时）
func (s *ExchangeRateService) StartAutoUpdate() {
	// 立即更新一次
	s.UpdateRates()

	// 每小时更新
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			s.UpdateRates()
		}
	}()
}
