// Package mcpgateway serves a stateless MCP Streamable HTTP endpoint.
package mcpgateway

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"smlcloudplatform/internal/mcptoken"
)

type Authenticate func(context.Context, string) (mcptoken.Principal, error)
type Execute func(context.Context, mcptoken.Principal, string, json.RawMessage) (any, error)
type Handler struct {
	authenticate Authenticate
	execute      Execute
}

func New(auth Authenticate, execute Execute) *Handler { return &Handler{auth, execute} }

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func rpcError(w http.ResponseWriter, id json.RawMessage, code int, message string) {
	if len(id) == 0 {
		id = json.RawMessage("null")
	}
	writeJSON(w, 200, map[string]any{"jsonrpc": "2.0", "id": id, "error": map[string]any{"code": code, "message": message}})
}
func validVersion(v string) bool { return v == "2025-03-26" || v == "2025-06-18" || v == "2025-11-25" }
func validOrigin(origin string) bool {
	if origin == "" {
		return true
	}
	allowed := os.Getenv("BCAI_MCP_ALLOWED_ORIGINS")
	if allowed == "" {
		allowed = "https://account.bcaicloud.com"
	}
	for _, item := range strings.Split(allowed, ",") {
		if origin == strings.TrimSpace(item) {
			return true
		}
	}
	return false
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if !validOrigin(r.Header.Get("Origin")) {
		http.Error(w, "Origin denied", 403)
		return
	}
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		unauthorized(w)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 35*time.Second)
	defer cancel()
	p, err := h.authenticate(ctx, parts[1])
	if err != nil {
		unauthorized(w)
		return
	}
	if p.Kind != "mcp" || (p.Mode != "readonly" && p.Mode != "readwrite") {
		unauthorized(w)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method not allowed", 405)
		return
	}
	if v := r.Header.Get("MCP-Protocol-Version"); v != "" && !validVersion(v) {
		http.Error(w, "Unsupported MCP protocol version", 400)
		return
	}
	media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if media != "application/json" {
		http.Error(w, "Expected application/json", 415)
		return
	}
	accept := r.Header.Get("Accept")
	if !strings.Contains(accept, "application/json") || !strings.Contains(accept, "text/event-stream") {
		http.Error(w, "Accept must include application/json and text/event-stream", 406)
		return
	}
	var request rpcRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20))
	if decoder.Decode(&request) != nil {
		rpcError(w, nil, -32700, "Invalid JSON request")
		return
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF || request.JSONRPC != "2.0" || request.Method == "" {
		rpcError(w, request.ID, -32600, "Invalid request")
		return
	}
	if len(request.ID) == 0 {
		if request.Method != "notifications/initialized" && request.Method != "notifications/cancelled" {
			http.Error(w, "Unsupported notification", 400)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		return
	}
	var id any
	idDecoder := json.NewDecoder(strings.NewReader(string(request.ID)))
	idDecoder.UseNumber()
	_ = idDecoder.Decode(&id)
	switch id.(type) {
	case string, json.Number:
	default:
		rpcError(w, nil, -32600, "Invalid request id")
		return
	}
	var result any
	switch request.Method {
	case "initialize":
		var params struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		if json.Unmarshal(request.Params, &params) != nil || params.ProtocolVersion == "" {
			rpcError(w, request.ID, -32602, "protocolVersion is required")
			return
		}
		version := params.ProtocolVersion
		if !validVersion(version) {
			version = "2025-11-25"
		}
		result = map[string]any{"protocolVersion": version, "capabilities": map[string]any{"tools": map[string]any{"listChanged": false}}, "serverInfo": map[string]string{"name": "BC GL", "version": "1.0.0"}, "instructions": "Use decimal strings for accounting amounts. Holding is fixed by the token. Supply companyCode for one company, or companyCodes for readonly multi-company reads. Readwrite requests permit only one company, including reads. Multi-company results remain separate; amounts are not consolidated. Legacy branch restrictions remain enforced. Writes use existing GL validation and require explicit tool calls; imports do not post automatically."}
	case "ping":
		result = map[string]any{}
	case "tools/list":
		result = map[string]any{"tools": tools(p.Mode)}
	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if json.Unmarshal(request.Params, &params) != nil {
			rpcError(w, request.ID, -32602, "Invalid tool parameters")
			return
		}
		if params.Name == "gl_command" && p.Mode != "readwrite" {
			rpcError(w, request.ID, -32602, "Token is readonly; write denied")
			return
		}
		if params.Name != "gl_list" && params.Name != "gl_get" && params.Name != "gl_report" && params.Name != "gl_command" {
			rpcError(w, request.ID, -32602, "Unknown tool")
			return
		}
		data, err := h.execute(ctx, p, params.Name, params.Arguments)
		if err != nil {
			result = map[string]any{"content": []any{map[string]string{"type": "text", "text": err.Error()}}, "isError": true}
		} else {
			encoded, e := json.Marshal(data)
			if e != nil {
				rpcError(w, request.ID, -32603, "Unable to encode result")
				return
			}
			isError := false
			if envelope, ok := data.(map[string]any); ok {
				if success, ok := envelope["success"].(bool); ok {
					isError = !success
				}
			}
			result = map[string]any{"content": []any{map[string]string{"type": "text", "text": string(encoded)}}, "isError": isError}
		}
	default:
		rpcError(w, request.ID, -32601, "Method not found")
		return
	}
	writeJSON(w, 200, map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
}
func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="BC GL MCP"`)
	writeJSON(w, 401, map[string]string{"error": "invalid_token"})
}

func tools(mode string) []any {
	stringField := map[string]any{"type": "string"}
	query := map[string]any{"type": "object", "additionalProperties": stringField, "description": "Optional GL query parameters as strings: q, page, limit, bookcode, kind, status, from, to, fiscalyear, accountcode, branchcode, departmentcode, projectcode, snapshot, asof (YYYY-MM-DD for journal-support), companywide (true requests authorized whole-company report scope)."}
	makeTool := func(name, description string, properties map[string]any, required []string, read bool) any {
		properties["companyCode"] = map[string]any{"type": "string", "description": "Company code authorized for this token within its Holding"}
		properties["companyCodes"] = map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "minItems": 1, "maxItems": 50, "description": "Readonly only: select multiple authorized companies. Readwrite permits only one company per request, including reads."}
		return map[string]any{"name": name, "description": description, "inputSchema": map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}, "annotations": map[string]any{"readOnlyHint": read, "destructiveHint": !read, "openWorldHint": false}}
	}
	result := []any{
		makeTool("gl_list", "Read GL accounts, journals, journal-books and other GL master records with pagination. For accounting evidence use resource journal-support and query.kind: partners, bank-accounts, documents, allocations, settlements, statements, bank-lines or matches.", map[string]any{"resource": stringField, "query": query}, []string{"resource"}, true),
		makeTool("gl_get", "Read one GL record or journal-reviews by ID.", map[string]any{"resource": stringField, "id": stringField}, []string{"resource", "id"}, true),
		makeTool("gl_report", "Read GL reports such as ledger, trialbalance, pnl, balancesheet, ar-outstanding, ap-outstanding and bank-unmatched. Use query.to for the accounting as-of date.", map[string]any{"report": stringField, "query": query}, []string{"report"}, true),
	}
	if mode == "readwrite" {
		result = append(result, makeTool("gl_command", "Execute one GL command using the same contract as POST /gl/v2/command. Requires resource, action and unique UUID requestid; updates require version. Money must be decimal strings. Posting/reversal/closing are explicit actions, never automatic. Journal imports use source_type, source_system and source_record_id to prevent duplicate source records. AR/AP/bank evidence is journal.details. On posted journals use action reconcile with id, version, reason and journal.details containing new statement_lines, settlements, matches or withdrawals; replacing an allocation requires withdrawal plus replacement allocations in the same request. Accounting lines remain immutable. journal.bookcode must be the code of an active journal-books record that has a booktype; journal-books records are {code (max 15 characters), name (Thai, required), nameen, booktype 1 general, 2 payment, 3 receipt, 4 sales, 5 purchase, 6 opening, isactive}; a book used by any journal cannot be deleted (deactivate it instead) and its code/booktype cannot change. Codes (docno max 30, account code max 20) may contain Thai letters, vowel signs and tone marks, digits and _ . - / # ( ) : but no spaces. A company-wide token must send journal.branchcode (an active branch of the company).", map[string]any{"command": map[string]any{"type": "object", "required": []string{"resource", "action", "requestid"}, "properties": map[string]any{"resource": stringField, "action": stringField, "requestid": stringField}, "additionalProperties": true}}, []string{"command"}, false))
	}
	return result
}
