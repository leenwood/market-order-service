package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"market-order-service/internal/core/domain"
)

func writeError(c *gin.Context, err error) {
	switch err {
	case domain.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case domain.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case domain.ErrInvalidTransition, domain.ErrCancelNotAllowed:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case domain.ErrEmptyCart:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
