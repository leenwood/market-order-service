package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	internal "market-order-service/internal"
	"market-order-service/internal/infra/storage/postgres"
	"market-order-service/internal/platform/logger"
	"market-order-service/internal/platform/metrics"
	"market-order-service/internal/platform/tracing"
)

type Infra struct {
	Cfg             *internal.Config
	Log             *slog.Logger
	Metrics         *metrics.Metrics
	DB              *postgres.DB
	shutdownTracing tracing.ShutdownFunc
}

func initInfra(ctx context.Context) (*Infra, error) {
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cfg, err := internal.Load()
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	m := metrics.New()

	shutdownTrace, err := tracing.Init(initCtx, tracing.Config{
		Enabled:     cfg.OTel.Enabled,
		Exporter:    cfg.OTel.Exporter,
		Endpoint:    cfg.OTel.Endpoint,
		ServiceName: cfg.OTel.ServiceName,
	})
	if err != nil {
		return nil, fmt.Errorf("tracing: %w", err)
	}

	db, err := postgres.New(initCtx, cfg.Postgres.DSN)
	if err != nil {
		_ = shutdownTrace(initCtx)
		return nil, fmt.Errorf("postgres: %w", err)
	}

	return &Infra{
		Cfg:             cfg,
		Log:             log,
		Metrics:         m,
		DB:              db,
		shutdownTracing: shutdownTrace,
	}, nil
}

func (i *Infra) Shutdown(ctx context.Context) {
	i.DB.Close()
	_ = i.shutdownTracing(ctx)
}
