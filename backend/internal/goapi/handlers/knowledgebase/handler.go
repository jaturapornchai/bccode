package knowledgebase

// REST handlers for the BC Knowledge Base UI.
//
// These handlers proxy to RAGFlow for storage/parsing/retrieval but maintain
// BC-specific metadata (branch, status, schedule) in MongoDB. The response
// shape matches what the existing Flutter UI expects (DocumentModel).
//
// Routes (mounted under /api/v1/kb in bootstrap.go):
//   POST /list             — list docs for a shop (with branch filter)
//   POST /upload           — upload one doc (base64 in JSON, like the old API)
//   POST /delete           — delete by filename
//   POST /update-status    — toggle on/off
//   POST /update-allday    — toggle 24/7 vs scheduled
//   POST /update-schedule  — set start/end datetime

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/ragflow"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
)

// ====== request structs ======

type listRequest struct {
	ShopID   string `json:"shop_id"`
	BranchID string `json:"branch_id"`
	Limit    int    `json:"limit"`
	Skip     int    `json:"skip"`
}

type uploadRequest struct {
	ShopID      string   `json:"shop_id"`
	BranchID    string   `json:"branch_id"`
	Filename    string   `json:"filename"`
	ContentType string   `json:"content_type"`
	Content     string   `json:"content"` // base64
	UploadedBy  string   `json:"uploaded_by"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Status      bool     `json:"status"`
	AllDay      bool     `json:"all_day"`
}

type deleteRequest struct {
	ShopID   string `json:"shop_id"`
	BranchID string `json:"branch_id"`
	Filename string `json:"filename"`
}

type updateStatusRequest struct {
	ShopID   string `json:"shop_id"`
	BranchID string `json:"branch_id"`
	Filename string `json:"filename"`
	Status   bool   `json:"status"`
}

type updateAllDayRequest struct {
	ShopID   string `json:"shop_id"`
	BranchID string `json:"branch_id"`
	Filename string `json:"filename"`
	AllDay   bool   `json:"all_day"`
}

type updateScheduleRequest struct {
	ShopID        string  `json:"shop_id"`
	BranchID      string  `json:"branch_id"`
	Filename      string  `json:"filename"`
	StartDateTime *string `json:"start_date_time"`
	EndDateTime   *string `json:"end_date_time"`
}

// ====== response shape — matches Flutter DocumentModel ======

type documentResponse struct {
	ShopID        string   `json:"shop_id"`
	BranchID      string   `json:"branch_id"`
	Filename      string   `json:"filename"`
	ContentType   string   `json:"content_type"`
	Size          int64    `json:"size"`
	UploadedAt    string   `json:"uploaded_at"`
	UploadedBy    string   `json:"uploaded_by"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	Version       int      `json:"version"`
	Status        bool     `json:"status"`
	AllDay        bool     `json:"all_day"`
	StartDateTime *string  `json:"start_date_time,omitempty"`
	EndDateTime   *string  `json:"end_date_time,omitempty"`
}

func toResponse(meta KBDocMeta) documentResponse {
	tags := meta.Tags
	if tags == nil {
		tags = []string{}
	}
	return documentResponse{
		ShopID:        meta.ShopID,
		BranchID:      meta.BranchID,
		Filename:      meta.Filename,
		ContentType:   meta.ContentType,
		Size:          meta.Size,
		UploadedAt:    meta.UploadedAt.UTC().Format(time.RFC3339),
		UploadedBy:    meta.UploadedBy,
		Description:   meta.Description,
		Tags:          tags,
		Version:       meta.Version,
		Status:        meta.Status,
		AllDay:        meta.AllDay,
		StartDateTime: meta.StartDateTime,
		EndDateTime:   meta.EndDateTime,
	}
}

func errorJSON(c echo.Context, status int, msg string) error {
	return c.JSON(status, map[string]any{"success": false, "message": msg})
}

// ====== handlers ======

// ListDocuments — POST /api/v1/kb/list
func ListDocuments(c echo.Context) error {
	var req listRequest
	if err := c.Bind(&req); err != nil {
		return errorJSON(c, http.StatusBadRequest, "invalid request: "+err.Error())
	}
	if req.ShopID == "" {
		return errorJSON(c, http.StatusBadRequest, "shop_id required")
	}
	metas, err := GetMetaByShop(req.ShopID, req.BranchID)
	if err != nil {
		logger.Error("[KB] list meta failed: %v", err)
		return errorJSON(c, http.StatusInternalServerError, err.Error())
	}
	docs := make([]documentResponse, 0, len(metas))
	for _, m := range metas {
		docs = append(docs, toResponse(m))
	}
	return c.JSON(http.StatusOK, map[string]any{
		"success":   true,
		"documents": docs,
	})
}

