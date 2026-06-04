package tools

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ==================== get_api_spec ====================

type APISpecRequest struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

type APISpecResponse struct {
	Endpoints []APISpecDetail `json:"endpoints"`
	Count     int             `json:"count"`
}

type APISpecDetail struct {
	APIEndpoint
	CurlExample string `json:"curl_example"`
	Notes       string `json:"notes,omitempty"`
}

func GetAPISpec(req APISpecRequest) (*APISpecResponse, error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	// ดึงทุก endpoints (no limit)
	catalog, err := GetAPICatalog(APICatalogRequest{Limit: 2000})
	if err != nil {
		return nil, err
	}

	pathLower := strings.ToLower(req.Path)
	methodUpper := strings.ToUpper(req.Method)

	// Priority: exact match first, then partial match
	var exactResults []APISpecDetail
	var partialResults []APISpecDetail
	for _, ep := range catalog.Endpoints {
		epPathLower := strings.ToLower(ep.Path)
		if methodUpper != "" && ep.Method != methodUpper {
			continue
		}

		detail := APISpecDetail{
			APIEndpoint: ep,
			CurlExample: generateCurl(ep),
			Notes:       generateNotes(ep),
		}

		if epPathLower == pathLower {
			exactResults = append(exactResults, detail)
		} else if strings.Contains(epPathLower, pathLower) {
			partialResults = append(partialResults, detail)
		}
	}

	// ถ้ามี exact match ใช้ exact เท่านั้น, ไม่งั้นใช้ partial
	results := exactResults
	if len(results) == 0 {
		results = partialResults
	}

	return &APISpecResponse{
		Endpoints: results,
		Count:     len(results),
	}, nil
}

func generateCurl(ep APIEndpoint) string {
	baseURL := "http://localhost:8888"
	var parts []string

	parts = append(parts, "curl")

	if ep.Method != "GET" {
		parts = append(parts, "-X", ep.Method)
	}

	// Headers
	if ep.AuthRequired {
		parts = append(parts, `-H "X-API-Key: YOUR_API_KEY"`)
	}

	// Body
	if ep.RequestBody != nil && ep.RequestBody.Example != nil {
		bodyJSON, _ := json.Marshal(ep.RequestBody.Example)
		parts = append(parts, `-H "Content-Type: application/json"`)
		parts = append(parts, fmt.Sprintf(`-d '%s'`, string(bodyJSON)))
	}

	parts = append(parts, fmt.Sprintf(`"%s%s"`, baseURL, ep.Path))

	return strings.Join(parts, " \\\n  ")
}

func generateNotes(ep APIEndpoint) string {
	var notes []string

	if ep.AuthRequired {
		notes = append(notes, "Requires API key (X-API-Key header)")
	}
	if ep.Source == "goapi" {
		notes = append(notes, "GoAPI endpoint (prefix: /goapi)")
	} else {
		notes = append(notes, "MainAPI endpoint (requires AccessToken auth)")
	}

	if ep.RequestBody != nil && ep.RequestBody.ContentType == "multipart/form-data" {
		notes = append(notes, "File upload: use multipart/form-data")
	}

	return strings.Join(notes, ". ")
}

// ==================== get_api_example ====================

type APIExampleRequest struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

type APIExampleResponse struct {
	Path              string                 `json:"path"`
	Method            string                 `json:"method"`
	Description       string                 `json:"description"`
	CurlCommand       string                 `json:"curl_command"`
	RequestExample    map[string]interface{} `json:"request_example,omitempty"`
	ResponseExample   map[string]interface{} `json:"response_example,omitempty"`
	DartExample       string                 `json:"dart_example"`
	TypeScriptExample string                 `json:"typescript_example"`
}

