package shop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/pkg/apperr"
)

// The settings screen shows these responses to Thai users as-is: every mapped error must carry a
// stable code and a readable Thai message, and never leak the raw service/database text.
func TestShopUserAppErrorNeverLeaksRawErrors(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		code   string
		status int
	}{
		{"permission denied", errors.New("permission denied"), "FORBIDDEN", http.StatusForbidden},
		{"caller not a member", ErrShopUserNotFound, "FORBIDDEN", http.StatusForbidden},
		{"member not found", errors.New("user not found"), "NOT_FOUND", http.StatusNotFound},
		{"own role", errors.New("can not edit self permission"), "CONFLICT", http.StatusConflict},
		{"own account", errors.New("can't delete your permission"), "CONFLICT", http.StatusConflict},
		{"creator delete", errors.New("creator_cannot_delete"), "CONFLICT", http.StatusConflict},
		{"creator disable", errors.New("creator_access_cannot_be_disabled"), "CONFLICT", http.StatusConflict},
		{"database", errors.New("pq: connection refused; find failed"), "INTERNAL_ERROR", http.StatusInternalServerError},
		{"scope", errAccessScopeInvalid, "VALIDATION_FAILED", http.StatusBadRequest},
		{"username required", fmt.Errorf("%w: username required", errMemberRequestInvalid), "VALIDATION_FAILED", http.StatusBadRequest},
		{"holding required", fmt.Errorf("%w: holdingCode and request required", errMemberRequestInvalid), "VALIDATION_FAILED", http.StatusBadRequest},
		{"invalid useruid", fmt.Errorf("%w: invalid useruid", errMemberRequestInvalid), "VALIDATION_FAILED", http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := shopUserAppError(tc.err, "th")
			if got.Code != tc.code || got.StatusCode() != tc.status {
				t.Fatalf("got %s/%d, want %s/%d", got.Code, got.StatusCode(), tc.code, tc.status)
			}
			body, err := json.Marshal(got.ToResponse())
			if err != nil {
				t.Fatal(err)
			}
			if tc.err != errAccessScopeInvalid && strings.Contains(string(body), tc.err.Error()) {
				t.Fatalf("response leaks raw error: %s", body)
			}
			if !strings.ContainsAny(got.ThaiMsg, "กขคงจฉชซดตถทนบปผพฟมยรลวศสหอะาิีึืุูเแโใไ") || strings.HasPrefix(got.ThaiMsg, "ss_err_") {
				t.Fatalf("message_th is not Thai text: %q", got.ThaiMsg)
			}
		})
	}
}

// The repository rejects a save the caller can fix before touching the database; the handler
// must answer 400 (not 500) so the screen says "check the data" instead of "try again".
func TestSaveFullProfileRejectsFixableRequestsAsValidation(t *testing.T) {
	repo := &ShopUserPostgresRepository{}
	for name, req := range map[string]*models.UserRoleRequest{
		"no username or uid": {Role: models.ROLE_USER},
		"malformed useruid":  {Username: "staff", UserUID: "not-a-uuid"},
	} {
		t.Run(name, func(t *testing.T) {
			err := repo.SaveFullProfile(context.Background(), "rungrueng", req)
			if !errors.Is(err, errMemberRequestInvalid) {
				t.Fatalf("err = %v, want errMemberRequestInvalid", err)
			}
			if got := shopUserAppError(err, "th"); got.StatusCode() != http.StatusBadRequest || got.Code != "VALIDATION_FAILED" {
				t.Fatalf("mapped to %s/%d, want VALIDATION_FAILED/400", got.Code, got.StatusCode())
			}
		})
	}
	if err := repo.SaveFullProfile(context.Background(), " ", &models.UserRoleRequest{Username: "staff"}); !errors.Is(err, errMemberRequestInvalid) {
		t.Fatalf("empty holding: err = %v, want errMemberRequestInvalid", err)
	}
}

