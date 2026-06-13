package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	sharedConfig "github.com/jamaah-in/v2/internal/shared/config"
	sharedDB "github.com/jamaah-in/v2/internal/shared/database"
	"github.com/jamaah-in/v2/internal/shared/events"
	sharedHealth "github.com/jamaah-in/v2/internal/shared/health"
	sharedLogger "github.com/jamaah-in/v2/internal/shared/logger"
	sharedMW "github.com/jamaah-in/v2/internal/shared/middleware"
	"github.com/jamaah-in/v2/internal/shared/outbox"
	sharedResponse "github.com/jamaah-in/v2/internal/shared/response"
	"github.com/jamaah-in/v2/internal/tabungan/handler"
	"github.com/jamaah-in/v2/internal/tabungan/repository"
	"github.com/jamaah-in/v2/internal/tabungan/service"
)

func main() {
	cfg := sharedConfig.Load()
	cfg.Database.DBName = "jamaah_tabungan"
	cfg.Server.Port = 50063
	if p := os.Getenv("TABUNGAN_SERVICE_PORT"); p != "" {
		cfg.Server.Port, _ = strconv.Atoi(p)
	}

	cfg.Validate()
	logger := sharedLogger.New(cfg.App.Env)
	sharedResponse.SetLogger(logger)
	logger.Infof("starting tabungan service on :%d", cfg.Server.Port)

	ctx := context.Background()
	pool, err := sharedDB.Connect(ctx, cfg.Database.DSN())
	if err != nil {
		logger.Fatalf("connect to database: %v", err)
	}
	defer sharedDB.Close(pool)
	logger.Info("connected to database")

	var jwtManager *sharedAuth.JWTManager
	if _, err := os.Stat(cfg.JWT.PrivateKeyPath); err == nil {
		jwtManager, err = sharedAuth.NewJWTManager(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath,
			sharedAuth.ParseDuration(cfg.JWT.AccessTTL), sharedAuth.ParseDuration(cfg.JWT.RefreshTTL))
		if err != nil {
			logger.Fatalf("init jwt manager: %v", err)
		}
		logger.Info("JWT manager initialized")
	} else if cfg.App.Env == "production" {
		logger.Fatal("JWT keys not found; refusing to start without auth in production")
	} else {
		logger.Warn("JWT keys not found - running without auth (dev only)")
	}

	repo := repository.NewRepo(pool)
	svc := service.NewService(repo, logger, os.Getenv("INVOICE_SERVICE_ADDR"), cfg.Internal.APIKey)
	h := handler.NewHandler(svc)

	// Integration Bus: outbox relay for savings.deposited / savings.converted.
	if bus, berr := events.Connect(cfg.NATS.Addr, logger); berr != nil {
		logger.Errorf("event bus unavailable (outbox relay disabled): %v", berr)
	} else {
		defer bus.Close()
		go outbox.NewRelay(outbox.NewStore(pool), bus, logger, "tabungan").Start(ctx, 2*time.Second)
	}

	app := fiber.New(fiber.Config{
		AppName:      "jamaah-tabungan-service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	app.Use(recover.New(), sharedMW.RequestID(), sharedMW.RequestLogger(logger))
	app.Get("/health", sharedHealth.Handler("tabungan",
		sharedHealth.Check{Name: "database", Ping: pool.Ping}))

	authMW := sharedMW.AuthMiddleware(jwtManager)
	finRole := sharedMW.RequireRole("owner", "admin", "finance")

	g := app.Group("/api/v1/tabungan", authMW)
	g.Get("/", h.List)
	g.Post("/", finRole, h.Create)
	g.Get("/:id", h.Get)
	g.Post("/:id/deposit", finRole, h.Deposit)
	g.Post("/:id/convert", finRole, h.Convert)

	go func() {
		if err := app.Listen(":" + strconv.Itoa(cfg.Server.Port)); err != nil {
			logger.Fatalf("tabungan service listen: %v", err)
		}
	}()
	logger.Infof("tabungan service listening on :%d", cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down tabungan service...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Errorf("tabungan service shutdown: %v", err)
	}
}