// UploadDocument — POST /api/v1/kb/upload
//
// Decodes base64 content, uploads to RAGFlow, triggers async parse,
// and saves BC metadata in Mongo.
func UploadDocument(c echo.Context) error {
	var req uploadRequest
	if err := c.Bind(&req); err != nil {
		return errorJSON(c, http.StatusBadRequest, "invalid request: "+err.Error())
	}
	if req.ShopID == "" || req.Filename == "" || req.Content == "" {
		return errorJSON(c, http.StatusBadRequest, "shop_id, filename, content required")
	}
	if req.BranchID == "" {
		req.BranchID = "*"
	}

	rawBytes, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil {
		return errorJSON(c, http.StatusBadRequest, "invalid base64 content: "+err.Error())
	}

	client := ragflow.GetClient()
	if !client.IsConfigured() {
		return errorJSON(c, http.StatusServiceUnavailable, "RAGFlow not configured (RAGFLOW_API_KEY missing)")
	}

	datasetID, err := client.EnsureDataset(req.ShopID)
	if err != nil {
		logger.Error("[KB] EnsureDataset failed shop=%s: %v", req.ShopID, err)
		return errorJSON(c, http.StatusInternalServerError, "ensure dataset: "+err.Error())
	}

	docs, err := client.UploadDocument(datasetID, req.Filename, bytes.NewReader(rawBytes))
	if err != nil {
		logger.Error("[KB] UploadDocument failed shop=%s file=%s: %v", req.ShopID, req.Filename, err)
		return errorJSON(c, http.StatusInternalServerError, "upload to ragflow: "+err.Error())
	}
	if len(docs) == 0 {
		return errorJSON(c, http.StatusInternalServerError, "ragflow returned no doc id")
	}
	uploaded := docs[0]

	// Fire-and-forget parse trigger so chunks are searchable.
	// Errors here aren't fatal — the doc is saved and parse will retry on next access.
	go func(dsID, docID string) {
		if perr := client.ParseDocument(dsID, []string{docID}); perr != nil {
			logger.Warn("[KB] ParseDocument trigger failed doc=%s: %v", docID, perr)
		} else {
			logger.Info("[KB] ParseDocument triggered doc=%s", docID)
		}
	}(datasetID, uploaded.ID)

	// Save BC metadata
	meta := &KBDocMeta{
		ShopID:       req.ShopID,
		RagflowDocID: uploaded.ID,
		Filename:     req.Filename,
		BranchID:     req.BranchID,
		ContentType:  req.ContentType,
		Size:         int64(len(rawBytes)),
		UploadedBy:   req.UploadedBy,
		Description:  req.Description,
		Tags:         req.Tags,
		Status:       req.Status,
		AllDay:       req.AllDay,
		Version:      1,
		UploadedAt:   time.Now(),
	}
	if err := UpsertMeta(meta); err != nil {
		logger.Error("[KB] save meta failed: %v", err)
		// Don't fail the request — file is in RAGFlow already.
	}

	logger.Info("[KB] uploaded shop=%s file=%s ragflow_doc=%s size=%d",
		req.ShopID, req.Filename, uploaded.ID, len(rawBytes))

	return c.JSON(http.StatusOK, map[string]any{
		"success":  true,
		"message":  "uploaded",
		"document": toResponse(*meta),
	})
}

// DeleteDocument — POST /api/v1/kb/delete
func DeleteDocument(c echo.Context) error {
	var req deleteRequest
	if err := c.Bind(&req); err != nil {
		return errorJSON(c, http.StatusBadRequest, "invalid request: "+err.Error())
	}
	if req.ShopID == "" || req.Filename == "" {
		return errorJSON(c, http.StatusBadRequest, "shop_id, filename required")
	}

	meta, err := FindByFilename(req.ShopID, req.Filename)
	if err != nil {
		return errorJSON(c, http.StatusInternalServerError, err.Error())
	}
	if meta == nil {
		return errorJSON(c, http.StatusNotFound, "document not found")
	}

	client := ragflow.GetClient()
	if client.IsConfigured() {
		datasetID, err := client.EnsureDataset(req.ShopID)
		if err == nil && meta.RagflowDocID != "" {
			if derr := client.DeleteDocument(datasetID, []string{meta.RagflowDocID}); derr != nil {
				logger.Warn("[KB] ragflow delete failed (continuing with meta delete): %v", derr)
			}
		}
	}

	if err := DeleteMeta(req.ShopID, meta.RagflowDocID); err != nil {
		return errorJSON(c, http.StatusInternalServerError, err.Error())
	}

	logger.Info("[KB] deleted shop=%s file=%s", req.ShopID, req.Filename)
	return c.JSON(http.StatusOK, map[string]any{"success": true})
}

