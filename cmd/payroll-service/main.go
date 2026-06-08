package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"

	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	sharedConfig "github.com/jamaah-in/v2/internal/shared/config"
	sharedDB "github.com/jamaah-in/v2/internal/shared/database"
	sharedLogger "github.com/jamaah-in/v2/internal/shared/logger"
	sharedHealth "github.com/jamaah-in/v2/internal/shared/health"
	sharedMW "github.com/jamaah-in/v2/internal/shared/middleware"
	sharedResponse "github.com/jamaah-in/v2/internal/shared/response"

	"github.com/jamaah-in/v2/internal/payroll/handler"
	"github.com/jamaah-in/v2/internal/payroll/repository"
	"github.com/jamaah-in/v2/internal/payroll/service"
)

func main() {
	cfg := sharedConfig.Load()
	cfg.Database.DBName = "jamaah_payroll"
	cfg.Server.Port = 50060
	if p := os.Getenv("PAYROLL_SERVICE_PORT"); p != "" {
		cfg.Server.Port, _ = strconv.Atoi(p)
	}

	cfg.Validate()
	logger := sharedLogger.New(cfg.App.Env)
	sharedResponse.SetLogger(logger)
	logger.Infof("starting payroll service on :%d", cfg.Server.Port)

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

	payrollRepo := repository.NewPayrollRepo(pool)
	payrollSvc := service.NewPayrollService(payrollRepo)
	payrollHandler := handler.NewPayrollHandler(payrollSvc)

	app := fiber.New(fiber.Config{
		AppName:      "jamaah-payroll-service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	app.Use(recover.New(), sharedMW.RequestID(), sharedMW.RequestLogger(logger))

	app.Get("/health", sharedHealth.Handler("payroll",
		sharedHealth.Check{Name: "database", Ping: pool.Ping}))

	authMW := authMiddleware(jwtManager, logger)

	api := app.Group("/api/v1/payroll", authMW)
	api.Get("/summary", payrollHandler.GetSummary)
	api.Post("/employees", payrollHandler.CreateEmployee)
	api.Get("/employees", payrollHandler.ListEmployees)
	api.Get("/employees/:id", payrollHandler.GetEmployee)
	api.Put("/employees/:id", payrollHandler.UpdateEmployee)
	api.Post("/slips", payrollHandler.CreateSalarySlip)
	api.Get("/slips", payrollHandler.ListSalarySlips)
	api.Put("/slips/:id/finalize", payrollHandler.FinalizeSlip)
	api.Get("/slips/:id/pdf", payrollHandler.ExportSlipPDF)
	api.Post("/advances", payrollHandler.CreateAdvance)
	api.Get("/advances", payrollHandler.ListAdvances)
	api.Put("/advances/:id/repay", payrollHandler.RepayAdvance)

	go func() {
		if err := app.Listen(":" + strconv.Itoa(cfg.Server.Port)); err != nil {
			logger.Fatalf("payroll service listen: %v", err)
		}
	}()
	logger.Infof("payroll service listening on :%d", cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down payroll service...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Errorf("payroll service shutdown: %v", err)
	}
}

func authMiddleware(jwtMgr *sharedAuth.JWTManager, logger *zap.SugaredLogger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"success": false, "error": "missing authorization header"})
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			return c.Status(401).JSON(fiber.Map{"success": false, "error": "invalid authorization format"})
		}
		if jwtMgr == nil {
			return c.Status(500).JSON(fiber.Map{"success": false, "error": "JWT not configured"})
		}
		claims, err := jwtMgr.ValidateToken(tokenStr)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"success": false, "error": "invalid or expired token"})
		}
		c.Locals("claims", claims)
		return c.Next()
	}
}
