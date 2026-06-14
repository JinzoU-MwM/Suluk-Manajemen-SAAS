package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/jamaah-in/v2/internal/accounting/handler"
	"github.com/jamaah-in/v2/internal/accounting/repository"
	"github.com/jamaah-in/v2/internal/accounting/service"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	sharedConfig "github.com/jamaah-in/v2/internal/shared/config"
	sharedDB "github.com/jamaah-in/v2/internal/shared/database"
	"github.com/jamaah-in/v2/internal/shared/events"
	sharedHealth "github.com/jamaah-in/v2/internal/shared/health"
	sharedLogger "github.com/jamaah-in/v2/internal/shared/logger"
	sharedMW "github.com/jamaah-in/v2/internal/shared/middleware"
	sharedResponse "github.com/jamaah-in/v2/internal/shared/response"
)

func main() {
	backfill := flag.Bool("backfill", false, "run one-off accounting backfill of historical data, then exit")
	flag.Parse()

	cfg := sharedConfig.Load()
	cfg.Database.DBName = "jamaah_accounting"
	cfg.Server.Port = 50062
	if p := os.Getenv("ACCOUNTING_SERVICE_PORT"); p != "" {
		cfg.Server.Port, _ = strconv.Atoi(p)
	}

	cfg.Validate()
	logger := sharedLogger.New(cfg.App.Env)
	sharedResponse.SetLogger(logger)
	logger.Infof("starting accounting service on :%d", cfg.Server.Port)

	ctx := context.Background()
	pool, err := sharedDB.Connect(ctx, cfg.Database.DSN())
	if err != nil {
		logger.Fatalf("connect to database: %v", err)
	}
	defer sharedDB.Close(pool)
	logger.Info("connected to database")

	repo := repository.NewRepo(pool)
	svc := service.NewService(repo, logger)

	if *backfill {
		if err := runBackfill(ctx, cfg, svc, logger); err != nil {
			logger.Fatalf("backfill: %v", err)
		}
		logger.Info("backfill complete")
		return
	}

	var jwtManager *sharedAuth.JWTManager
	if _, err := os.Stat(cfg.JWT.PrivateKeyPath); err == nil {
		accessTTL := sharedAuth.ParseDuration(cfg.JWT.AccessTTL)
		refreshTTL := sharedAuth.ParseDuration(cfg.JWT.RefreshTTL)
		jwtManager, err = sharedAuth.NewJWTManager(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath, accessTTL, refreshTTL)
		if err != nil {
			logger.Fatalf("init jwt manager: %v", err)
		}
		logger.Info("JWT manager initialized")
	} else if cfg.App.Env == "production" {
		logger.Fatal("JWT keys not found; refusing to start without auth in production")
	} else {
		logger.Warn("JWT keys not found - running without auth (dev only)")
	}

	// Connect the event bus and start the journal-posting consumer. If NATS is
	// unreachable we keep serving the read API (reports), and log loudly.
	var bus *events.Bus
	bus, err = events.Connect(cfg.NATS.Addr, logger)
	if err != nil {
		logger.Errorf("event bus unavailable (consumer disabled): %v", err)
	} else {
		defer bus.Close()
		if serr := svc.StartConsumer(ctx, bus); serr != nil {
			logger.Errorf("start consumer: %v", serr)
		} else {
			logger.Info("accounting event consumer started")
		}
	}

	h := handler.NewHandler(svc)

	app := fiber.New(fiber.Config{
		AppName:      "jamaah-accounting-service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	app.Use(recover.New(), sharedMW.RequestID(), sharedMW.RequestLogger(logger))

	app.Get("/health", sharedHealth.Handler("accounting",
		sharedHealth.Check{Name: "database", Ping: pool.Ping}))

	authMW := sharedMW.AuthMiddleware(jwtManager)
	finRole := sharedMW.RequireRole("owner", "admin", "finance")

	coa := app.Group("/api/v1/coa", authMW)
	coa.Get("/", h.ListAccounts)
	coa.Post("/", finRole, h.CreateAccount)

	journals := app.Group("/api/v1/journals", authMW)
	journals.Get("/", h.ListJournals)
	journals.Get("/:id", h.GetJournal)

	reports := app.Group("/api/v1/reports", authMW)
	reports.Get("/trial-balance", h.TrialBalance)
	reports.Get("/neraca", h.BalanceSheet)
	reports.Get("/laba-rugi", h.IncomeStatement)
	reports.Get("/ledger/:accountId", h.GeneralLedger)

	go func() {
		if err := app.Listen(":" + strconv.Itoa(cfg.Server.Port)); err != nil {
			logger.Fatalf("accounting service listen: %v", err)
		}
	}()
	logger.Infof("accounting service listening on :%d", cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down accounting service...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Errorf("accounting service shutdown: %v", err)
	}
}