func GetAPIExample(req APIExampleRequest) (*APIExampleResponse, error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	// หา endpoint ที่ตรง
	catalog, err := GetAPICatalog(APICatalogRequest{Limit: 500})
	if err != nil {
		return nil, err
	}

	pathLower := strings.ToLower(req.Path)
	methodUpper := strings.ToUpper(req.Method)

	var matched *APIEndpoint
	for _, ep := range catalog.Endpoints {
		if strings.Contains(strings.ToLower(ep.Path), pathLower) {
			if methodUpper != "" && ep.Method != methodUpper {
				continue
			}
			epCopy := ep
			matched = &epCopy
			break
		}
	}

	if matched == nil {
		return nil, fmt.Errorf("ไม่พบ API ที่ตรงกับ path: %s", req.Path)
	}

	// ดึง hardcoded example ถ้ามี
	reqExample, respExample := getHardcodedExample(matched.Path, matched.Method)

	// ถ้าไม่มี hardcoded ให้ใช้จาก catalog
	if reqExample == nil && matched.RequestBody != nil {
		reqExample = matched.RequestBody.Example
	}
	if respExample == nil && matched.Response != nil {
		respExample = matched.Response.Example
	}

	return &APIExampleResponse{
		Path:              matched.Path,
		Method:            matched.Method,
		Description:       matched.Description,
		CurlCommand:       generateCurl(*matched),
		RequestExample:    reqExample,
		ResponseExample:   respExample,
		DartExample:       generateDartExample(*matched, reqExample),
		TypeScriptExample: generateTSExample(*matched, reqExample),
	}, nil
}

// ==================== Hardcoded Examples ====================

func getHardcodedExample(path, method string) (req map[string]interface{}, resp map[string]interface{}) {
	key := method + " " + path

	examples := map[string][2]map[string]interface{}{
		"POST /goapi/api/product/search": {
			{"holding_code": "SHOP001", "keyword": "น้ำตาล", "whcode": "WH01", "limit": 50},
			{"success": true, "products": []map[string]interface{}{
				{"barcode": "8850999220017", "name": "น้ำตาลทราย 1 kg", "price": 35.00, "balance": "10 ถุง", "unit_code": "BAG"},
			}},
		},
		"POST /goapi/api/product/barcode": {
			{"holding_code": "SHOP001", "barcode": "8850999220017"},
			{"success": true, "product": map[string]interface{}{"barcode": "8850999220017", "name": "น้ำตาลทราย", "price": 35.00}},
		},
		"POST /goapi/api/transaction/calculate": {
			{"holding_code": "SHOP001", "items": []map[string]interface{}{{"barcode": "001", "qty": 2, "price": 100}}, "tax_type": 1, "discount_amount": 50},
			{"success": true, "total_before_discount": 200, "discount_amount": 50, "total_after_discount": 150, "vat_amount": 9.81, "net_amount": 150},
		},
		"POST /goapi/api/report/sales/summary": {
			{"holding_code": "SHOP001", "from_date": "2025-01-01", "to_date": "2025-12-31"},
			{"success": true, "total_amount": 1500000, "total_cost": 1000000, "total_profit": 500000, "document_count": 5000},
		},
		"POST /goapi/api/approval/po-status/submit": {
			{"holding_code": "SHOP001", "docno": "PO-2025-001", "submitted_by": "EMP001"},
			{"success": true, "status": "pending", "message": "PO submitted for approval"},
		},
		"POST /goapi/api/approval/po-status/approve": {
			{"holding_code": "SHOP001", "docno": "PO-2025-001", "approved_by": "MGR001", "comment": "อนุมัติ"},
			{"success": true, "status": "approved"},
		},
		"POST /goapi/get": {
			{"holding_code": "SHOP001", "sql": "SELECT docno, docdate, totalamount FROM saleinvoice ORDER BY docdate DESC LIMIT 10"},
			{"success": true, "data": []map[string]interface{}{{"docno": "INV-001", "docdate": "2025-01-15", "total_amount": 15000}}},
		},
		"POST /goapi/genpdf": {
			{"holding_code": "SHOP001", "docno": "INV-001", "template": "invoice"},
			{"success": true, "url": "/s3/file/SHOP001/pdf/INV-001.pdf", "file_size": 125000},
		},
		"POST /goapi/mcp/invoke": {
			{"tool": "get_daily_sales", "params": map[string]interface{}{"date": "2025-01-15"}},
			{"success": true, "data": map[string]interface{}{"date": "2025-01-15", "total_amount": 45000, "document_count": 120}, "tool": "get_daily_sales"},
		},
		"POST /goapi/image/upload": {
			{"holding_code": "SHOP001", "category": "product"},
			{"success": true, "file_name": "abc123.jpg", "url": "/s3/file/SHOP001/images/abc123.jpg"},
		},
		"POST /goapi/api/lineoa/configs": {
			{"holding_code": "SHOP001"},
			{"success": true, "configs": []map[string]interface{}{{"channel_id": "123456", "channel_name": "My Shop LINE"}}},
		},
		"POST /goapi/api/v1/chatbot/chat-gemini": {
			{"holding_code": "SHOP001", "message": "ยอดขายวันนี้เท่าไร?"},
			{"success": true, "response": "ยอดขายวันนี้รวม 45,000 บาท จาก 120 บิล"},
		},
		"POST /goapi/api/v1/unified/query": {
			{"holding_code": "SHOP001", "query": "สินค้าขายดี 10 อันดับเดือนนี้"},
			{"success": true, "data": []map[string]interface{}{{"name": "น้ำตาลทราย", "total_qty": 500, "total_amount": 17500}}},
		},
	}

	if pair, ok := examples[key]; ok {
		return pair[0], pair[1]
	}
	return nil, nil
}

