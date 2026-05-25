package ragflow

// Retrieval API — used by the AI agent's query_knowledge_base tool.
//
// Unlike the chat assistant, retrieval returns RAW chunks with similarity scores
// — the agent's LLM does the synthesis itself. This keeps responses fast,
// avoids double-LLM cost, and gives the agent flexibility on how to present
// the data.

import (
	"fmt"
	"strings"
)

// RetrievalRequest matches RAGFlow's POST /api/v1/retrieval schema.
type RetrievalRequest struct {
	Question         string   `json:"question"`
	DatasetIDs       []string `json:"dataset_ids"`
	DocumentIDs      []string `json:"document_ids,omitempty"`
	PageSize         int      `json:"page_size,omitempty"` // chunks per result (default 30)
	SimilarityThresh float64  `json:"similarity_threshold,omitempty"`
	VectorWeight     float64  `json:"vector_similarity_weight,omitempty"`
	TopK             int      `json:"top_k,omitempty"`
	RerankID         string   `json:"rerank_id,omitempty"`
	KeywordSearch    bool     `json:"keyword,omitempty"`
	Highlight        bool     `json:"highlight,omitempty"`
}

// RetrievalChunk — one matched piece of a document
type RetrievalChunk struct {
	ID                string   `json:"id"`
	Content           string   `json:"content"`
	ContentLTKS       string   `json:"content_ltks"`
	DocumentID        string   `json:"document_id"`
	DocumentKeyword   string   `json:"document_keyword"`
	Highlight         string   `json:"highlight"`
	Img               string   `json:"img_id"`
	ImportantKeywords []string `json:"important_keywords"`
	KbID              string   `json:"kb_id"`
	Similarity        float64  `json:"similarity"`
	TermSimilarity    float64  `json:"term_similarity"`
	VectorSimilarity  float64  `json:"vector_similarity"`
}

type retrievalResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Chunks  []RetrievalChunk `json:"chunks"`
		DocAggs []map[string]any `json:"doc_aggs"`
		Total   int              `json:"total"`
	} `json:"data"`
}

// Retrieve searches one or more datasets for chunks matching the question.
// Returns top-K chunks ranked by similarity.
func (c *Client) Retrieve(question string, datasetIDs []string, topK int) ([]RetrievalChunk, error) {
	if strings.TrimSpace(question) == "" {
		return nil, fmt.Errorf("question is required")
	}
	if len(datasetIDs) == 0 {
		return nil, fmt.Errorf("at least one dataset_id is required")
	}
	if topK <= 0 {
		topK = 8
	}

	// Workaround for RAGFlow's Thai tokenizer limitation:
	// `qryr.question()` returns None for pure Thai input → AssertionError in
	// the ES query builder. We must inject English tokens so the tokenizer
	// produces a valid MatchTextExpr. The injected tokens also need to actually
	// match document text — RAGFlow's hybrid scoring filters out chunks where
	// BM25 contribution is zero, so a generic word like "search" returns nothing.
	//
	// Since every document in BC Account's KB is a BC Account document, prepending
	// "BC Account document" gives near-universal BM25 hits while the multilingual
	// embedding (gemma) handles the actual Thai semantic match via vector similarity.
	//
	// We deliberately do NOT use `keyword=true` (which would invoke the chat model
	// to extract keywords) because RAGFlow's chat-model integration mangles
	// Thai UTF-8 chars into `?` before sending to the model, returning 0 results.
	q := question
	if !containsASCIILetter(question) {
		q = "BC Account document " + question
	}

	req := RetrievalRequest{
		Question:         q,
		DatasetIDs:       datasetIDs,
		PageSize:         topK,
		SimilarityThresh: 0.05, // low threshold so semantic matches survive even when term-overlap is small
		VectorWeight:     0.85, // heavy on vector — multilingual embedding handles Thai↔English
		TopK:             1024, // RAGFlow internal candidate pool — must be reasonably large
		KeywordSearch:    false,
		Highlight:        false,
	}

	var resp retrievalResponse
	if err := c.doJSON("POST", "/api/v1/retrieval", req, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("ragflow retrieval: code=%d msg=%s", resp.Code, resp.Message)
	}
	return resp.Data.Chunks, nil
}

// RetrieveByDocument returns chunks for a specific document in the dataset.
// Uses a generic "BC Account document" query to satisfy RAGFlow's required
// question param, then filters by document_id server-side.
func (c *Client) RetrieveByDocument(datasetID, documentID string, topK int) ([]RetrievalChunk, error) {
	if datasetID == "" || documentID == "" {
		return nil, fmt.Errorf("datasetID and documentID are required")
	}
	if topK <= 0 {
		topK = 50
	}
	req := RetrievalRequest{
		Question:         "BC Account document",
		DatasetIDs:       []string{datasetID},
		DocumentIDs:      []string{documentID},
		PageSize:         topK,
		SimilarityThresh: 0.0,
		VectorWeight:     0.5,
		TopK:             1024,
		KeywordSearch:    false,
		Highlight:        false,
	}
	var resp retrievalResponse
	if err := c.doJSON("POST", "/api/v1/retrieval", req, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("ragflow retrieval: code=%d msg=%s", resp.Code, resp.Message)
	}
	return resp.Data.Chunks, nil
}

// containsASCIILetter reports whether s has at least one A-Z/a-z character.
// Used to decide whether to inject an English bridge word for RAGFlow's
// Thai-tokenizer workaround.
func containsASCIILetter(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			return true
		}
	}
	return false
}