// UpdateStatus — POST /api/v1/kb/update-status
func UpdateStatus(c echo.Context) error {
	var req updateStatusRequest
	if err := c.Bind(&req); err != nil {
		return errorJSON(c, http.StatusBadRequest, "invalid request: "+err.Error())
	}
	if req.ShopID == "" || req.Filename == "" {
		return errorJSON(c, http.StatusBadRequest, "shop_id, filename required")
	}
	if err := UpdateFields(req.ShopID, req.Filename, bson.M{"status": req.Status}); err != nil {
		return errorJSON(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"success": true})
}

// UpdateAllDay — POST /api/v1/kb/update-allday
func UpdateAllDay(c echo.Context) error {
	var req updateAllDayRequest
	if err := c.Bind(&req); err != nil {
		return errorJSON(c, http.StatusBadRequest, "invalid request: "+err.Error())
	}
	if req.ShopID == "" || req.Filename == "" {
		return errorJSON(c, http.StatusBadRequest, "shop_id, filename required")
	}
	if err := UpdateFields(req.ShopID, req.Filename, bson.M{"allday": req.AllDay}); err != nil {
		return errorJSON(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"success": true})
}

// UpdateSchedule — POST /api/v1/kb/update-schedule
func UpdateSchedule(c echo.Context) error {
	var req updateScheduleRequest
	if err := c.Bind(&req); err != nil {
		return errorJSON(c, http.StatusBadRequest, "invalid request: "+err.Error())
	}
	if req.ShopID == "" || req.Filename == "" {
		return errorJSON(c, http.StatusBadRequest, "shop_id, filename required")
	}
	set := bson.M{}
	if req.StartDateTime != nil && strings.TrimSpace(*req.StartDateTime) != "" {
		set["startdatetime"] = *req.StartDateTime
	} else {
		set["startdatetime"] = nil
	}
	if req.EndDateTime != nil && strings.TrimSpace(*req.EndDateTime) != "" {
		set["enddatetime"] = *req.EndDateTime
	} else {
		set["enddatetime"] = nil
	}
	if err := UpdateFields(req.ShopID, req.Filename, set); err != nil {
		return errorJSON(c, http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"success": true})
}

// Health — GET /api/v1/kb/health
//
// Reports whether RAGFlow is reachable + configured. Used by the Flutter KB
// screen to show a banner when admin hasn't set RAGFLOW_API_KEY yet, so
// uploads/queries don't fail mysteriously.
func Health(c echo.Context) error {
	client := ragflow.GetClient()
	resp := map[string]any{
		"configured": client.IsConfigured(),
		"base_url":   client.BaseURL,
	}

	if !client.IsConfigured() {
		resp["status"] = "not_configured"
		resp["message"] = "RAGFLOW_API_KEY ยังไม่ถูกตั้งค่าใน mainapi — Knowledge Base ยังใช้งานไม่ได้"
		return c.JSON(http.StatusOK, resp)
	}

	// Probe by listing datasets (lightweight call) — verifies network + auth
	start := time.Now()
	if _, err := client.FindDatasetByName("__health_probe__"); err != nil {
		resp["status"] = "unreachable"
		resp["message"] = "ไม่สามารถเชื่อม RAGFlow ได้: " + err.Error()
		resp["latency_ms"] = time.Since(start).Milliseconds()
		return c.JSON(http.StatusOK, resp)
	}
	resp["status"] = "ok"
	resp["message"] = "RAGFlow online"
	resp["latency_ms"] = time.Since(start).Milliseconds()
	return c.JSON(http.StatusOK, resp)
}

// Query — POST /api/v1/kb/query
//
// Direct retrieval endpoint — bypasses the AI agent and returns raw chunks
// from the shop's RAGFlow dataset. Useful for:
//   - Admin previews the KB ("does the doc index actually contain this answer?")
//   - Smoke-testing the full Go→RAGFlow path without depending on a tool-calling LLM
//   - Frontend "search KB" feature (separate from chat)
type queryRequest struct {
	ShopID string `json:"shop_id"`
	Query  string `json:"query"`
	TopK   int    `json:"top_k"`
}

