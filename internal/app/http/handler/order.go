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

// Create godoc
// @Summary      Create a new order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        X-User-ID  header    string                  true  "User UUID"
// @Param        body       body      dto.CreateOrderRequest  true  "Order request"
// @Success      201        {object}  dto.OrderResponse
// @Failure      400        {object}  map[string]string
// @Failure      503        {object}  map[string]string
// @Router       /api/v1/orders [post]
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

// List godoc
// @Summary      List orders for a user
// @Tags         orders
// @Produce      json
// @Param        X-User-ID  header    string  true   "User UUID"
// @Param        status     query     string  false  "Filter by status"
// @Param        page       query     int     false  "Page number"     default(1)
// @Param        page_size  query     int     false  "Page size"       default(20)
// @Success      200        {object}  dto.ListOrdersResponse
// @Failure      400        {object}  map[string]string
// @Router       /api/v1/orders [get]
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

// Get godoc
// @Summary      Get order by ID
// @Tags         orders
// @Produce      json
// @Param        X-User-ID  header    string  true  "User UUID"
// @Param        id         path      string  true  "Order UUID"
// @Success      200        {object}  dto.OrderResponse
// @Failure      400        {object}  map[string]string
// @Failure      404        {object}  map[string]string
// @Router       /api/v1/orders/{id} [get]
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

// UpdateStatus godoc
// @Summary      Update order status
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        X-User-ID    header    string                    true  "User UUID"
// @Param        X-User-Role  header    string                    true  "User role (seller|admin)"
// @Param        id           path      string                    true  "Order UUID"
// @Param        body         body      dto.UpdateStatusRequest   true  "New status"
// @Success      200          {object}  dto.OrderResponse
// @Failure      400          {object}  map[string]string
// @Failure      403          {object}  map[string]string
// @Failure      404          {object}  map[string]string
// @Failure      409          {object}  map[string]string
// @Router       /api/v1/orders/{id}/status [patch]
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

// Cancel godoc
// @Summary      Cancel an order
// @Tags         orders
// @Produce      json
// @Param        X-User-ID    header    string  true  "User UUID"
// @Param        X-User-Role  header    string  true  "User role (buyer|seller|admin)"
// @Param        id           path      string  true  "Order UUID"
// @Success      200          {object}  dto.OrderResponse
// @Failure      400          {object}  map[string]string
// @Failure      403          {object}  map[string]string
// @Failure      409          {object}  map[string]string
// @Router       /api/v1/orders/{id}/cancel [post]
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
