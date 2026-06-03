package portfolio

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"api/db"
	"api/middleware"
	"api/models"
	"api/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateTradeRequest struct {
	AssetType string  `json:"asset_type"`
	Symbol    string  `json:"symbol"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Price     float64 `json:"price"`
	Total     float64 `json:"total"`
	Currency  string  `json:"currency"`
}

type TradeExport struct {
	UUID      string  `json:"uuid"`
	AssetType string  `json:"asset_type"`
	Symbol    string  `json:"symbol"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Price     float64 `json:"price"`
	Total     float64 `json:"total"`
	Currency  string  `json:"currency"`
	CreatedAt string  `json:"created_at"`
}

type ExportData struct {
	Version     string        `json:"version"`
	Exported    string        `json:"exported"`
	Trades      []TradeExport `json:"trades"`
	Fingerprint string        `json:"fingerprint"`
}

type ImportPreviewRequest struct {
	Data ExportData `json:"data"`
}

type ConflictItem struct {
	Trade  TradeExport `json:"trade"`
	Reason string      `json:"reason"`
}

type ImportPreviewResponse struct {
	TotalTrades   int            `json:"total_trades"`
	NewTrades     int            `json:"new_trades"`
	Conflicts     int            `json:"conflicts"`
	ConflictItems []ConflictItem `json:"conflict_items"`
}

