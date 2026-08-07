package apperr

import (
	"smlcloudplatform/pkg/microservice"
)

// Respond writes the AppError as a structured JSON response via the microservice
// context and returns the error for handler propagation (logging).
//
// Usage in handlers:
//
//	if err != nil {
//	    return apperr.Respond(ctx, apperr.ErrNotFound.WithField("code"))
//	}
//
// The JSON response shape is backward-compatible with the existing
// {success: false, message: "..."} format, enriched with errorcode and statuscode.
func Respond(ctx microservice.IContext, appErr *AppError) error {
	ctx.Response(appErr.StatusCode(), appErr.ToResponse())
	return appErr
}

// RespondErr converts a generic error to an AppError (if not already), writes
// the response, and returns it. Useful for wrapping unexpected service errors.
func RespondErr(ctx microservice.IContext, err error) error {
	if appErr := FromError(err); appErr != nil {
		return Respond(ctx, appErr)
	}
	appErr := InternalWrap(err)
	return Respond(ctx, appErr)
}
