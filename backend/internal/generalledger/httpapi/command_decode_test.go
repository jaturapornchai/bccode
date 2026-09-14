package httpapi

import (
	"encoding/json"
	"strings"
	"testing"
)

// The exact JSON the chart-of-accounts screen sends (frontend/src/app/gl/gl-masters.tsx:123
// + frontend/src/lib/general-ledger.ts emptyAccount(), plus the requestid added by
// frontend/src/lib/general-ledger-api.ts).
const frontendAccountCreate = `{
  "resource": "accounts",
  "action": "create",
  "requestid": "0f8c1c0a4a3b4f2e9d1a7c6b5e4d3c2b",
  "reason": "สร้างข้อมูลใหม่",
  "account": {
    "accountcode": "1100",
    "names": [{"code": "th", "name": "เงินสด"}],
    "accounttype": "asset",
    "parentaccountcode": "",
    "normalbalance": "debit",
    "allowposting": true,
    "isactive": true,
    "accountgroup": "",
    "iscash": true,
    "level": 1
  }
}`

func TestFrontendAccountCreateBodyDecodesWithCurrentSource(t *testing.T) {
	cmd, err := decodeCommand(strings.NewReader(frontendAccountCreate))
	if err != nil {
		t.Fatalf("frontend account create body rejected by current source: %v", err)
	}
	if cmd.Resource != "accounts" || cmd.Action != "create" || cmd.RequestID == "" {
		t.Fatalf("command decoded wrong: %+v", cmd)
	}
	if cmd.Account == nil {
		t.Fatal("account payload missing after decode")
	}
	if cmd.Account.AccountCode != "1100" || cmd.Account.Level != 1 || !cmd.Account.IsCash || len(cmd.Account.Names) != 1 {
		t.Fatalf("account decoded wrong: %+v", *cmd.Account)
	}
	if err := cmd.Account.Validate(); err != nil {
		t.Fatalf("account payload fails model validation: %v", err)
	}
}

// Any key the backend struct does not know makes decodeCommand fail, which the
// handler reports as the (misleading) amount message. This test pins that trap.
func TestUnknownAccountFieldIsRejectedByDecodeGuard(t *testing.T) {
	body := strings.Replace(frontendAccountCreate, `"level": 1`, `"level": 1, "amount": "0.00"`, 1)
	_, err := decodeCommand(strings.NewReader(body))
	if err == nil {
		t.Fatal("expected an unknown field to be rejected")
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unexpected decode error: %v", err)
	}
}

// A journal-shaped amount inside the account object is the only way to reach the
// amount wording; the accounts struct has no amount field at all.
func TestAccountPayloadHasNoAmountField(t *testing.T) {
	cmd, err := decodeCommand(strings.NewReader(frontendAccountCreate))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, err := decodeCommand(strings.NewReader(strings.Replace(
		frontendAccountCreate, `"iscash": true`, `"iscash": true, "lines": []`, 1))); err == nil {
		t.Fatal("account body accepted journal-only field `lines`")
	}
	_ = cmd
}

// UAT fix: a decode failure is now a Thai, machine-readable 400 instead of the
// old fixed "amount" wording, and never leaks the decoder's English text.
func TestDecodeFailureIsThaiInvalidPayload(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		wantPart   string
		wantAmount bool
	}{
		{
			name:     "unknown field",
			body:     strings.Replace(frontendAccountCreate, `"level": 1`, `"level": 1, "amount": "0.00"`, 1),
			wantPart: "ไม่รองรับฟิลด์ amount",
		},
		{
			name:     "wrong type",
			body:     strings.Replace(frontendAccountCreate, `"level": 1`, `"level": "1"`, 1),
			wantPart: "ชนิดข้อมูลของฟิลด์ account.level ไม่ถูกต้อง",
		},
		{
			name:       "numeric amount in a journal line",
			body:       `{"resource":"journals","action":"create","requestid":"0f8c1c0a4a3b4f2e9d1a7c6b5e4d3c2b","journal":{"lines":[{"accountcode":"1100","debit":100}]}}`,
			wantAmount: true,
		},
	}
	for _, tc := range cases {
		_, err := decodeCommand(strings.NewReader(tc.body))
		if err == nil {
			t.Fatalf("%s: expected a decode failure", tc.name)
		}
		payload := decodeFailure(err)
		if payload.Code != "invalid_payload" || payload.Success {
			t.Fatalf("%s: payload = %+v, want code invalid_payload", tc.name, payload)
		}
		if tc.wantAmount {
			if payload.Message != amountFieldMessage {
				t.Fatalf("%s: message = %q", tc.name, payload.Message)
			}
		} else if !strings.Contains(payload.Message, tc.wantPart) {
			t.Fatalf("%s: message = %q, want it to contain %q", tc.name, payload.Message, tc.wantPart)
		}
		for _, leak := range []string{"json:", "cannot unmarshal", "Go struct", "httpapi.gl."} {
			if strings.Contains(payload.Message, leak) {
				t.Fatalf("%s: message leaks decoder text: %q", tc.name, payload.Message)
			}
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("%s: marshal: %v", tc.name, err)
		}
		t.Logf("%s -> 400 %s", tc.name, raw)
	}
}
