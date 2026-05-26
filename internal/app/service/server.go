package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	apphttp "market-order-service/internal/app/http"
	"market-order-service/internal/core/usecase"
	"market-order-service/internal/infra/outbound/cartclient"
	"market-order-service/internal/infra/outbound/httpclient"
	"market-order-service/internal/infra/storage/postgres"
)

func RunServer(ctx context.Context) error {
	infra, err := initInfra(ctx)
	if err != nil {
		return fmt.Errorf("init infra: %w", err)
	}

	repo := postgres.NewOrderRepo(infra.DB)

	cartHTTP := httpclient.New(httpclient.Config{
		Target:  "marketplace-bucket",
		BaseURL: infra.Cfg.App.CartServiceURL,
		Metrics: infra.Metrics,
	})
	cart := cartclient.New(cartHTTP)

	srv := apphttp.NewServer(apphttp.Config{
		Addr:         infra.Cfg.HTTP.Addr,
		ReadTimeout:  infra.Cfg.HTTP.ReadTimeout,
		WriteTimeout: infra.Cfg.HTTP.WriteTimeout,
		IdleTimeout:  infra.Cfg.HTTP.IdleTimeout,
		PprofEnabled: infra.Cfg.HTTP.PprofEnabled,
	}, apphttp.Deps{
		Log:          infra.Log,
		Metrics:      infra.Metrics,
		DB:           infra.DB,
		CreateOrder:  usecase.NewCreateOrder(repo, cart),
		GetOrder:     usecase.NewGetOrder(repo),
		ListOrders:   usecase.NewListOrders(repo),
		UpdateStatus: usecase.NewUpdateStatus(repo),
		CancelOrder:  usecase.NewCancelOrder(repo),
	})

	infra.Log.Info("starting server", slog.String("addr", infra.Cfg.HTTP.Addr))

	srvErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			srvErr <- err
		}
	}()

	select {
	case err := <-srvErr:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
	}

	shutCtx, shutCancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer shutCancel()

	infra.Log.Info("shutting down")
	if err := srv.Shutdown(shutCtx); err != nil {
		infra.Log.Error("http shutdown", slog.Any("error", err))
	}
	infra.Shutdown(shutCtx)
	return nil
}
