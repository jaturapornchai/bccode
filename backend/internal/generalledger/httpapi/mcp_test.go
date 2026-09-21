package httpapi

import (
	"context"
	"encoding/json"
	"testing"

	"smlcloudplatform/internal/mcptoken"
)

func TestMCPRejectsScopeOverridesAndReadonlyWritesBeforeLedger(t *testing.T) {
	h := &Http{}
	for _, tc := range []struct{ tool, mode, args string }{
		{"gl_command", "readonly", `{"command":{"resource":"journals","action":"reconcile","journal":{"details":{"matches":[]}}}}`},
		{"gl_list", "readwrite", `{"resource":"journal-support","companyCodes":["C01","C02"],"query":{"kind":"documents","asof":"2026-06-30"}}`},
		{"gl_command", "readonly", `{"command":{"resource":"journals","action":"post"}}`},
		{"gl_list", "readwrite", `{"resource":"accounts","companyCodes":["C01","C02"]}`},
		{"gl_command", "readwrite", `{"companyCodes":["C01","C02"],"command":{"resource":"journals","action":"post"}}`},
		{"gl_list", "readonly", `{"resource":"accounts","holdingcode":"OTHER"}`},
		{"gl_list", "readonly", `{"resource":"accounts","query":{"companycode":"OTHER"}}`},
		{"gl_get", "readonly", `{"resource":"mcp-tokens","id":"anything"}`},
		{"gl_report", "readonly", `{"report":"sql","query":{"q":"DROP TABLE"}}`},
	} {
		_, err := h.executeMCP(context.Background(), mcptoken.Principal{Mode: tc.mode}, tc.tool, json.RawMessage(tc.args))
		if err == nil {
			t.Fatalf("accepted unsafe call %s", tc.tool)
		}
	}
}
