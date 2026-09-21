package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"smlcloudplatform/internal/mcpgateway"
	"smlcloudplatform/internal/mcptoken"
	"smlcloudplatform/pkg/microservice"
)

// MCP reuses the authenticated HTTP handlers, including their permission checks,
// branch constraints, decimal decoder, optimistic concurrency and idempotency.
func (h *Http) registerMCP() {
	handler := mcpgateway.New(mcptoken.Authenticate, h.executeMCP)
	h.ms.Echo().Any("/mcp/gl", echo.WrapHandler(handler))
	h.registerTokenAPI()
}

type mcpGLContext struct {
	microservice.IContext
	request   *http.Request
	params    map[string]string
	status    int
	payload   any
	tokenID   string
	tokenKind string
}

func (c *mcpGLContext) Request() *http.Request        { return c.request }
func (c *mcpGLContext) Param(name string) string      { return c.params[name] }
func (c *mcpGLContext) QueryParam(name string) string { return c.request.URL.Query().Get(name) }
func (c *mcpGLContext) Response(code int, data any)   { c.status = code; c.payload = data }
func (c *mcpGLContext) ResponseError(code int, message string) {
	c.Response(code, map[string]any{"success": false, "message": message})
}

func (h *Http) executeMCP(ctx context.Context, p mcptoken.Principal, tool string, raw json.RawMessage) (any, error) {
	var args struct {
		CompanyCodes []string          `json:"companyCodes"`
		CompanyCode  string            `json:"companyCode"`
		Resource     string            `json:"resource"`
		ID           string            `json:"id"`
		Report       string            `json:"report"`
		Query        map[string]string `json:"query"`
		Command      json.RawMessage   `json:"command"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&args); err != nil {
		return nil, fmt.Errorf("รูปแบบ arguments ไม่ถูกต้อง")
	}
	query := url.Values{}
	allowedQuery := map[string]bool{"q": true, "page": true, "limit": true, "bookcode": true, "kind": true, "status": true, "from": true, "to": true, "fiscalyear": true, "accountcode": true, "branchcode": true, "departmentcode": true, "projectcode": true, "snapshot": true, "asof": true, "companywide": true}
	for key, value := range args.Query {
		if !allowedQuery[key] {
			return nil, fmt.Errorf("query parameter ไม่ถูกต้อง: %s", key)
		}
		query.Set(key, value)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "http://mcp.internal/gl?"+query.Encode(), bytes.NewReader(args.Command))
	req.Header.Set("Content-Type", "application/json")
	// The response is captured by mcpGLContext; the underlying writer is never
	// exposed to callers or used to make a second network request.
	base := echo.New().NewContext(req, &discardResponseWriter{header: http.Header{}})
	base.Set("UserInfo", p.User)
	request := &mcpGLContext{IContext: microservice.NewHTTPContext(h.ms, base), request: req, params: map[string]string{"resource": args.Resource, "id": args.ID, "report": args.Report}, tokenID: p.ID, tokenKind: "mcp"}
	var run func(microservice.IContext) error
	switch tool {
	case "gl_list", "gl_get":
		_, master := resourceScreens[args.Resource]
		if !master && args.Resource != "journals" && !(tool == "gl_get" && args.Resource == "journal-reviews") {
			return nil, fmt.Errorf("ไม่พบ resource")
		}
		if tool == "gl_get" {
			if args.ID == "" {
				return nil, fmt.Errorf("กรุณาระบุ id")
			}
			run = h.get
		} else {
			run = h.list
		}
	case "gl_report":
		if _, ok := reportScreens[args.Report]; !ok {
			return nil, fmt.Errorf("ไม่พบรายงาน")
		}
		run = h.report
	case "gl_command":
		if p.Mode != "readwrite" {
			return nil, fmt.Errorf("token นี้มีสิทธิ์ readonly")
		}
		if len(args.Command) == 0 {
			return nil, fmt.Errorf("กรุณาระบุ command")
		}
		run = h.command
	default:
		return nil, fmt.Errorf("ไม่พบ tool")
	}
	companies, err := mcptoken.CompanySelection(p.Mode, args.CompanyCode, args.CompanyCodes)
	if err != nil {
		return nil, err
	}
	principals := make([]mcptoken.Principal, 0, len(companies))
	for _, code := range companies {
		resolved, err := mcptoken.ResolveCompany(ctx, p, code)
		if err != nil {
			return nil, fmt.Errorf("ไม่มีสิทธิ์เข้าถึงบริษัทที่เลือก")
		}
		principals = append(principals, resolved)
	}
	results := make([]any, 0, len(principals))
	allSucceeded := true
	for _, principal := range principals {
		base.Set("UserInfo", principal.User)
		request.status, request.payload = 0, nil
		if err := run(request); err != nil {
			return nil, fmt.Errorf("ไม่สามารถดำเนินการได้")
		}
		if len(principals) == 1 {
			if request.status >= 400 {
				encoded, _ := json.Marshal(request.payload)
				return nil, fmt.Errorf("%s", encoded)
			}
			return request.payload, nil
		}
		results = append(results, map[string]any{"companyCode": principal.User.BusinessCode, "response": request.payload})
		allSucceeded = allSucceeded && request.status > 0 && request.status < 400
	}
	return map[string]any{"success": allSucceeded, "data": map[string]any{"companies": results}}, nil
}

type discardResponseWriter struct{ header http.Header }

func (w *discardResponseWriter) Header() http.Header         { return w.header }
func (w *discardResponseWriter) WriteHeader(int)             {}
func (w *discardResponseWriter) Write(b []byte) (int, error) { return io.Discard.Write(b) }
