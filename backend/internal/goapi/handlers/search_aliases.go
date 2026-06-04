package handlers

import (
	"net/http"
	"strconv"

	"smlcloudplatform/internal/goapi/logger"
	mypg "smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
)

// ==================== Search Aliases CRUD ====================

// SearchAlias — model สำหรับ alias
type SearchAlias struct {
	ID        int    `json:"id"`
	Alias     string `json:"alias"`
	Target    string `json:"target"`
	AliasType string `json:"alias_type"`
}

// SearchAliasCreateRequest — request body สำหรับสร้าง alias
type SearchAliasCreateRequest struct {
	HoldingCode string `json:"holding_code"`
	Alias       string `json:"alias"`
	Target      string `json:"target"`
	AliasType   string `json:"alias_type"`
}

// SearchAliasListRequest — request body สำหรับดูรายการ aliases
type SearchAliasListRequest struct {
	HoldingCode string `json:"holding_code"`
}

// SearchAliasDeleteRequest — request body สำหรับลบ alias
type SearchAliasDeleteRequest struct {
	HoldingCode string `json:"holding_code"`
}

// SearchAliasListHandler — GET /api/search/aliases?holding_code=xxx
func SearchAliasListHandler(c echo.Context) error {
	holdingCode := c.QueryParam("holding_code")
	if holdingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Missing required parameter: holding_code",
		})
	}

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("SearchAliasListHandler: connect PG: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Database connection failed",
		})
	}

	rows, err := db.Query("SELECT id, alias, target, COALESCE(alias_type,'brand') FROM search_aliases ORDER BY alias")
	if err != nil {
		logger.Error("SearchAliasListHandler: query: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Query failed",
		})
	}
	defer rows.Close()

	var aliases []SearchAlias
	for rows.Next() {
		var a SearchAlias
		if err := rows.Scan(&a.ID, &a.Alias, &a.Target, &a.AliasType); err != nil {
			logger.Error("SearchAliasListHandler: scan: %v", err)
			continue
		}
		aliases = append(aliases, a)
	}

	if aliases == nil {
		aliases = []SearchAlias{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    aliases,
		"total":   len(aliases),
	})
}

// SearchAliasCreateHandler — POST /api/search/aliases
func SearchAliasCreateHandler(c echo.Context) error {
	var req SearchAliasCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
	}

	if req.HoldingCode == "" || req.Alias == "" || req.Target == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Missing required fields: holding_code, alias, target",
		})
	}

	if req.AliasType == "" {
		req.AliasType = "brand"
	}

	db, err := mypg.PgSqlFastConnect(req.HoldingCode)
	if err != nil {
		logger.Error("SearchAliasCreateHandler: connect PG: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Database connection failed",
		})
	}

	var id int
	err = db.QueryRow(
		"INSERT INTO search_aliases (alias, target, alias_type) VALUES ($1, $2, $3) RETURNING id",
		req.Alias, req.Target, req.AliasType,
	).Scan(&id)
	if err != nil {
		logger.Error("SearchAliasCreateHandler: insert: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Insert failed",
		})
	}

	logger.Info("SearchAlias created: %s → %s (id=%d)", req.Alias, req.Target, id)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": SearchAlias{
			ID:        id,
			Alias:     req.Alias,
			Target:    req.Target,
			AliasType: req.AliasType,
		},
	})
}

// SearchAliasDeleteHandler — DELETE /api/search/aliases/:id?holding_code=xxx
func SearchAliasDeleteHandler(c echo.Context) error {
	holdingCode := c.QueryParam("holding_code")
	if holdingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Missing required parameter: holding_code",
		})
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid id",
		})
	}

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("SearchAliasDeleteHandler: connect PG: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Database connection failed",
		})
	}

	result, err := db.Exec("DELETE FROM search_aliases WHERE id = $1", id)
	if err != nil {
		logger.Error("SearchAliasDeleteHandler: delete: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Delete failed",
		})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "Alias not found",
		})
	}

	logger.Info("SearchAlias deleted: id=%d", id)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Deleted",
	})
}
