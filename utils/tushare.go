package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitee.com/palagend/fx/config"
)

const TushareAPIURL = "https://api.tushare.pro"

// TushareClient Tushare API客户端
type TushareClient struct {
	token  string
	client *http.Client
}

// NewTushareClient 创建Tushare客户端
func NewTushareClient() *TushareClient {
	cfg := config.GetConfig()
	token := cfg.API.TushareKey
	if token == "" {
		token = "your_tushare_token_here"
	}
	return &TushareClient{
		token: token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TushareRequest Tushare请求结构
type TushareRequest struct {
	Token string `json:"token"`
	API   string `json:"api_name"`
	Param string `json:"params"`
}

// TushareResponse Tushare响应结构
type TushareResponse struct {
	Code    int             `json:"code"`
	Msg     string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
}

// DailyPrice 日线数据
type DailyPrice struct {
	TSCode    string  `json:"ts_code"`
	TradeDate string  `json:"trade_date"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	PreClose  float64 `json:"pre_close"`
	Change    float64 `json:"change"`
	PctChg    float64 `json:"pct_chg"`
	Vol       float64 `json:"vol"`
	Amount    float64 `json:"amount"`
}

// GetUSStockPrice 获取美股实时价格
// ts_code格式: AAPL.US
func (c *TushareClient) GetUSStockPrice(tsCode string) (*DailyPrice, error) {
	// 构建参数
	params := map[string]string{
		"ts_code": tsCode,
		"limit":   "1",
	}
	paramJSON, _ := json.Marshal(params)

	reqBody := TushareRequest{
		Token: c.token,
		API:   "us_daily",
		Param: string(paramJSON),
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", TushareAPIURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TushareResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("Tushare API错误: %s", result.Msg)
	}

	// 解析数据
	var data struct {
		Fields []string    `json:"fields"`
		Items  [][]interface{} `json:"items"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		return nil, err
	}

	if len(data.Items) == 0 {
		return nil, fmt.Errorf("未找到股票数据: %s", tsCode)
	}

	// 将数组转换为结构体
	price := &DailyPrice{}
	for i, field := range data.Fields {
		if i >= len(data.Items[0]) {
			continue
		}
		value := data.Items[0][i]
		switch field {
		case "ts_code":
			price.TSCode = value.(string)
		case "trade_date":
			price.TradeDate = value.(string)
		case "open":
			price.Open = parseFloat(value)
		case "high":
			price.High = parseFloat(value)
		case "low":
			price.Low = parseFloat(value)
		case "close":
			price.Close = parseFloat(value)
		case "pre_close":
			price.PreClose = parseFloat(value)
		case "change":
			price.Change = parseFloat(value)
		case "pct_chg":
			price.PctChg = parseFloat(value)
		case "vol":
			price.Vol = parseFloat(value)
		case "amount":
			price.Amount = parseFloat(value)
		}
	}

	return price, nil
}

// GetUSStockList 获取美股列表
func (c *TushareClient) GetUSStockList() ([]map[string]interface{}, error) {
	reqBody := TushareRequest{
		Token: c.token,
		API:   "us_stock_basic",
		Param: "{}",
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", TushareAPIURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TushareResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("Tushare API错误: %s", result.Msg)
	}

	var data struct {
		Fields []string        `json:"fields"`
		Items  [][]interface{} `json:"items"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		return nil, err
	}

	// 转换为map列表
	var stocks []map[string]interface{}
	for _, item := range data.Items {
		stock := make(map[string]interface{})
		for i, field := range data.Fields {
			if i < len(item) {
				stock[field] = item[i]
			}
		}
		stocks = append(stocks, stock)
	}

	return stocks, nil
}

// parseFloat 安全解析float
func parseFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0
	}
}

// SymbolToTSCode 将股票代码转换为Tushare格式
// AAPL -> AAPL.US
func SymbolToTSCode(symbol string) string {
	if strings.HasSuffix(symbol, ".US") {
		return symbol
	}
	return symbol + ".US"
}
