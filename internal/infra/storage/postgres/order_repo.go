package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"market-order-service/internal/core/domain"
	"market-order-service/internal/core/dto"
	"market-order-service/internal/core/port"
)

type OrderRepo struct {
	db *DB
}

func NewOrderRepo(db *DB) *OrderRepo {
	return &OrderRepo{db: db}
}

var _ port.OrderRepository = (*OrderRepo)(nil)

func (r *OrderRepo) Create(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx, `
		INSERT INTO orders (id, user_id, status, total_amount, payment_method,
		                    delivery_city, delivery_street, delivery_zip, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		order.ID, order.UserID, order.Status, order.TotalAmount, order.PaymentMethod,
		order.DeliveryAddress.City, order.DeliveryAddress.Street, order.DeliveryAddress.Zip,
		order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (id, order_id, product_id, name, price, quantity, subtotal)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			item.ID, order.ID, item.ProductID, item.Name, item.Price, item.Quantity, item.Subtotal,
		)
		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *OrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	row := r.db.Pool.QueryRow(ctx, `
		SELECT id, user_id, status, total_amount, payment_method,
		       delivery_city, delivery_street, delivery_zip, created_at, updated_at
		FROM orders WHERE id = $1`, id)

	order := &domain.Order{}
	err := row.Scan(
		&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.PaymentMethod,
		&order.DeliveryAddress.City, &order.DeliveryAddress.Street, &order.DeliveryAddress.Zip,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get order: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, order_id, product_id, name, price, quantity, subtotal
		 FROM order_items WHERE order_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Name,
			&item.Price, &item.Quantity, &item.Subtotal); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}
	return order, rows.Err()
}

func (r *OrderRepo) ListByUserID(ctx context.Context, params dto.ListOrdersParams) (*dto.ListOrdersResponse, error) {
	args := []any{params.UserID}
	where := "WHERE user_id = $1"
	argIdx := 2

	if params.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, params.Status)
		argIdx++
	}

	var total int
	if err := r.db.Pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM orders "+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count orders: %w", err)
	}

	offset := (params.Page - 1) * params.PageSize
	args = append(args, params.PageSize, offset)
	rows, err := r.db.Pool.Query(ctx,
		fmt.Sprintf(`SELECT id, user_id, status, total_amount, payment_method,
		       delivery_city, delivery_street, delivery_zip, created_at, updated_at
		FROM orders %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
			where, argIdx, argIdx+1), args...)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.PaymentMethod,
			&o.DeliveryAddress.City, &o.DeliveryAddress.Street, &o.DeliveryAddress.Zip,
			&o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := 0
	if params.PageSize > 0 {
		totalPages = (total + params.PageSize - 1) / params.PageSize
	}

	items := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		items[i] = dto.OrderResponse{
			ID:          o.ID.String(),
			UserID:      o.UserID.String(),
			Status:      string(o.Status),
			TotalAmount: o.TotalAmount,
			DeliveryAddress: dto.DeliveryAddressDTO{
				City:   o.DeliveryAddress.City,
				Street: o.DeliveryAddress.Street,
				Zip:    o.DeliveryAddress.Zip,
			},
			PaymentMethod: o.PaymentMethod,
			CreatedAt:     o.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     o.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &dto.ListOrdersResponse{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	tag, err := r.db.Pool.Exec(ctx,
		`UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
