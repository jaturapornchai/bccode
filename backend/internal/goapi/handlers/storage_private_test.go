package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
)

func TestStoragePrivateURLDefaultUsesBackendProxy(t *testing.T) {
	t.Setenv(storageAllowPresignedURLEnv, "")

	got, err := getPresignedURL(nil, "SHOP001/images/a.png", 60)
	if err != nil {
		t.Fatalf("getPresignedURL returned error: %v", err)
	}

	want := "/s3/file/SHOP001/images/a.png"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestStoragePrivateURLDirectPresignedRequiresOptIn(t *testing.T) {
	t.Setenv(storageAllowPresignedURLEnv, "true")
	if !storageAllowsDirectPresignedURL() {
		t.Fatal("expected direct presigned URL to be enabled when env=true")
	}

	t.Setenv(storageAllowPresignedURLEnv, "false")
	if storageAllowsDirectPresignedURL() {
		t.Fatal("expected direct presigned URL to be disabled when env=false")
	}
}

func TestStorageProxyURLAlwaysIncludesShopPrefix(t *testing.T) {
	want := "/s3/file/SHOP001/images/a.png"
	if got := storageProxyURL("SHOP001/images/a.png"); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
	if got := storageProxyURL("/SHOP001/images/a.png/"); got != want {
		t.Fatalf("expected proxy URL to trim slashes, got %q", got)
	}
	if got := storageProxyURL("SHOP001\\images\\a.png"); got != want {
		t.Fatalf("expected proxy URL to normalize backslashes, got %q", got)
	}
}

func TestStorageObjectBelongsToShop(t *testing.T) {
	tests := []struct {
		name        string
		objectKey   string
		holdingCode string
		want        bool
	}{
		{name: "same shop", objectKey: "SHOP001/images/a.png", holdingCode: "SHOP001", want: true},
		{name: "different shop", objectKey: "SHOP002/images/a.png", holdingCode: "SHOP001", want: false},
		{name: "path traversal", objectKey: "SHOP001/../SHOP002/a.png", holdingCode: "SHOP001", want: false},
		{name: "missing shop", objectKey: "uploads/20260521/a.png", holdingCode: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := storageObjectBelongsToShop(tt.objectKey, tt.holdingCode)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestS3FileProxyBlocksCrossShopKeyBeforeStorage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/s3/file/SHOP002/images/a.png", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("*")
	c.SetParamValues("SHOP002/images/a.png")
	c.Set("UserInfo", msmodels.UserInfo{Username: "user@example.com", HoldingCode: "SHOP001"})

	if err := S3FileProxyHandler(c); err != nil {
		t.Fatalf("S3FileProxyHandler returned error: %v", err)
	}

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestStorageAuthorizedHoldingCode(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("UserInfo", msmodels.UserInfo{Username: "user@example.com", HoldingCode: "SHOP001"})

	holdingCode, status := storageAuthorizedHoldingCode(c, "")
	if status != http.StatusOK || holdingCode != "SHOP001" {
		t.Fatalf("expected token shop SHOP001 with status 200, got shop=%q status=%d", holdingCode, status)
	}

	holdingCode, status = storageAuthorizedHoldingCode(c, "SHOP001")
	if status != http.StatusOK || holdingCode != "SHOP001" {
		t.Fatalf("expected matching requested shop SHOP001 with status 200, got shop=%q status=%d", holdingCode, status)
	}

	_, status = storageAuthorizedHoldingCode(c, "SHOP002")
	if status != http.StatusForbidden {
		t.Fatalf("expected status %d for cross-shop request, got %d", http.StatusForbidden, status)
	}
}

func TestStorageAuthorizedHoldingCodeRejectsMissingToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if _, status := storageAuthorizedHoldingCode(c, ""); status != http.StatusUnauthorized {
		t.Fatalf("expected status %d when token is missing, got %d", http.StatusUnauthorized, status)
	}
}

func TestS3FileProxyBlocksPathTraversal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/s3/file/SHOP001/../SHOP002/a.png", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("*")
	c.SetParamValues("SHOP001/../SHOP002/a.png")
	c.Set("UserInfo", msmodels.UserInfo{Username: "user@example.com", HoldingCode: "SHOP001"})

	if err := S3FileProxyHandler(c); err != nil {
		t.Fatalf("S3FileProxyHandler returned error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d for path traversal, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestS3FileProxyRequiresAuthShop(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/s3/file/SHOP001/images/a.png", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("*")
	c.SetParamValues("SHOP001/images/a.png")

	if err := S3FileProxyHandler(c); err != nil {
		t.Fatalf("S3FileProxyHandler returned error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d when shop not selected, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func newMultipartUploadContext(t *testing.T, e *echo.Echo, formHoldingCode string) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if formHoldingCode != "" {
		if err := writer.WriteField("holding_code", formHoldingCode); err != nil {
			t.Fatalf("write holding_code field: %v", err)
		}
	}
	part, err := writer.CreateFormFile("file", "noop.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte("noop")); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/image/upload", body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestImageUploadHandlerRejectsCrossShopForm(t *testing.T) {
	e := echo.New()
	c, rec := newMultipartUploadContext(t, e, "SHOP002")
	c.Set("UserInfo", msmodels.UserInfo{Username: "user@example.com", HoldingCode: "SHOP001"})

	if err := ImageUploadHandler(c); err != nil {
		t.Fatalf("ImageUploadHandler returned error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d when form holding_code mismatches token, got %d", http.StatusForbidden, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Forbidden") {
		t.Fatalf("expected response body to mention Forbidden, got %s", rec.Body.String())
	}
}

func TestImageUploadHandlerRejectsMissingShopToken(t *testing.T) {
	e := echo.New()
	c, rec := newMultipartUploadContext(t, e, "")

	if err := ImageUploadHandler(c); err != nil {
		t.Fatalf("ImageUploadHandler returned error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d when shop not selected, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestFileUploadHandlerRejectsCrossShopForm(t *testing.T) {
	e := echo.New()
	c, rec := newMultipartUploadContext(t, e, "SHOP002")
	c.Set("UserInfo", msmodels.UserInfo{Username: "user@example.com", HoldingCode: "SHOP001"})

	if err := FileUploadHandler(c); err != nil {
		t.Fatalf("FileUploadHandler returned error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d when form holding_code mismatches token, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestAttachmentUploadHandlerRejectsCrossShopForm(t *testing.T) {
	e := echo.New()
	c, rec := newMultipartUploadContext(t, e, "SHOP002")
	c.Set("UserInfo", msmodels.UserInfo{Username: "user@example.com", HoldingCode: "SHOP001"})

	if err := AttachmentUploadHandler(c); err != nil {
		t.Fatalf("AttachmentUploadHandler returned error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d when form holding_code mismatches token, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestAttachmentDownloadHandlerRejectsCrossShopQuery(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/attachment/download/507f1f77bcf86cd799439011?holding_code=SHOP002", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("507f1f77bcf86cd799439011")
	c.Set("UserInfo", msmodels.UserInfo{Username: "user@example.com", HoldingCode: "SHOP001"})

	if err := AttachmentDownloadHandler(c); err != nil {
		t.Fatalf("AttachmentDownloadHandler returned error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d when query holding_code mismatches token, got %d", http.StatusForbidden, rec.Code)
	}
}
