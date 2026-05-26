package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"market-order-service/internal/core/dto"
	"market-order-service/internal/core/usecase"
)

type OrderHandler struct {
	createOrder  *usecase.CreateOrderUseCase
	getOrder     *usecase.GetOrderUseCase
	listOrders   *usecase.ListOrdersUseCase
	updateStatus *usecase.UpdateStatusUseCase
	cancelOrder  *usecase.CancelOrderUseCase
}

func NewOrderHandler(
	create *usecase.CreateOrderUseCase,
	get *usecase.GetOrderUseCase,
	list *usecase.ListOrdersUseCase,
	update *usecase.UpdateStatusUseCase,
	cancel *usecase.CancelOrderUseCase,
) *OrderHandler {
	return &OrderHandler{
		createOrder:  create,
		getOrder:     get,
		listOrders:   list,
		updateStatus: update,
		cancelOrder:  cancel,
	}
}

func (h *OrderHandler) Create(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid X-User-ID"})
		return
	}

	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.createOrder.Execute(c.Request.Context(), usecase.CreateOrderInput{
		UserID:  userID,
		Request: req,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *OrderHandler) List(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid X-User-ID"})
		return
	}

	resp, err := h.listOrders.Execute(c.Request.Context(), dto.ListOrdersParams{
		UserID:   userID.String(),
		Status:   c.Query("status"),
		Page:     parseIntQuery(c, "page", 1),
		PageSize: parseIntQuery(c, "page_size", 20),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *OrderHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	resp, err := h.getOrder.Execute(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.updateStatus.Execute(c.Request.Context(), usecase.UpdateStatusInput{
		OrderID: id,
		Role:    c.GetHeader("X-User-Role"),
		Request: req,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *OrderHandler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	userID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid X-User-ID"})
		return
	}

	resp, err := h.cancelOrder.Execute(c.Request.Context(), usecase.CancelOrderInput{
		OrderID: id,
		UserID:  userID,
		Role:    c.GetHeader("X-User-Role"),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func parseUserID(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.GetHeader("X-User-ID"))
}

func parseIntQuery(c *gin.Context, key string, fallback int) int {
	if v := c.Query(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
