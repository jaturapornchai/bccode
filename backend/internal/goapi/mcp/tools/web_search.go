package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"time"
)

// WebSearchResponse ผลลัพธ์จากการค้นหาเว็บ
type WebSearchResponse struct {
	Query       string            `json:"query"`
	Results     []WebSearchResult `json:"results"`
	Count       int               `json:"count"`
	Source      string            `json:"source"`
	GeneratedAt time.Time         `json:"generated_at"`
}

// WebSearchResult รายการผลลัพธ์แต่ละรายการ
type WebSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// WebSearch ค้นหาข้อมูลจาก internet ผ่าน DuckDuckGo
func WebSearch(ctx context.Context, query string, limit int) (*WebSearchResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if limit <= 0 {
		limit = 5
	}
	if limit > 10 {
		limit = 10
	}

	logger.Info("[MCP WebSearch] query=%s, limit=%d", query, limit)

	// 1. ลอง DuckDuckGo Instant Answer API ก่อน (structured data)
	results, err := searchDDGInstant(ctx, query)
	if err == nil && len(results) > 0 {
		if len(results) > limit {
			results = results[:limit]
		}
		return &WebSearchResponse{
			Query:       query,
			Results:     results,
			Count:       len(results),
			Source:      "duckduckgo_instant",
			GeneratedAt: time.Now(),
		}, nil
	}

	// 2. Fallback: DuckDuckGo HTML search
	results, err = searchDDGHTML(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("web search failed: %w", err)
	}

	return &WebSearchResponse{
		Query:       query,
		Results:     results,
		Count:       len(results),
		Source:      "duckduckgo_html",
		GeneratedAt: time.Now(),
	}, nil
}

// searchDDGInstant ใช้ DuckDuckGo Instant Answer API
func searchDDGInstant(ctx context.Context, query string) ([]WebSearchResult, error) {
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1",
		url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "BCAccount-AI-Agent/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ddgResp struct {
		Abstract       string `json:"Abstract"`
		AbstractURL    string `json:"AbstractURL"`
		AbstractSource string `json:"AbstractSource"`
		RelatedTopics  []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
	}
	if err := json.Unmarshal(body, &ddgResp); err != nil {
		return nil, err
	}

	var results []WebSearchResult

	// Abstract (main answer)
	if ddgResp.Abstract != "" {
		results = append(results, WebSearchResult{
			Title:   ddgResp.AbstractSource,
			URL:     ddgResp.AbstractURL,
			Snippet: ddgResp.Abstract,
		})
	}

	// Related topics
	for _, topic := range ddgResp.RelatedTopics {
		if topic.Text != "" && topic.FirstURL != "" {
			results = append(results, WebSearchResult{
				Title:   extractTitle(topic.Text),
				URL:     topic.FirstURL,
				Snippet: topic.Text,
			})
		}
	}

	return results, nil
}

// searchDDGHTML ค้นหาผ่าน DuckDuckGo HTML (no JS required)
func searchDDGHTML(ctx context.Context, query string, limit int) ([]WebSearchResult, error) {
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; BCAccount-AI-Agent/1.0)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)
	return parseDDGHTML(html, limit), nil
}

// parseDDGHTML — parse DuckDuckGo HTML results
func parseDDGHTML(html string, limit int) []WebSearchResult {
	var results []WebSearchResult

	// Pattern: <a class="result__a" href="...">Title</a>
	// + <a class="result__snippet" href="...">Snippet</a>
	linkRe := regexp.MustCompile(`<a[^>]+class="result__a"[^>]+href="([^"]*)"[^>]*>(.*?)</a>`)
	snippetRe := regexp.MustCompile(`<a[^>]+class="result__snippet"[^>]*>(.*?)</a>`)

	links := linkRe.FindAllStringSubmatch(html, limit*2)
	snippets := snippetRe.FindAllStringSubmatch(html, limit*2)

	for i, match := range links {
		if len(results) >= limit {
			break
		}

		rawURL := match[1]
		title := stripHTML(match[2])

		// DuckDuckGo wraps URLs in redirects — extract actual URL
		actualURL := extractDDGURL(rawURL)
		if actualURL == "" {
			actualURL = rawURL
		}

		snippet := ""
		if i < len(snippets) {
			snippet = stripHTML(snippets[i][1])
		}

		if title != "" && actualURL != "" {
			results = append(results, WebSearchResult{
				Title:   title,
				URL:     actualURL,
				Snippet: snippet,
			})
		}
	}

	return results
}

// extractDDGURL — extract actual URL from DDG redirect
func extractDDGURL(ddgURL string) string {
	// Pattern: //duckduckgo.com/l/?uddg=ENCODED_URL&...
	if strings.Contains(ddgURL, "uddg=") {
		u, err := url.Parse(ddgURL)
		if err != nil {
			return ddgURL
		}
		uddg := u.Query().Get("uddg")
		if uddg != "" {
			decoded, err := url.QueryUnescape(uddg)
			if err == nil {
				return decoded
			}
			return uddg
		}
	}
	return ddgURL
}

// extractTitle — extract first meaningful part of text as title
func extractTitle(text string) string {
	if idx := strings.Index(text, " - "); idx > 0 && idx < 100 {
		return text[:idx]
	}
	if len(text) > 80 {
		return text[:80] + "..."
	}
	return text
}

// stripHTML — ลบ HTML tags
func stripHTML(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	clean := re.ReplaceAllString(s, "")
	clean = strings.ReplaceAll(clean, "&amp;", "&")
	clean = strings.ReplaceAll(clean, "&lt;", "<")
	clean = strings.ReplaceAll(clean, "&gt;", ">")
	clean = strings.ReplaceAll(clean, "&quot;", "\"")
	clean = strings.ReplaceAll(clean, "&#x27;", "'")
	clean = strings.ReplaceAll(clean, "&nbsp;", " ")
	return strings.TrimSpace(clean)
}
