package apperr

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Middleware returns an Echo middleware that catches *AppError returned from
// handlers and converts them into consistent JSON error responses.
// Non-AppError errors are returned as generic 500 responses.
//
// Usage in bootstrap:
//
//	e.Use(apperr.Middleware())
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			// If the handler already wrote a response, do not overwrite.
			if c.Response().Committed {
				return err
			}

			appErr := FromError(err)
			if appErr != nil {
				return c.JSON(appErr.StatusCode(), appErr.ToResponse())
			}

			// Handle Echo's built-in HTTP errors (e.g. 404 route not found).
			var echoErr *echo.HTTPError
			if ok := asEchoHTTPError(err, &echoErr); ok {
				msg := "request error"
				if m, ok := echoErr.Message.(string); ok {
					msg = m
				}
				return c.JSON(echoErr.Code, Response{
					Success:    false,
					ErrorCode:  http.StatusText(echoErr.Code),
					Message:    msg,
					StatusCode: echoErr.Code,
				})
			}

			// Fallback: unknown error → 500.
			return c.JSON(http.StatusInternalServerError, Response{
				Success:    false,
				ErrorCode:  "INTERNAL_ERROR",
				Message:    "internal server error",
				MessageTH:  "เกิดข้อผิดพลาดในระบบ",
				StatusCode: http.StatusInternalServerError,
			})
		}
	}
}

// asEchoHTTPError is a helper to type-assert echo.HTTPError.
func asEchoHTTPError(err error, target **echo.HTTPError) bool {
	if he, ok := err.(*echo.HTTPError); ok {
		*target = he
		return true
	}
	return false
}

// Handler is a convenience wrapper for Echo handlers that return *AppError.
// It converts the AppError to an echo.HTTPError so the middleware can catch it.
//
// Usage:
//
//	func MyHandler(c echo.Context) error {
//	    result, err := service.DoSomething()
//	    if err != nil {
//	        return apperr.Handler(c, apperr.ErrNotFound.WithField("code"))
//	    }
//	    return c.JSON(200, result)
//	}
func Handler(c echo.Context, appErr *AppError) error {
	return c.JSON(appErr.StatusCode(), appErr.ToResponse())
}
