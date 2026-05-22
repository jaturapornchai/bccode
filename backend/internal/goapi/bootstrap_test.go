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

func TestGoAPIRequestShopIDReadsJSONAndPreservesBody(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/get", strings.NewReader(`{"shop_id":"SHOP001"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	shopID, err := goAPIRequestShopID(c)
	if err != nil {
		t.Fatalf("goAPIRequestShopID returned error: %v", err)
	}
	if shopID != "SHOP001" {
		t.Fatalf("shopID = %q, want SHOP001", shopID)
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		t.Fatalf("read preserved body: %v", err)
	}
	if string(body) != `{"shop_id":"SHOP001"}` {
		t.Fatalf("body = %q, want original body", string(body))
	}
}

func TestShopIDFromPayloadReadsNestedJSONBody(t *testing.T) {
	shopID, err := shopIDFromPayload(map[string]interface{}{
		"body": `{"shopid":"SHOP002"}`,
	})
	if err != nil {
		t.Fatalf("shopIDFromPayload returned error: %v", err)
	}
	if shopID != "SHOP002" {
		t.Fatalf("shopID = %q, want SHOP002", shopID)
	}
}

func TestAuthenticateGoAPIJWTTokenUsesConfiguredSecret(t *testing.T) {
	t.Setenv("MODE", "production")
	t.Setenv("JWT_SECRET_KEY", "test-secret")

	claims := microservice.CustomClaims{
		RegisteredClaims: &jwt.RegisteredClaims{},
		UserInfo: msmodels.UserInfo{
			Username: "user@example.com",
			ShopID:   "SHOP003",
			Role:     2,
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
	if userInfo.ShopID != "SHOP003" || userInfo.Username != "user@example.com" {
		t.Fatalf("userInfo = %+v", userInfo)
	}
}