type ImportConfirmRequest struct {
	Data             ExportData `json:"data"`
	ConflictStrategy string     `json:"conflict_strategy"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/portfolio")
	path = strings.TrimPrefix(path, "/")

	// 健康检查不需要认证
	if path == "health" {
		if r.Method == http.MethodGet {
			utils.Success(w, map[string]interface{}{
				"status": "healthy",
			})
			return
		}
		utils.Error(w, http.StatusNotFound, "接口不存在")
		return
	}

	// 所有其他路由都需要 JWT 认证
	middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// 根据路径和方法分发
		if path == "" || path == "dashboard" {
			if r.Method == http.MethodGet {
				handleDashboard(w, r)
				return
			}
		}

		if path == "trades" || strings.HasPrefix(path, "trades/") {
			handleTrades(w, r)
			return
		}

		if path == "export" {
			if r.Method == http.MethodGet {
				handleExport(w, r)
				return
			}
		}

		if path == "import/preview" {
			if r.Method == http.MethodPost {
				handleImportPreview(w, r)
				return
			}
		}

		if path == "import/confirm" {
			if r.Method == http.MethodPost {
				handleImportConfirm(w, r)
				return
			}
		}

		utils.Error(w, http.StatusNotFound, "接口不存在")
	})(w, r)
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	database := db.GetDB()

	var holdings []models.Holding
	database.Where("user_id = ?", userID).Find(&holdings)

	var recentTrades []models.Trade
	database.Where("user_id = ?", userID).Order("created_at desc").Limit(5).Find(&recentTrades)

	var totalTrades int64
	database.Model(&models.Trade{}).Where("user_id = ?", userID).Count(&totalTrades)

	utils.Success(w, map[string]interface{}{
		"holdings":      holdings,
		"recent_trades": recentTrades,
		"stats": map[string]interface{}{
			"total_trades": totalTrades,
		},
	})
}

func handleTrades(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTrades(w, r)
	case http.MethodPost:
		createTrade(w, r)
	case http.MethodDelete:
		if strings.HasSuffix(r.URL.Path, "/trades") {
			clearTrades(w, r)
		} else {
			deleteTrade(w, r)
		}
	default:
		utils.Error(w, http.StatusMethodNotAllowed, "方法不允许")
	}
}

func getTrades(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	var trades []models.Trade
	db.GetDB().Where("user_id = ?", userID).Order("created_at desc").Find(&trades)

	utils.Success(w, trades)
}

func createTrade(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	var req CreateTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "无效的请求体")
		return
	}

	if req.Symbol == "" || req.Type == "" || req.Amount <= 0 || req.Price <= 0 {
		utils.BadRequest(w, "缺少必填字段或数值无效")
		return
	}

	if req.AssetType == "" {
		req.AssetType = models.AssetTypeCrypto
	}
	if req.Currency == "" {
		req.Currency = models.GetCurrencyByAssetType(req.AssetType)
	}
	if req.Total == 0 {
		req.Total = req.Amount * req.Price
	}

	trade := models.Trade{
		UUID:      uuid.New().String(),
		UserID:    userID,
		AssetType: req.AssetType,
		Symbol:    req.Symbol,
		Type:      req.Type,
		Amount:    req.Amount,
		Price:     req.Price,
		Total:     req.Total,
		Currency:  req.Currency,
	}

	database := db.GetDB()
	if err := database.Create(&trade).Error; err != nil {
		utils.InternalError(w, "创建交易失败")
		return
	}

	if err := recalcHoldings(database, userID); err != nil {
		utils.InternalError(w, "更新持仓失败")
		return
	}

	utils.Success(w, trade)
}

func deleteTrade(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/portfolio/trades/")
	id, err := strconv.ParseUint(path, 10, 32)
	if err != nil {
		utils.BadRequest(w, "无效的交易ID")
		return
	}

	database := db.GetDB()

	result := database.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Trade{})
	if result.Error != nil {
		utils.InternalError(w, "删除交易失败")
		return
	}
	if result.RowsAffected == 0 {
		utils.NotFound(w, "交易记录不存在")
		return
	}

	if err := recalcHoldings(database, userID); err != nil {
		utils.InternalError(w, "更新持仓失败")
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "删除成功",
	})
}

func clearTrades(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	database := db.GetDB()

	if err := database.Unscoped().Where("user_id = ?", userID).Delete(&models.Trade{}).Error; err != nil {
		utils.InternalError(w, "清空交易失败")
		return
	}

	if err := database.Unscoped().Where("user_id = ?", userID).Delete(&models.Holding{}).Error; err != nil {
		utils.InternalError(w, "清空持仓失败")
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "清空成功",
	})
}

func recalcHoldings(tx *gorm.DB, uid uint) error {
	if err := tx.Where("user_id = ?", uid).Delete(&models.Holding{}).Error; err != nil {
		return err
	}

	var trades []models.Trade
	if err := tx.Where("user_id = ?", uid).Order("created_at asc").Find(&trades).Error; err != nil {
		return err
	}

	holdings := make(map[string]*models.Holding)
	var cashHolding *models.Holding

	for _, t := range trades {
		if t.Type == "recharge" {
			if cashHolding == nil {
				cashHolding = &models.Holding{
					UserID:    uid,
					AssetType: models.AssetTypeCash,
					Symbol:    "USD",
					Currency:  models.CurrencyUSD,
				}
			}
			cashHolding.Amount += t.Amount
			continue
		}

		key := t.AssetType + ":" + t.Symbol
		if holdings[key] == nil {
			holdings[key] = &models.Holding{
				UserID:    uid,
				AssetType: t.AssetType,
				Symbol:    t.Symbol,
				Currency:  t.Currency,
			}
		}

		switch t.Type {
		case "buy":
			holdings[key].Amount += t.Amount
			if cashHolding == nil {
				cashHolding = &models.Holding{
					UserID:    uid,
					AssetType: models.AssetTypeCash,
					Symbol:    "USD",
					Currency:  models.CurrencyUSD,
				}
			}
			cashHolding.Amount -= t.Total
		case "sell":
			holdings[key].Amount -= t.Amount
			if cashHolding == nil {
				cashHolding = &models.Holding{
					UserID:    uid,
					AssetType: models.AssetTypeCash,
					Symbol:    "USD",
					Currency:  models.CurrencyUSD,
				}
			}
			cashHolding.Amount += t.Total
		}
	}

	for _, h := range holdings {
		if h.Amount != 0 {
			if err := tx.Create(h).Error; err != nil {
				return err
			}
		}
	}

	if cashHolding != nil && cashHolding.Amount != 0 {
		if err := tx.Create(cashHolding).Error; err != nil {
			return err
		}
	}

	return nil
}

func handleExport(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	var trades []models.Trade
	if err := db.GetDB().Where("user_id = ?", userID).Order("created_at asc").Find(&trades).Error; err != nil {
		utils.InternalError(w, "获取交易记录失败")
		return
	}

	tradeExports := make([]TradeExport, len(trades))
	for i, t := range trades {
		tradeExports[i] = TradeExport{
			UUID:      t.UUID,
			AssetType: t.AssetType,
			Symbol:    t.Symbol,
			Type:      t.Type,
			Amount:    t.Amount,
			Price:     t.Price,
			Total:     t.Total,
			Currency:  t.Currency,
			CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	exported := time.Now().Format("2006-01-02 15:04:05")
	exportData := ExportData{
		Version:     "1.0",
		Exported:    exported,
		Trades:      tradeExports,
		Fingerprint: calculateFingerprint("1.0", exported, tradeExports),
	}

	utils.Success(w, map[string]interface{}{
		"data": exportData,
	})
}

func calculateFingerprint(version, exported string, trades []TradeExport) string {
	data := struct {
		Version  string        `json:"version"`
		Exported string        `json:"exported"`
		Trades   []TradeExport `json:"trades"`
	}{
		Version:  version,
		Exported: exported,
		Trades:   trades,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

func handleImportPreview(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	var req ImportPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "无效的请求数据")
		return
	}

	database := db.GetDB()

	var existingTrades []models.Trade
	database.Where("user_id = ?", userID).Find(&existingTrades)

	existingUUIDs := make(map[string]bool)
	for _, t := range existingTrades {
		if t.UUID != "" {
			existingUUIDs[t.UUID] = true
		}
	}

	preview := ImportPreviewResponse{
		TotalTrades:   len(req.Data.Trades),
		NewTrades:     0,
		Conflicts:     0,
		ConflictItems: []ConflictItem{},
	}

	for _, trade := range req.Data.Trades {
		key := trade.UUID
		if key == "" {
			key = fmt.Sprintf("%s_%s_%s_%s", trade.AssetType, trade.Symbol, trade.Type, trade.CreatedAt)
		}

		if existingUUIDs[key] {
			preview.Conflicts++
			preview.ConflictItems = append(preview.ConflictItems, ConflictItem{
				Trade:  trade,
				Reason: "与现有记录UUID相同",
			})
		} else {
			preview.NewTrades++
		}
	}

	utils.Success(w, map[string]interface{}{
		"preview": preview,
	})
}

func handleImportConfirm(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	var req ImportConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "无效的请求数据")
		return
	}

	if req.Data.Fingerprint != "" {
		expectedFingerprint := calculateFingerprint(req.Data.Version, req.Data.Exported, req.Data.Trades)
		if expectedFingerprint != req.Data.Fingerprint {
			utils.BadRequest(w, "数据指纹校验失败，文件可能已被篡改")
			return
		}
	}

	if req.ConflictStrategy != "skip" && req.ConflictStrategy != "overwrite" {
		req.ConflictStrategy = "skip"
	}

	database := db.GetDB()

	var existingTrades []models.Trade
	database.Where("user_id = ?", userID).Find(&existingTrades)

	existingUUIDs := make(map[string]uint)
	for _, t := range existingTrades {
		if t.UUID != "" {
			existingUUIDs[t.UUID] = t.ID
		}
	}

	var imported, skipped, overwritten int

	tx := database.Begin()

	for _, trade := range req.Data.Trades {
		key := trade.UUID
		if key == "" {
			key = fmt.Sprintf("%s_%s_%s_%s", trade.AssetType, trade.Symbol, trade.Type, trade.CreatedAt)
		}

		if existingID, exists := existingUUIDs[key]; exists {
			if req.ConflictStrategy == "overwrite" {
				if err := tx.Unscoped().Delete(&models.Trade{}, existingID).Error; err != nil {
					tx.Rollback()
					utils.InternalError(w, "删除旧记录失败")
					return
				}
				overwritten++
			} else {
				skipped++
				continue
			}
		}

		createdAt, _ := time.Parse("2006-01-02 15:04:05", trade.CreatedAt)
		if createdAt.IsZero() {
			createdAt = time.Now()
		}

		newTrade := models.Trade{
			UUID:      trade.UUID,
			UserID:    userID,
			AssetType: trade.AssetType,
			Symbol:    trade.Symbol,
			Type:      trade.Type,
			Amount:    trade.Amount,
			Price:     trade.Price,
			Total:     trade.Total,
			Currency:  trade.Currency,
		}
		if newTrade.UUID == "" {
			newTrade.UUID = uuid.New().String()
		}
		newTrade.CreatedAt = createdAt

		if err := tx.Create(&newTrade).Error; err != nil {
			tx.Rollback()
			utils.InternalError(w, "创建交易记录失败")
			return
		}

		imported++
	}

	if err := recalcHoldings(tx, userID); err != nil {
		tx.Rollback()
		utils.InternalError(w, "重新计算持仓失败")
		return
	}

	tx.Commit()

	utils.Success(w, map[string]interface{}{
		"imported":    imported,
		"skipped":     skipped,
		"overwritten": overwritten,
	})
}
