//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"market-order-service/internal/core/domain"
	"market-order-service/internal/core/dto"
	pgstore "market-order-service/internal/infra/storage/postgres"
)

func setupDB(t *testing.T) *pgstore.DB {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("orders"),
		tcpostgres.WithUsername("orders"),
		tcpostgres.WithPassword("orders"),
		tcpostgres.WithInitScripts(
			"../../../migrations/postgres/00001_create_orders.sql",
			"../../../migrations/postgres/00002_create_order_items.sql",
		),
		tcpostgres.WithWaitStrategyAndDeadline(
			30*time.Second,
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	db, err := pgstore.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(db.Close)
	return db
}

func TestOrderRepo_CreateAndGetByID(t *testing.T) {
	db := setupDB(t)
	repo := pgstore.NewOrderRepo(db)
	ctx := context.Background()

	uid := uuid.New()
	order := &domain.Order{
		ID:            uuid.New(),
		UserID:        uid,
		Status:        domain.StatusPending,
		TotalAmount:   199.99,
		PaymentMethod: "card",
		DeliveryAddress: domain.DeliveryAddress{City: "Moscow", Street: "St 1", Zip: "100000"},
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		Items: []domain.OrderItem{
			{ID: uuid.New(), ProductID: uuid.New(), Name: "Widget", Price: 199.99, Quantity: 1, Subtotal: 199.99},
		},
	}

	if err := repo.Create(ctx, order); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != order.ID {
		t.Errorf("ID mismatch")
	}
	if len(got.Items) != 1 {
		t.Errorf("items count: want 1 got %d", len(got.Items))
	}
}

func TestOrderRepo_ListByUserID_Pagination(t *testing.T) {
	db := setupDB(t)
	repo := pgstore.NewOrderRepo(db)
	ctx := context.Background()

	uid := uuid.New()
	for i := 0; i < 3; i++ {
		o := &domain.Order{
			ID: uuid.New(), UserID: uid, Status: domain.StatusPending,
			TotalAmount: 10, PaymentMethod: "card",
			DeliveryAddress: domain.DeliveryAddress{City: "C", Street: "S", Zip: "Z"},
			CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		}
		if err := repo.Create(ctx, o); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	resp, err := repo.ListByUserID(ctx, dto.ListOrdersParams{UserID: uid.String(), Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if resp.Total != 3 {
		t.Errorf("total: want 3 got %d", resp.Total)
	}
	if len(resp.Items) != 2 {
		t.Errorf("items: want 2 got %d", len(resp.Items))
	}
	if resp.TotalPages != 2 {
		t.Errorf("total_pages: want 2 got %d", resp.TotalPages)
	}
}

func TestOrderRepo_UpdateStatus_NotFound(t *testing.T) {
	db := setupDB(t)
	repo := pgstore.NewOrderRepo(db)
	ctx := context.Background()

	err := repo.UpdateStatus(ctx, uuid.New(), domain.StatusCancelled)
	if err != domain.ErrNotFound {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}