// ==================== Code Generation ====================

func generateDartExample(ep APIEndpoint, reqBody map[string]interface{}) string {
	baseURL := "http://localhost:8888"

	if ep.Method == "GET" {
		code := fmt.Sprintf(`import 'package:http/http.dart' as http;

final response = await http.get(
  Uri.parse('%s%s'),`, baseURL, ep.Path)
		if ep.AuthRequired {
			code += `
  headers: {'X-API-Key': apiKey},`
		}
		code += `
);
final data = jsonDecode(response.body);`
		return code
	}

	bodyStr := "{}"
	if reqBody != nil {
		b, _ := json.MarshalIndent(reqBody, "  ", "  ")
		bodyStr = string(b)
	}

	code := fmt.Sprintf(`import 'dart:convert';
import 'package:http/http.dart' as http;

final response = await http.post(
  Uri.parse('%s%s'),
  headers: {
    'Content-Type': 'application/json',`, baseURL, ep.Path)
	if ep.AuthRequired {
		code += `
    'X-API-Key': apiKey,`
	}
	code += fmt.Sprintf(`
  },
  body: jsonEncode(%s),
);
final data = jsonDecode(response.body);`, bodyStr)

	return code
}

func generateTSExample(ep APIEndpoint, reqBody map[string]interface{}) string {
	baseURL := "http://localhost:8888"

	if ep.Method == "GET" {
		code := fmt.Sprintf(`const response = await fetch('%s%s'`, baseURL, ep.Path)
		if ep.AuthRequired {
			code += `, {
  headers: { 'X-API-Key': apiKey },
}`
		}
		code += `);
const data = await response.json();`
		return code
	}

	bodyStr := "{}"
	if reqBody != nil {
		b, _ := json.MarshalIndent(reqBody, "  ", "  ")
		bodyStr = string(b)
	}

	code := fmt.Sprintf(`const response = await fetch('%s%s', {
  method: '%s',
  headers: {
    'Content-Type': 'application/json',`, baseURL, ep.Path, ep.Method)
	if ep.AuthRequired {
		code += `
    'X-API-Key': apiKey,`
	}
	code += fmt.Sprintf(`
  },
  body: JSON.stringify(%s),
});
const data = await response.json();`, bodyStr)

	return code
}