func Query(c echo.Context) error {
	var req queryRequest
	if err := c.Bind(&req); err != nil {
		return errorJSON(c, http.StatusBadRequest, "invalid request: "+err.Error())
	}
	if req.ShopID == "" || strings.TrimSpace(req.Query) == "" {
		return errorJSON(c, http.StatusBadRequest, "shop_id and query are required")
	}
	if req.TopK <= 0 || req.TopK > 50 {
		req.TopK = 8
	}

	client := ragflow.GetClient()
	if !client.IsConfigured() {
		return errorJSON(c, http.StatusServiceUnavailable, "RAGFLOW_API_KEY not configured")
	}

	datasetID, err := client.EnsureDataset(req.ShopID)
	if err != nil {
		logger.Error("[KB.Query] EnsureDataset shop=%s: %v", req.ShopID, err)
		return errorJSON(c, http.StatusInternalServerError, "ensure dataset: "+err.Error())
	}

	chunks, err := client.Retrieve(req.Query, []string{datasetID}, req.TopK)
	if err != nil {
		logger.Error("[KB.Query] Retrieve shop=%s: %v", req.ShopID, err)
		return errorJSON(c, http.StatusInternalServerError, "retrieve: "+err.Error())
	}

	results := make([]map[string]any, 0, len(chunks))
	for i, ch := range chunks {
		results = append(results, map[string]any{
			"rank":       i + 1,
			"content":    ch.Content,
			"doc_name":   ch.DocumentKeyword,
			"similarity": ch.Similarity,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":    true,
		"shop_id":    req.ShopID,
		"dataset_id": datasetID,
		"query":      req.Query,
		"total":      len(chunks),
		"chunks":     results,
	})
}

// ViewDocument — GET /api/v1/kb/view?shop_id=X&doc_id=Y
// Returns an HTML page rendering all chunks of the specified document.
// Used as clickable citation target in AI chatbot responses.
func ViewDocument(c echo.Context) error {
	shopID := strings.TrimSpace(c.QueryParam("shop_id"))
	docID := strings.TrimSpace(c.QueryParam("doc_id"))
	if shopID == "" || docID == "" {
		return c.HTML(http.StatusBadRequest, "<h3>Missing shop_id or doc_id</h3>")
	}

	client := ragflow.GetClient()
	if !client.IsConfigured() {
		return c.HTML(http.StatusServiceUnavailable, "<h3>KB service not configured</h3>")
	}

	datasetID, err := client.EnsureDataset(shopID)
	if err != nil {
		logger.Error("[KB.View] EnsureDataset shop=%s: %v", shopID, err)
		return c.HTML(http.StatusInternalServerError, "<h3>Failed to load dataset</h3>")
	}

	// Use empty query + doc_id filter via Retrieve — request large topK then filter client-side.
	// RAGFlow /retrieval accepts document_ids; we reuse Retrieve with a generic query.
	chunks, err := client.RetrieveByDocument(datasetID, docID, 50)
	if err != nil {
		logger.Error("[KB.View] Retrieve shop=%s doc=%s: %v", shopID, docID, err)
		return c.HTML(http.StatusInternalServerError, "<h3>Failed to load document</h3>")
	}

	var b strings.Builder
	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"><title>KB Document</title>`)
	b.WriteString(`<style>body{font-family:system-ui,sans-serif;padding:16px;line-height:1.6;color:#222}h2{color:#1976d2;border-bottom:2px solid #1976d2;padding-bottom:8px}.chunk{background:#f5f5f5;border-left:4px solid #1976d2;padding:12px;margin:12px 0;border-radius:4px}.meta{color:#666;font-size:12px;margin-bottom:6px}</style>`)
	b.WriteString(`</head><body>`)
	docName := "เอกสาร"
	if len(chunks) > 0 {
		docName = chunks[0].DocumentKeyword
	}
	fmt.Fprintf(&b, `<h2>📚 %s</h2>`, echoEscape(docName))
	if len(chunks) == 0 {
		b.WriteString(`<p>ไม่พบเนื้อหาในเอกสารนี้</p>`)
	} else {
		fmt.Fprintf(&b, `<p>พบ %d chunks</p>`, len(chunks))
		for i, ch := range chunks {
			fmt.Fprintf(&b, `<div class="chunk"><div class="meta">Chunk #%d — similarity: %.2f</div>%s</div>`,
				i+1, ch.Similarity, echoEscape(ch.Content))
		}
	}
	b.WriteString(`</body></html>`)
	return c.HTML(http.StatusOK, b.String())
}

// echoEscape — minimal HTML escape for untrusted content
func echoEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;")
	return r.Replace(s)
}

// RegisterRoutes mounts all KB endpoints under the given group.
// Called from bootstrap.go: knowledgebase.RegisterRoutes(g.Group("/api/v1/kb"))
func RegisterRoutes(g *echo.Group) {
	g.GET("/health", Health)
	g.GET("/view", ViewDocument)
	g.POST("/list", ListDocuments)
	g.POST("/upload", UploadDocument)
	g.POST("/delete", DeleteDocument)
	g.POST("/update-status", UpdateStatus)
	g.POST("/update-allday", UpdateAllDay)
	g.POST("/update-schedule", UpdateSchedule)
	g.POST("/query", Query)
}
