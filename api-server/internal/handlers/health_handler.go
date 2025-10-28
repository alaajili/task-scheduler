package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/alaajili/task-scheduler/shared/database"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct{
	db *database.DB
}

func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time": time.Now().UTC().Unix(),
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.HealthCheck(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unready",
			"error":  "database not reachable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"time": time.Now().UTC().Unix(),
	})
}