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

	"github.com/jamaah-in/v2/internal/finance/handler"
	"github.com/jamaah-in/v2/internal/finance/repository"
	"github.com/jamaah-in/v2/internal/finance/service"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	sharedConfig "github.com/jamaah-in/v2/internal/shared/config"
	sharedDB "github.com/jamaah-in/v2/internal/shared/database"
	sharedHealth "github.com/jamaah-in/v2/internal/shared/health"
	sharedLogger "github.com/jamaah-in/v2/internal/shared/logger"
	sharedMW "github.com/jamaah-in/v2/internal/shared/middleware"
	sharedResponse "github.com/jamaah-in/v2/internal/shared/response"
)

func main() {
	cfg := sharedConfig.Load()
	cfg.Database.DBName = "jamaah_finance"
	cfg.Server.Port = 50055
	if p := os.Getenv("FINANCE_SERVICE_PORT"); p != "" {
		cfg.Server.Port, _ = strconv.Atoi(p)
	}

	cfg.Validate()
	logger := sharedLogger.New(cfg.App.Env)
	sharedResponse.SetLogger(logger)
	logger.Infof("starting finance service on :%d", cfg.Server.Port)

	ctx := context.Background()
	pool, err := sharedDB.Connect(ctx, cfg.Database.DSN())
	if err != nil {
		logger.Fatalf("connect to database: %v", err)
	}
	defer sharedDB.Close(pool)
	logger.Info("connected to database")

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

	financeRepo := repository.NewFinanceRepo(pool)
	invoiceAddr := getEnv("INVOICE_SERVICE_ADDR", "localhost:50054")
	vendorAddr := getEnv("VENDOR_SERVICE_ADDR", "localhost:50057")
	packageAddr := getEnv("PACKAGE_SERVICE_ADDR", "localhost:50052")
	jamaahAddr := getEnv("JAMAAH_SERVICE_ADDR", "localhost:50053")
	financeService := service.NewFinanceService(financeRepo, invoiceAddr, vendorAddr, packageAddr, jamaahAddr)
	financeHandler := handler.NewFinanceHandler(financeService)

	app := fiber.New(fiber.Config{
		AppName:      "jamaah-finance-service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	app.Use(recover.New(), sharedMW.RequestID(), sharedMW.RequestLogger(logger))

	app.Get("/health", sharedHealth.Handler("finance",
		sharedHealth.Check{Name: "database", Ping: pool.Ping}))

	authMW := sharedMW.AuthMiddleware(jwtManager)

	finance := app.Group("/api/v1/finance", authMW, sharedMW.RequireRole("owner", "admin", "finance"))
	expenses := finance.Group("/expenses")
	expenses.Post("/", financeHandler.CreateExpense)
	expenses.Get("/", financeHandler.ListExpenses)
	expenses.Get("/summary", financeHandler.GetSummary)
	expenses.Get("/overdue", financeHandler.GetOverdueExpenses)
	expenses.Get("/package/:pkgId", financeHandler.ListExpensesByPackage)
	expenses.Get("/:id", financeHandler.GetExpense)
	expenses.Put("/:id", financeHandler.UpdateExpense)
	expenses.Delete("/:id", financeHandler.DeleteExpense)

	pnl := finance.Group("/pnl")
	pnl.Get("/:pkgId", financeHandler.GetPnL)

	export := finance.Group("/export")
	export.Get("/pnl", financeHandler.ExportPnL)
	export.Get("/expenses", financeHandler.ExportExpenses)

	dashboard := app.Group("/api/v1/dashboard", authMW)
	dashboard.Get("/owner", financeHandler.GetOwnerDashboard)

	go func() {
		if err := app.Listen(":" + strconv.Itoa(cfg.Server.Port)); err != nil {
			logger.Fatalf("finance service listen: %v", err)
		}
	}()
	logger.Infof("finance service listening on :%d", cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down finance service...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Errorf("finance service shutdown: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
