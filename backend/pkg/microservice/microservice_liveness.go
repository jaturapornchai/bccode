package microservice

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (ms *Microservice) isCacherAlive() bool {
	if ms.cacher == nil {
		return true
	}
	return ms.cacher.Healthcheck() == nil
}

func (ms *Microservice) isPostgreSQLAlive() bool {
	for _, pst := range ms.persisters {
		err := pst.TestConnect()
		if err != nil {
			return false
		}
	}
	return true
}

func (ms *Microservice) isAlive() (bool, string) {
	isAlive := ms.isCacherAlive()
	if !isAlive {
		return false, "Cacher healthcheck failed"
	}

	isAlive = ms.isPostgreSQLAlive()
	if !isAlive {
		return false, "PostgreSQL healthcheck failed"
	}

	return true, ""
}

func (ms *Microservice) responseProbeOK(resp *echo.Response) {
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte("ok"))
}

func (ms *Microservice) responseProbeFailed(resp *echo.Response, reason string) {
	errMsg := "Healthcheck failed because of " + reason
	resp.WriteHeader(http.StatusInternalServerError)
	resp.Write([]byte(errMsg))
}

// RegisterLivenessProbeEndpoint register endpoint for liveness probe
func (ms *Microservice) RegisterLivenessProbeEndpoint(path string) {
	ms.echo.GET(path, func(c echo.Context) error {
		ok, reason := ms.isAlive()
		if !ok {
			ms.Logger.Errorf("Failed Check Liveness : %s", reason)
			ms.responseProbeFailed(c.Response(), reason)
			return nil
		}
		ms.responseProbeOK(c.Response())
		return nil
	})
}