func TestShopUserAppErrorKeepsCauseForLogsOnly(t *testing.T) {
	cause := errors.New("connect central database: timeout")
	got := shopUserAppError(cause, "en")
	if !errors.Is(got, cause) {
		t.Fatalf("INTERNAL error must wrap the cause for the server log")
	}
	if strings.Contains(got.Message, "timeout") || got.Message == got.ThaiMsg {
		t.Fatalf("en message = %q, want the English row text", got.Message)
	}
	appErr := apperr.ErrDuplicate.WithField("username")
	if shopUserAppError(appErr, "th") != appErr {
		t.Fatalf("an AppError from the service must pass through unchanged")
	}
}

func TestInvalidBodyResponseHidesDecoderText(t *testing.T) {
	var target struct{ Items []string }
	decodeErr := json.Unmarshal([]byte(`{"Items":{"a":1}}`), &target)
	if decodeErr == nil {
		t.Fatal("expected a decode error")
	}
	got := localizedAppError(apperr.ErrBadRequest, "ss_err_invalid_data", "th").WithWrap(decodeErr)
	body, _ := json.Marshal(got.ToResponse())
	if got.StatusCode() != http.StatusBadRequest || strings.Contains(string(body), "unmarshal") || strings.Contains(string(body), "Go struct") {
		t.Fatalf("bind error response = %s", body)
	}
}

// Save rejections must name the form field to fix, with a 4xx status (never the generic
// "try again" 500). Their Thai text comes from the languages.tsv rows of the same keys.
func TestShopUserSaveRejectionsNameTheField(t *testing.T) {
	cases := []struct {
		err    error
		code   string
		status int
		field  string
		key    string
	}{
		{errMemberAlreadyExists, "DUPLICATE", http.StatusConflict, "username", "ss_err_user_already_member"},
		{errUsernameTaken, "DUPLICATE", http.StatusConflict, "username", "ss_err_user_code_taken"},
		{errLoginExists, "LOGIN_EXISTS", http.StatusConflict, "username", "ss_err_user_login_exists"},
		{errUserCodeLocked, "CONFLICT", http.StatusConflict, "username", "ss_err_user_code_locked"},
		{errUserEmailLocked, "CONFLICT", http.StatusConflict, "email", "ss_err_user_email_locked"},
		{errCreatorAccessExpiry, "CONFLICT", http.StatusConflict, "accessexpirydate", "ss_err_creator_access_expiry"},
		{models.ErrInvalidAccessExpiryDate, "VALIDATION_FAILED", http.StatusBadRequest, "accessexpirydate", "ss_err_access_expiry_date_invalid"},
		{errSaveTargetNotFound, "NOT_FOUND", http.StatusNotFound, "", "ss_err_not_found"},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			base, key, field := shopUserErrorRow(fmt.Errorf("save: %w", tc.err))
			got := shopUserAppError(tc.err, "th")
			if key != tc.key || field != tc.field || base.Code != tc.code {
				t.Fatalf("row = %s/%s/%s, want %s/%s/%s", base.Code, key, field, tc.code, tc.key, tc.field)
			}
			if got.Code != tc.code || got.StatusCode() != tc.status || got.Field != tc.field {
				t.Fatalf("response = %s/%d field %q", got.Code, got.StatusCode(), got.Field)
			}
			if body, _ := json.Marshal(got.ToResponse()); strings.Contains(string(body), tc.err.Error()) {
				t.Fatalf("response leaks raw error: %s", body)
			}
		})
	}
}

func TestSaveFullProfileRejectsBadExpiryBeforeTouchingTheDatabase(t *testing.T) {
	repo := &ShopUserPostgresRepository{} // nil db: reaching SQL would panic
	err := repo.SaveFullProfile(context.Background(), "rungrueng", &models.UserRoleRequest{Username: "staff", AccessExpiryDate: "31/12/2026"})
	if !errors.Is(err, models.ErrInvalidAccessExpiryDate) {
		t.Fatalf("err = %v, want ErrInvalidAccessExpiryDate", err)
	}
	if got := shopUserAppError(err, "th"); got.StatusCode() != http.StatusBadRequest || got.Field != "accessexpirydate" {
		t.Fatalf("response = %d field %q, want 400 accessexpirydate", got.StatusCode(), got.Field)
	}
}
