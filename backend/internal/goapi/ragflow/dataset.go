package ragflow

// Dataset operations — multi-tenant per shop.
//
// Each shop gets a deterministic dataset name: bcacctshop<holdingCode>
// EnsureDataset is the main entry point — looks up by name, creates if missing,
// caches the dataset_id in-memory for fast subsequent calls.

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	defaultChunkMode = "naive"
)

// DatasetMeta — minimal RAGFlow dataset info we care about
type DatasetMeta struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CreateDatasetRequest matches RAGFlow's POST /api/v1/datasets schema.
//
// Note: we deliberately do NOT set language or embedding_model — RAGFlow
// will use the tenant defaults that were configured at provisioning time
// during provisioning. This avoids version
// drift in field names/values across RAGFlow releases.
//
// embedding_model format on newer RAGFlow versions is "<model>@<provider>"
// which depends on which provider+model the admin set up in the Web UI —
// hard to hardcode here, so leave it to tenant default.
type CreateDatasetRequest struct {
	Name           string `json:"name"`
	EmbeddingModel string `json:"embedding_model,omitempty"`
	ChunkMethod    string `json:"chunk_method,omitempty"`
}

type createDatasetResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    *DatasetMeta `json:"data"`
}

type listDatasetsResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    []DatasetMeta `json:"data"`
}

// ====== In-memory cache: holdingCode → datasetID ======

var (
	datasetCache   = map[string]cachedDataset{}
	datasetCacheMu sync.RWMutex
)

type cachedDataset struct {
	id       string
	cachedAt time.Time
}

const datasetCacheTTL = 30 * time.Minute

func getCachedDataset(holdingCode string) string {
	datasetCacheMu.RLock()
	defer datasetCacheMu.RUnlock()
	v, ok := datasetCache[holdingCode]
	if !ok {
		return ""
	}
	if time.Since(v.cachedAt) > datasetCacheTTL {
		return ""
	}
	return v.id
}

func putCachedDataset(holdingCode, datasetID string) {
	datasetCacheMu.Lock()
	defer datasetCacheMu.Unlock()
	datasetCache[holdingCode] = cachedDataset{id: datasetID, cachedAt: time.Now()}
}

// InvalidateDatasetCache clears a single shop's cached dataset id (call after delete)
func InvalidateDatasetCache(holdingCode string) {
	datasetCacheMu.Lock()
	defer datasetCacheMu.Unlock()
	delete(datasetCache, holdingCode)
}

// ====== Public API ======

// DatasetNameFromShop returns the deterministic dataset name for a shop.
// Allows reverse lookup without needing a separate mapping table.
func DatasetNameFromShop(holdingCode string) string {
	// Lowercase + replace any whitespace — RAGFlow allows alphanumeric/underscore
	cleaned := strings.ToLower(strings.TrimSpace(holdingCode))
	cleaned = strings.ReplaceAll(cleaned, " ", "_")
	cleaned = strings.ReplaceAll(cleaned, "-", "_")
	return "bcacctshop" + cleaned
}

// EnsureDataset returns the dataset ID for a shop, creating one if it doesn't exist.
// Cached in-memory after first lookup.
func (c *Client) EnsureDataset(holdingCode string) (string, error) {
	if holdingCode == "" {
		return "", fmt.Errorf("holdingcode is required")
	}
	if id := getCachedDataset(holdingCode); id != "" {
		return id, nil
	}

	name := DatasetNameFromShop(holdingCode)

	// Try lookup first
	id, err := c.FindDatasetByName(name)
	if err != nil {
		return "", fmt.Errorf("lookup dataset: %w", err)
	}
	if id != "" {
		putCachedDataset(holdingCode, id)
		return id, nil
	}

	// Not found — create
	created, err := c.CreateDataset(name)
	if err != nil {
		return "", fmt.Errorf("create dataset: %w", err)
	}
	putCachedDataset(holdingCode, created.ID)
	return created.ID, nil
}

// FindDatasetByName lists datasets filtered by name. Returns "" if not found.
// Useful as a lightweight health probe (verifies network + auth + response shape).
func (c *Client) FindDatasetByName(name string) (string, error) {
	q := url.Values{}
	q.Set("name", name)
	q.Set("page_size", "1")
	var resp listDatasetsResponse
	if err := c.doJSON("GET", "/api/v1/datasets?"+q.Encode(), nil, &resp); err != nil {
		return "", err
	}
	for _, d := range resp.Data {
		if strings.EqualFold(d.Name, name) {
			return d.ID, nil
		}
	}
	return "", nil
}

// CreateDataset creates a new dataset with the given name and BC defaults.
// Embedding model + language are inherited from the tenant defaults set in
// RAGFlow Web UI / via REST during provisioning.
func (c *Client) CreateDataset(name string) (*DatasetMeta, error) {
	req := CreateDatasetRequest{
		Name:        name,
		ChunkMethod: defaultChunkMode,
	}
	var resp createDatasetResponse
	if err := c.doJSON("POST", "/api/v1/datasets", req, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 || resp.Data == nil {
		return nil, fmt.Errorf("ragflow create dataset failed: code=%d msg=%s", resp.Code, resp.Message)
	}
	return resp.Data, nil
}

// DeleteDataset removes a dataset by ID. Used for cleanup; not exposed via UI.
func (c *Client) DeleteDataset(datasetID string) error {
	body := map[string]any{"ids": []string{datasetID}}
	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := c.doJSON("DELETE", "/api/v1/datasets", body, &resp); err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("ragflow delete dataset: code=%d msg=%s", resp.Code, resp.Message)
	}
	return nil
}
