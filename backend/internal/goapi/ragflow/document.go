package ragflow

// Document operations within a dataset.

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"strconv"
)

// DocumentMeta — minimal RAGFlow doc info we expose
type DocumentMeta struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	ChunkCount int    `json:"chunk_count"`
	TokenCount int    `json:"token_count"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
	Status     string `json:"status"` // RAGFlow status (UNSTART/RUNNING/SUCCESS/FAIL)
	Run        string `json:"run"`    // run state
}

type uploadResponse struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    []DocumentMeta `json:"data"`
}

type listDocsResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Total int            `json:"total"`
		Docs  []DocumentMeta `json:"docs"`
	} `json:"data"`
}

// UploadDocument streams a file into a dataset. Returns RAGFlow's doc IDs.
func (c *Client) UploadDocument(datasetID, filename string, content io.Reader) ([]DocumentMeta, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, content); err != nil {
		return nil, fmt.Errorf("copy file content: %w", err)
	}
	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("close multipart: %w", err)
	}

	var resp uploadResponse
	path := fmt.Sprintf("/api/v1/datasets/%s/documents", url.PathEscape(datasetID))
	if err := c.doMultipart("POST", path, mw.FormDataContentType(), &buf, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("ragflow upload: code=%d msg=%s", resp.Code, resp.Message)
	}
	return resp.Data, nil
}

// ParseDocument triggers async chunking/embedding for one or more docs.
// Returns immediately — call ListDocuments to poll status.
func (c *Client) ParseDocument(datasetID string, docIDs []string) error {
	body := map[string]any{"document_ids": docIDs}
	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	path := fmt.Sprintf("/api/v1/datasets/%s/chunks", url.PathEscape(datasetID))
	if err := c.doJSON("POST", path, body, &resp); err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("ragflow parse: code=%d msg=%s", resp.Code, resp.Message)
	}
	return nil
}

// ListDocuments returns all docs in a dataset (paginated under the hood).
func (c *Client) ListDocuments(datasetID string, page, pageSize int) ([]DocumentMeta, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 30
	}
	q := url.Values{}
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", strconv.Itoa(pageSize))

	var resp listDocsResponse
	path := fmt.Sprintf("/api/v1/datasets/%s/documents?%s", url.PathEscape(datasetID), q.Encode())
	if err := c.doJSON("GET", path, nil, &resp); err != nil {
		return nil, 0, err
	}
	if resp.Code != 0 {
		return nil, 0, fmt.Errorf("ragflow list docs: code=%d msg=%s", resp.Code, resp.Message)
	}
	return resp.Data.Docs, resp.Data.Total, nil
}

// DeleteDocument removes one or more documents from a dataset.
func (c *Client) DeleteDocument(datasetID string, docIDs []string) error {
	body := map[string]any{"ids": docIDs}
	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	path := fmt.Sprintf("/api/v1/datasets/%s/documents", url.PathEscape(datasetID))
	if err := c.doJSON("DELETE", path, body, &resp); err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("ragflow delete docs: code=%d msg=%s", resp.Code, resp.Message)
	}
	return nil
}
