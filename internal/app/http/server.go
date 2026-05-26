package apphttp

import (
	"log/slog"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"market-order-service/internal/app/http/handler"
	"market-order-service/internal/app/http/middleware"
	"market-order-service/internal/core/usecase"
	"market-order-service/internal/infra/storage/postgres"
	"market-order-service/internal/platform/metrics"
)

type Config struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	PprofEnabled bool
}

type Deps struct {
	Log          *slog.Logger
	Metrics      *metrics.Metrics
	DB           *postgres.DB
	CreateOrder  *usecase.CreateOrderUseCase
	GetOrder     *usecase.GetOrderUseCase
	ListOrders   *usecase.ListOrdersUseCase
	UpdateStatus *usecase.UpdateStatusUseCase
	CancelOrder  *usecase.CancelOrderUseCase
}

func NewServer(cfg Config, deps Deps) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(
		middleware.GinRecover(deps.Log),
		middleware.GinLogger(deps.Log, deps.Metrics),
		middleware.GinRequestID(),
		middleware.GinMaxBodySize(1<<20),
	)

	healthH := handler.NewHealthHandler(deps.DB)
	orderH := handler.NewOrderHandler(
		deps.CreateOrder, deps.GetOrder, deps.ListOrders,
		deps.UpdateStatus, deps.CancelOrder,
	)

	r.GET("/health", healthH.Live)
	r.GET("/ready", healthH.Ready)

	metricsH := promhttp.HandlerFor(deps.Metrics.Registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
	r.GET("/metrics", gin.WrapH(metricsH))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	v1 := r.Group("/api/v1")
	{
		orders := v1.Group("/orders")
		orders.POST("", orderH.Create)
		orders.GET("", orderH.List)
		orders.GET("/:id", orderH.Get)
		orders.PATCH("/:id/status", orderH.UpdateStatus)
		orders.POST("/:id/cancel", orderH.Cancel)
	}

	if cfg.PprofEnabled {
		r.GET("/debug/pprof/*any", gin.WrapF(pprof.Index))
	}

	otelHandler := otelhttp.NewHandler(r, "order-service",
		otelhttp.WithFilter(func(r *http.Request) bool {
			return r.URL.Path != "/metrics" && r.URL.Path != "/health"
		}),
	)

	return &http.Server{
		Addr:         cfg.Addr,
		Handler:      otelHandler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
