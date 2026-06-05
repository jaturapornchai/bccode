package tools

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ==================== getapispec ====================

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
	CurlExample string `json:"curlexample"`
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

// ==================== getapiexample ====================

type APIExampleRequest struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

type APIExampleResponse struct {
	Path              string                 `json:"path"`
	Method            string                 `json:"method"`
	Description       string                 `json:"description"`
	CurlCommand       string                 `json:"curlcommand"`
	RequestExample    map[string]interface{} `json:"requestexample,omitempty"`
	ResponseExample   map[string]interface{} `json:"responseexample,omitempty"`
	DartExample       string                 `json:"dartexample"`
	TypeScriptExample string                 `json:"typescriptexample"`
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
			{"holdingcode": "SHOP001", "keyword": "น้ำตาล", "whcode": "WH01", "limit": 50},
			{"success": true, "products": []map[string]interface{}{
				{"barcode": "8850999220017", "name": "น้ำตาลทราย 1 kg", "price": 35.00, "balance": "10 ถุง", "unitcode": "BAG"},
			}},
		},
		"POST /goapi/api/product/barcode": {
			{"holdingcode": "SHOP001", "barcode": "8850999220017"},
			{"success": true, "product": map[string]interface{}{"barcode": "8850999220017", "name": "น้ำตาลทราย", "price": 35.00}},
		},
		"POST /goapi/api/transaction/calculate": {
			{"holdingcode": "SHOP001", "items": []map[string]interface{}{{"barcode": "001", "qty": 2, "price": 100}}, "taxtype": 1, "discountamount": 50},
			{"success": true, "totalbeforediscount": 200, "discountamount": 50, "totalafterdiscount": 150, "vatamount": 9.81, "netamount": 150},
		},
		"POST /goapi/api/report/sales/summary": {
			{"holdingcode": "SHOP001", "fromdate": "2025-01-01", "todate": "2025-12-31"},
			{"success": true, "totalamount": 1500000, "totalcost": 1000000, "totalprofit": 500000, "documentcount": 5000},
		},
		"POST /goapi/api/approval/po-status/submit": {
			{"holdingcode": "SHOP001", "docno": "PO-2025-001", "submittedby": "EMP001"},
			{"success": true, "status": "pending", "message": "PO submitted for approval"},
		},
		"POST /goapi/api/approval/po-status/approve": {
			{"holdingcode": "SHOP001", "docno": "PO-2025-001", "approvedby": "MGR001", "comment": "อนุมัติ"},
			{"success": true, "status": "approved"},
		},
		"POST /goapi/get": {
			{"holdingcode": "SHOP001", "sql": "SELECT docno, docdate, totalamount FROM saleinvoice ORDER BY docdate DESC LIMIT 10"},
			{"success": true, "data": []map[string]interface{}{{"docno": "INV-001", "docdate": "2025-01-15", "totalamount": 15000}}},
		},
		"POST /goapi/genpdf": {
			{"holdingcode": "SHOP001", "docno": "INV-001", "template": "invoice"},
			{"success": true, "url": "/s3/file/SHOP001/pdf/INV-001.pdf", "filesize": 125000},
		},
		"POST /goapi/mcp/invoke": {
			{"tool": "getdailysales", "params": map[string]interface{}{"date": "2025-01-15"}},
			{"success": true, "data": map[string]interface{}{"date": "2025-01-15", "totalamount": 45000, "documentcount": 120}, "tool": "getdailysales"},
		},
		"POST /goapi/image/upload": {
			{"holdingcode": "SHOP001", "category": "product"},
			{"success": true, "filename": "abc123.jpg", "url": "/s3/file/SHOP001/images/abc123.jpg"},
		},
		"POST /goapi/api/lineoa/configs": {
			{"holdingcode": "SHOP001"},
			{"success": true, "configs": []map[string]interface{}{{"channelid": "123456", "channelname": "My Shop LINE"}}},
		},
		"POST /goapi/api/v1/chatbot/chat-gemini": {
			{"holdingcode": "SHOP001", "message": "ยอดขายวันนี้เท่าไร?"},
			{"success": true, "response": "ยอดขายวันนี้รวม 45,000 บาท จาก 120 บิล"},
		},
		"POST /goapi/api/v1/unified/query": {
			{"holdingcode": "SHOP001", "query": "สินค้าขายดี 10 อันดับเดือนนี้"},
			{"success": true, "data": []map[string]interface{}{{"name": "น้ำตาลทราย", "totalqty": 500, "totalamount": 17500}}},
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
