package authentication

import (
	"encoding/json"
	"errors"
	"net/http"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/internal/logger"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"
	"unicode"
)

// LINE account linking ("เชื่อมต่อ LINE"): the BFF (frontend/src/app/api/auth/line/*) mints a code at the
// auth bridge, the user confirms it in LINE, then the BFF polls the bridge and PUTs the LINE profile to
// /profile/link-line. The bridge code alone identifies nobody, so the code is bound here to the session
// user who minted it: another signed-in user can neither poll it through the BFF nor complete the link.
const (
	// Longer than the 5-minute polling window of the link dialog (SOCIAL_POLL_TIMEOUT_MS).
	lineLinkCodeTTL       = 10 * time.Minute
	lineLinkCodeMaxLength = 128
	lineLinkCodeKeyPrefix = "linelinkcode:"
)

var (
	errLineLinkCodeInvalid  = errors.New("line link code invalid")
	errLineLinkCodeTaken    = errors.New("line link code is bound to another user")
	errLineLinkCodeNotOwned = errors.New("line link code is not bound to this user")
)

// lineLinkCodeStore is the part of microservice.ICacher (PostgreSQL cache_entries) this flow needs.
type lineLinkCodeStore interface {
	Get(key string) (string, error)
	SetS(key string, value string, expire time.Duration) error
	Del(keys ...string) error
}

type lineLinkCodeRequest struct {
	Code string `json:"code"`
}

func normalizeLineLinkCode(code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" || len(code) > lineLinkCodeMaxLength {
		return "", errLineLinkCodeInvalid
	}
	for _, r := range code {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return "", errLineLinkCodeInvalid
		}
	}
	return code, nil
}

// bindLineLinkCode records code → uid. The first binding wins; the same user may bind again (retry).
// Get-then-Set is not atomic (ICacher has no set-if-absent), so a rival would have to guess a code the
// bridge minted moments earlier and bind it inside that gap.
func bindLineLinkCode(store lineLinkCodeStore, code string, uid string) error {
	code, err := normalizeLineLinkCode(code)
	if err != nil {
		return err
	}
	if strings.TrimSpace(uid) == "" {
		return errLineLinkCodeNotOwned
	}
	key := lineLinkCodeKeyPrefix + code
	owner, err := store.Get(key)
	if err != nil {
		return err
	}
	if owner != "" && owner != uid {
		return errLineLinkCodeTaken
	}
	return store.SetS(key, uid, lineLinkCodeTTL)
}

// checkLineLinkCode passes only for a live code bound to uid; a missing, expired or foreign code gives the
// same error so the answer does not reveal whether someone else holds the code.
func checkLineLinkCode(store lineLinkCodeStore, code string, uid string) error {
	code, err := normalizeLineLinkCode(code)
	if err != nil || strings.TrimSpace(uid) == "" {
		return errLineLinkCodeNotOwned
	}
	owner, err := store.Get(lineLinkCodeKeyPrefix + code)
	if err != nil {
		return err
	}
	if owner != uid {
		return errLineLinkCodeNotOwned
	}
	return nil
}

func releaseLineLinkCode(store lineLinkCodeStore, code string) {
	code, err := normalizeLineLinkCode(code)
	if err != nil {
		return
	}
	if err := store.Del(lineLinkCodeKeyPrefix + code); err != nil {
		// The binding still expires with lineLinkCodeTTL; the link itself already succeeded.
		logger.GetLogger().Warnf("release LINE link code: %v", err)
	}
}

func readLineLinkCodeRequest(ctx microservice.IContext) (lineLinkCodeRequest, error) {
	req := lineLinkCodeRequest{}
	if err := json.Unmarshal([]byte(ctx.ReadInput()), &req); err != nil {
		return req, errLineLinkCodeInvalid
	}
	return req, nil
}

// BindLineLinkCode binds a freshly minted LINE link code to the signed-in user.
// @Description ผูกรหัสเชื่อมต่อ LINE ที่เพิ่งออกกับผู้ใช้ที่เข้าระบบอยู่ (BFF เรียกตอนออก code)
// @Tags		Authentication
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		400 {object}	common.AuthResponseFailed
// @Failure		409 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router		/profile/link-line/code [post]
func (h AuthenticationHttp) BindLineLinkCode(ctx microservice.IContext) error {
	req, err := readLineLinkCodeRequest(ctx)
	if err == nil {
		err = bindLineLinkCode(h.ms.Cacher(), req.Code, ctx.UserInfo().UID)
	}
	if err != nil {
		return respondLineLinkCodeError(ctx, err)
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true})
	return nil
}

// CheckLineLinkCode answers whether a LINE link code belongs to the signed-in user.
// @Description ตรวจว่ารหัสเชื่อมต่อ LINE เป็นของผู้ใช้ที่เข้าระบบอยู่ (BFF เรียกก่อนถาม bridge ทุกรอบ)
// @Tags		Authentication
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		404 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router		/profile/link-line/code/check [post]
func (h AuthenticationHttp) CheckLineLinkCode(ctx microservice.IContext) error {
	req, err := readLineLinkCodeRequest(ctx)
	if err == nil {
		err = checkLineLinkCode(h.ms.Cacher(), req.Code, ctx.UserInfo().UID)
	}
	if err != nil {
		return respondLineLinkCodeError(ctx, err)
	}
	ctx.Response(http.StatusOK, common.ApiResponse{Success: true})
	return nil
}

func respondLineLinkCodeError(ctx microservice.IContext, err error) error {
	var base *apperr.AppError
	var key string
	switch {
	case errors.Is(err, errLineLinkCodeInvalid):
		base, key = apperr.ErrBadRequest, "auth_err_line_link_code_invalid"
	case errors.Is(err, errLineLinkCodeTaken):
		base, key = apperr.ErrConflict, "auth_err_line_link_code_taken"
	case errors.Is(err, errLineLinkCodeNotOwned):
		base, key = apperr.ErrNotFound, "auth_err_line_link_code_not_owned"
	default:
		return apperr.RespondErr(ctx, err)
	}
	return apperr.Respond(ctx, base.WithMessage(language.Text(key, authRequestLanguage(ctx))).
		WithThaiMessage(language.Text(key, "th")))
}
