package goapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"smlcloudplatform/pkg/microservice"
	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

func TestGoAPIRequestHoldingCodeReadsJSONAndPreservesBody(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/get", strings.NewReader(`{"holding_code":"SHOP001"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	holdingCode, err := goAPIRequestHoldingCode(c)
	if err != nil {
		t.Fatalf("goAPIRequestHoldingCode returned error: %v", err)
	}
	if holdingCode != "SHOP001" {
		t.Fatalf("holdingCode = %q, want SHOP001", holdingCode)
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		t.Fatalf("read preserved body: %v", err)
	}
	if string(body) != `{"holding_code":"SHOP001"}` {
		t.Fatalf("body = %q, want original body", string(body))
	}
}

func TestHoldingCodeFromPayloadReadsNestedJSONBody(t *testing.T) {
	holdingCode, err := holdingCodeFromPayload(map[string]interface{}{
		"body": `{"holding_code":"SHOP002"}`,
	})
	if err != nil {
		t.Fatalf("holdingCodeFromPayload returned error: %v", err)
	}
	if holdingCode != "SHOP002" {
		t.Fatalf("holdingCode = %q, want SHOP002", holdingCode)
	}
}

func TestAuthenticateGoAPIJWTTokenUsesConfiguredSecret(t *testing.T) {
	t.Setenv("MODE", "production")
	t.Setenv("JWT_SECRET_KEY", "test-secret")

	claims := microservice.CustomClaims{
		RegisteredClaims: &jwt.RegisteredClaims{},
		UserInfo: msmodels.UserInfo{
			Username:    "user@example.com",
			HoldingCode: "SHOP003",
			Role:        2,
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	userInfo, err := authenticateGoAPIJWTToken(token)
	if err != nil {
		t.Fatalf("authenticateGoAPIJWTToken returned error: %v", err)
	}
	if userInfo.HoldingCode != "SHOP003" || userInfo.Username != "user@example.com" {
		t.Fatalf("userInfo = %+v", userInfo)
	}
}
