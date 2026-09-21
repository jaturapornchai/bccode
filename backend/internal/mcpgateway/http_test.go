package mcpgateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"smlcloudplatform/internal/mcptoken"
)

func call(t *testing.T, h http.Handler, body string, modify func(*http.Request)) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest("POST", "http://local/mcp/gl", strings.NewReader(body))
	r.Header.Set("Authorization", "Bearer test")
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json, text/event-stream")
	if modify != nil {
		modify(r)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestReadOnlyCannotReachWriteExecutor(t *testing.T) {
	runs := 0
	h := New(func(context.Context, string) (mcptoken.Principal, error) {
		return mcptoken.Principal{Kind: "mcp", Mode: "readonly"}, nil
	}, func(context.Context, mcptoken.Principal, string, json.RawMessage) (any, error) {
		runs++
		return nil, nil
	})
	w := call(t, h, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`, nil)
	if w.Code != 200 || strings.Contains(w.Body.String(), "gl_command") || !strings.Contains(w.Body.String(), "gl_list") {
		t.Fatal(w.Body.String())
	}
	for _, body := range []string{
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"gl_command","arguments":{"command":{"resource":"journals","action":"post"}}}}`,
		`{"jsonrpc":"2.0","method":"tools/call","params":{"name":"gl_command"}}`,
		`{"jsonrpc":"2.0","id":3,"method":"gl_command"}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"unknown"}}`,
	} {
		call(t, h, body, nil)
	}
	if runs != 0 {
		t.Fatalf("readonly invoked executor %d times", runs)
	}
}
func TestWriteAndDecimalStringsPreserved(t *testing.T) {
	var got json.RawMessage
	h := New(func(context.Context, string) (mcptoken.Principal, error) {
		return mcptoken.Principal{Kind: "mcp", ID: "token1", Mode: "readwrite"}, nil
	}, func(_ context.Context, p mcptoken.Principal, tool string, args json.RawMessage) (any, error) {
		if p.ID != "token1" || tool != "gl_command" {
			t.Fatal("principal or tool changed")
		}
		got = args
		return map[string]string{"amount": "9007199254740993.01"}, nil
	})
	w := call(t, h, `{"jsonrpc":"2.0","id":9007199254740993,"method":"tools/call","params":{"name":"gl_command","arguments":{"command":{"amount":"9007199254740993.01"}}}}`, nil)
	if w.Code != 200 || !strings.Contains(string(got), `"9007199254740993.01"`) || !strings.Contains(w.Body.String(), `"id":9007199254740993`) {
		t.Fatal(w.Body.String())
	}
}
func TestAuthenticationAndTransport(t *testing.T) {
	for _, test := range []struct {
		name    string
		edit    func(*http.Request)
		authErr bool
		mode    string
		want    int
	}{
		{"missing token", func(r *http.Request) { r.Header.Del("Authorization") }, false, "readonly", 401},
		{"query token rejected", func(r *http.Request) { r.Header.Del("Authorization"); r.URL.RawQuery = "access_token=test" }, false, "readonly", 401},
		{"revoked or DB unavailable", nil, true, "readonly", 401},
		{"invalid mode", nil, false, "admin", 401},
		{"invalid origin", func(r *http.Request) { r.Header.Set("Origin", "https://evil.example") }, false, "readonly", 403},
		{"no SSE", func(r *http.Request) { r.Method = "GET" }, false, "readonly", 405},
		{"unsupported version", func(r *http.Request) { r.Header.Set("MCP-Protocol-Version", "2099-01-01") }, false, "readonly", 400},
		{"wrong content", func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") }, false, "readonly", 415},
		{"wrong accept", func(r *http.Request) { r.Header.Set("Accept", "application/json") }, false, "readonly", 406},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := New(func(context.Context, string) (mcptoken.Principal, error) {
				if test.authErr {
					return mcptoken.Principal{}, errors.New("secret DB connection details")
				}
				return mcptoken.Principal{Kind: "mcp", Mode: test.mode}, nil
			}, func(context.Context, mcptoken.Principal, string, json.RawMessage) (any, error) {
				t.Fatal("executor reached")
				return nil, nil
			})
			w := call(t, h, `{"jsonrpc":"2.0","id":1,"method":"ping"}`, test.edit)
			if w.Code != test.want || strings.Contains(w.Body.String(), "secret DB") || w.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
		})
	}
}
func TestLifecycleAndToolError(t *testing.T) {
	h := New(func(context.Context, string) (mcptoken.Principal, error) {
		return mcptoken.Principal{Kind: "mcp", Mode: "readonly"}, nil
	}, func(context.Context, mcptoken.Principal, string, json.RawMessage) (any, error) {
		return nil, errors.New("denied by ledger")
	})
	w := call(t, h, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25"}}`, nil)
	if !strings.Contains(w.Body.String(), `"protocolVersion":"2025-11-25"`) {
		t.Fatal(w.Body.String())
	}
	w = call(t, h, `{"jsonrpc":"2.0","method":"notifications/initialized"}`, nil)
	if w.Code != 202 || w.Body.Len() != 0 {
		t.Fatal(w.Code, w.Body.String())
	}
	w = call(t, h, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"gl_list","arguments":{"resource":"accounts"}}}`, nil)
	if !strings.Contains(w.Body.String(), `"isError":true`) {
		t.Fatal(w.Body.String())
	}
	w = call(t, h, `{"jsonrpc":"2.0","id":[],"method":"ping"}`, nil)
	if !strings.Contains(w.Body.String(), `"code":-32600`) {
		t.Fatal(w.Body.String())
	}
	w = call(t, h, `{"jsonrpc":"2.0","id":3,"method":"ping"} {}`, nil)
	if !strings.Contains(w.Body.String(), `"code":-32600`) {
		t.Fatal(w.Body.String())
	}
}
