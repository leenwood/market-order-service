// @title           Market Order Service API
// @version         1.0
// @description     Handles order creation, listing, status updates and cancellation.
// @host            localhost:8082
// @BasePath        /
// @schemes         http

// @tag.name         orders
// @tag.description  Order management endpoints

// @tag.name         health
// @tag.description  Health and readiness probes

package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"market-order-service/internal/app/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := service.RunServer(ctx); err != nil {
		log.Fatalf("server: %v", err)
	}
}
