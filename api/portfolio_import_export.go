package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gitee.com/palagend/fx/config"
	"gitee.com/palagend/fx/models"
)




// ========== 导入/导出接口 ==========

// ExportData 导出用户数据
type ExportData struct {
	Version     string        `json:"version"`
	Exported    string        `json:"exported"`
	Trades      []TradeExport `json:"trades"`
	Fingerprint string        `json:"fingerprint"` // 数据指纹，用于防篡改校验
}

// calculateFingerprint 计算数据指纹（SHA-256）
func calculateFingerprint(version, exported string, trades []TradeExport) string {
	// 构建需要校验的数据结构
	data := struct {
		Version  string        `json:"version"`
		Exported string        `json:"exported"`
		Trades   []TradeExport `json:"trades"`
	}{
		Version:  version,
		Exported: exported,
		Trades:   trades,
	}

	// 序列化为JSON（确保顺序一致）
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	// 计算SHA-256哈希
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

type TradeExport struct {
	ID        int     `json:"id"`
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

// ExportDataHandler 导出数据接口
func ExportDataHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	var trades []models.Trade
	if err := db.Where("user_id = ?", uid).Order("created_at asc").Find(&trades).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易记录失败"})
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

	c.JSON(http.StatusOK, gin.H{"data": exportData})
}

// ImportPreviewRequest 导入预览请求
type ImportPreviewRequest struct {
	Data ExportData `json:"data"`
}

// ImportPreviewResponse 导入预览响应
type ImportPreviewResponse struct {
	TotalTrades   int            `json:"total_trades"`
	NewTrades     int            `json:"new_trades"`
	Conflicts     int            `json:"conflicts"`
	ConflictItems []ConflictItem `json:"conflict_items"`
}

type ConflictItem struct {
	Trade  TradeExport `json:"trade"`
	Reason string      `json:"reason"`
}

// ImportPreviewHandler 导入预览接口
func ImportPreviewHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req ImportPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	// 获取用户现有交易记录的UUID集合（用于检测冲突）
	var existingTrades []models.Trade
	db.Where("user_id = ?", uid).Find(&existingTrades)

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
		// 使用UUID检测冲突，如果没有UUID则使用时间戳作为备用
		var key string
		if trade.UUID != "" {
			key = trade.UUID
		} else {
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

	c.JSON(http.StatusOK, gin.H{"preview": preview})
}

// ImportConfirmRequest 导入确认请求
type ImportConfirmRequest struct {
	Data             ExportData `json:"data"`
	ConflictStrategy string     `json:"conflict_strategy"` // skip 或 overwrite
}

// ImportConfirmHandler 导入确认接口
func ImportConfirmHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req ImportConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 验证数据指纹（如果存在）
	if req.Data.Fingerprint != "" {
		expectedFingerprint := calculateFingerprint(req.Data.Version, req.Data.Exported, req.Data.Trades)
		if expectedFingerprint != req.Data.Fingerprint {
			c.JSON(http.StatusBadRequest, gin.H{"error": "数据指纹校验失败，文件可能已被篡改"})
			return
		}
	}

	if req.ConflictStrategy != "skip" && req.ConflictStrategy != "overwrite" {
		req.ConflictStrategy = "skip"
	}

	db := config.GetDB()
	uid := userID.(uint)

	// 获取用户现有交易记录的UUID集合
	var existingTrades []models.Trade
	db.Where("user_id = ?", uid).Find(&existingTrades)

	existingUUIDs := make(map[string]uint) // UUID -> trade ID
	for _, t := range existingTrades {
		if t.UUID != "" {
			existingUUIDs[t.UUID] = t.ID
		}
	}

	var imported, skipped, overwritten int

	tx := db.Begin()

	for _, trade := range req.Data.Trades {
		// 使用UUID检测冲突，如果没有UUID则使用时间戳作为备用
		var key string
		if trade.UUID != "" {
			key = trade.UUID
		} else {
			key = fmt.Sprintf("%s_%s_%s_%s", trade.AssetType, trade.Symbol, trade.Type, trade.CreatedAt)
		}

		if existingID, exists := existingUUIDs[key]; exists {
			// 存在冲突
			if req.ConflictStrategy == "overwrite" {
				// 删除旧记录
				if err := tx.Unscoped().Delete(&models.Trade{}, existingID).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "删除旧记录失败"})
					return
				}
				overwritten++
			} else {
				// 跳过
				skipped++
				continue
			}
		}

		// 解析时间
		createdAt, _ := time.Parse("2006-01-02 15:04:05", trade.CreatedAt)
		if createdAt.IsZero() {
			createdAt = time.Now()
		}

		// 创建新记录，保留原始UUID（如果有），否则生成新的UUID
		newTrade := models.Trade{
			UUID:      trade.UUID,
			UserID:    uid,
			AssetType: trade.AssetType,
			Symbol:    trade.Symbol,
			Type:      trade.Type,
			Amount:    trade.Amount,
			Price:     trade.Price,
			Total:     trade.Total,
			Currency:  trade.Currency,
		}
		// 如果导入的数据没有UUID，生成一个新的
		if newTrade.UUID == "" {
			newTrade.UUID = uuid.New().String()
		}
		newTrade.CreatedAt = createdAt

		if err := tx.Create(&newTrade).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建交易记录失败"})
			return
		}

		imported++
	}

	// 重新计算持仓
	if err := recalcAllHoldings(tx, uid); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重新计算持仓失败"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交导入失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"imported":    imported,
		"skipped":     skipped,
		"overwritten": overwritten,
	})
}
