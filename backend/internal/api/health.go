// Package api wires the HTTP handlers for the WC-Tournament backend.
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// serviceName identifies this service in health responses.
const serviceName = "wc-tournament"

// Version is the build version, set by main from its -ldflags-injected value
// ("dev" for local builds). Surfaced in the health payload so the running
// release is observable.
var Version = "dev"

// healthResponse is the payload returned by the health endpoints.
type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// RegisterHealthRoutes registers the health endpoints on the given router.
// It exposes both GET /healthz and GET /api/healthz.
func RegisterHealthRoutes(r gin.IRouter) {
	r.GET("/healthz", healthHandler)
	r.GET("/api/healthz", healthHandler)
}

// healthHandler returns a static liveness payload with HTTP 200.
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, healthResponse{
		Status:  "ok",
		Service: serviceName,
		Version: Version,
	})
}
