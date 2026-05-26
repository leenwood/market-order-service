package cartclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"market-order-service/internal/core/domain"
	"market-order-service/internal/core/dto"
	"market-order-service/internal/core/port"
	"market-order-service/internal/infra/outbound/httpclient"
)

type Client struct {
	hc *httpclient.Client
}

func New(hc *httpclient.Client) *Client {
	return &Client{hc: hc}
}

var _ port.CartClient = (*Client)(nil)

func (c *Client) GetCart(ctx context.Context, userID uuid.UUID) (*dto.CartResponse, error) {
	resp, err := c.hc.Do(ctx, http.MethodGet, "/api/v1/cart/"+userID.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("get cart: %w: %w", domain.ErrServiceUnavailable, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return &dto.CartResponse{Items: nil}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get cart: unexpected status %d", resp.StatusCode)
	}

	var cart dto.CartResponse
	if err := json.Unmarshal(resp.Body, &cart); err != nil {
		return nil, fmt.Errorf("decode cart: %w", err)
	}
	return &cart, nil
}

func (c *Client) ClearCart(ctx context.Context, userID uuid.UUID) error {
	resp, err := c.hc.Do(ctx, http.MethodDelete, "/api/v1/cart/"+userID.String(), nil)
	if err != nil {
		return fmt.Errorf("clear cart: %w: %w", domain.ErrServiceUnavailable, err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("clear cart: unexpected status %d", resp.StatusCode)
	}
	return nil
}
