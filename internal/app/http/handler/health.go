package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type pinger interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	db pinger
}

func NewHealthHandler(db pinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// Live godoc
// @Summary     Liveness probe
// @Tags        health
// @Produce     json
// @Success     200 {object} map[string]string
// @Router      /health [get]
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready godoc
// @Summary     Readiness probe
// @Tags        health
// @Produce     json
// @Success     200 {object} map[string]string
// @Failure     503 {object} map[string]string
// @Router      /ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	if err := h.db.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not ready"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
