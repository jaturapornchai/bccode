package authentication

import (
	"errors"
	"net/http"

	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/demo"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

// DemoLogin signs the caller in as the public demo account (button "ทดลองใช้ระบบ").
// Public endpoint, no secret: the demo account only ever holds sample data and is
// enabled per environment with BCAI_DEMO_LOGIN_ENABLED (docs/login.md "บัญชี Demo").
func (h AuthenticationHttp) DemoLogin(ctx microservice.IContext) error {
	if !demo.Enabled() {
		return apperr.Respond(ctx, apperr.ErrNotFound.WithMessage("demo login disabled"))
	}
	result, err := h.authenticationService.DemoLoginByUsername(demo.Username(), models.AuthenticationContext{Ip: ctx.RealIp()})
	if err != nil {
		if errors.Is(err, &models.UserDisableLoginError{}) {
			return apperr.Respond(ctx, apperr.ErrDisabled.WithMessage("demo user is disabled"))
		}
		if appErr := apperr.FromError(err); appErr != nil {
			return apperr.Respond(ctx, appErr)
		}
		return apperr.Respond(ctx, apperr.ErrUnauthorized.WithMessage("demo login failed"))
	}
	ctx.Response(http.StatusOK, map[string]interface{}{
		"success": true,
		"token":   result.Token,
		"refresh": result.Refresh,
	})
	return nil
}
