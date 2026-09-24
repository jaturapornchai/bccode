package organization

import (
	"errors"
	"strings"

	"smlcloudplatform/internal/goapi/language"
	orgpolicy "smlcloudplatform/internal/organization/access"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

// accessExpiredKey - ข้อความเมื่อเลยวันหมดอายุการเข้าใช้งาน (ใช้งานได้ถึงสิ้นวันที่กำหนด) — แถวเดียวกับที่ login ใช้
const accessExpiredKey = "user_access_expired"

// AccessExpiredError is the 403 for orgpolicy.ErrAccessExpired with the languages.tsv
// message in lang (Thai in message_th) that tells the user to ask the Holding admin to
// extend or clear the expiry date. nil when err is not an expired membership, so callers
// fall through to their generic membership message.
func AccessExpiredError(err error, lang string) *apperr.AppError {
	if !errors.Is(err, orgpolicy.ErrAccessExpired) {
		return nil
	}
	return apperr.ErrForbidden.WithMessage(language.Text(accessExpiredKey, lang)).WithThaiMessage(language.Text(accessExpiredKey, "th")).WithWrap(err)
}

// RequestLanguage is the caller's language for user-facing messages (?lang= first, then
// Accept-Language) — shared by the organization handlers that have no helper of their own.
func RequestLanguage(ctx microservice.IContext) string {
	if lang := strings.TrimSpace(ctx.QueryParam("lang")); lang != "" {
		return lang
	}
	return ctx.Header("Accept-Language")
}
