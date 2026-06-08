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

	"github.com/jamaah-in/v2/internal/invoice/handler"
	"github.com/jamaah-in/v2/internal/invoice/repository"
	"github.com/jamaah-in/v2/internal/invoice/service"
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
	cfg.Database.DBName = "jamaah_invoice"
	cfg.Server.Port = 50054
	if p := os.Getenv("INVOICE_SERVICE_PORT"); p != "" {
		cfg.Server.Port, _ = strconv.Atoi(p)
	}

	cfg.Validate()
	logger := sharedLogger.New(cfg.App.Env)
	sharedResponse.SetLogger(logger)
	logger.Infof("starting invoice service on :%d", cfg.Server.Port)

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

	invoiceRepo := repository.NewInvoiceRepo(pool)
	invoiceService := service.NewInvoiceService(invoiceRepo)
	invoiceHandler := handler.NewInvoiceHandler(invoiceService)

	refundSvc := service.NewRefundService(invoiceRepo)
	refundHandler := handler.NewRefundHandler(refundSvc)

	app := fiber.New(fiber.Config{
		AppName:      "jamaah-invoice-service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	app.Use(recover.New(), sharedMW.RequestID(), sharedMW.RequestLogger(logger))

	app.Get("/health", sharedHealth.Handler("invoice",
		sharedHealth.Check{Name: "database", Ping: pool.Ping}))

	authMW := sharedMW.AuthMiddleware(jwtManager)

	invoices := app.Group("/api/v1/invoices", authMW)
	invoices.Post("/", invoiceHandler.CreateInvoice)
	invoices.Get("/", invoiceHandler.ListInvoices)
	invoices.Get("/summary", invoiceHandler.GetSummary)
	invoices.Get("/revenue/monthly", invoiceHandler.GetMonthlyRevenue)
	invoices.Get("/balances", invoiceHandler.GetBalances)
	invoices.Get("/package/:pkgId/revenue", invoiceHandler.GetPackageRevenue)
	invoices.Get("/package/:pkgId", invoiceHandler.ListByPackage)
	invoices.Get("/number/:number", invoiceHandler.GetInvoiceByNumber)
	invoices.Get("/jamaah/:jamaahId", invoiceHandler.GetInvoicesByJamaah)
	invoices.Get("/:id", invoiceHandler.GetInvoice)
	invoices.Put("/:id", invoiceHandler.UpdateInvoice)
	invoices.Patch("/:id/cancel", invoiceHandler.CancelInvoice)

	invoices.Post("/:id/schedules", invoiceHandler.CreatePaymentSchedules)
	invoices.Get("/:id/schedules", invoiceHandler.GetPaymentSchedules)

	invoices.Post("/:id/payments", invoiceHandler.RecordPayment)
	invoices.Get("/:id/payments", invoiceHandler.GetPayments)
	invoices.Get("/:id/pdf", invoiceHandler.ExportInvoicePDF)

	payment := app.Group("/api/v1/payment", authMW)
	payment.Post("/create-order", invoiceHandler.CreatePaymentOrder)
	payment.Get("/status/:id", invoiceHandler.CheckPaymentStatus)

	refunds := app.Group("/api/v1/refunds", authMW)
	refunds.Get("/policies", refundHandler.ListPolicies)
	refunds.Post("/policies", refundHandler.CreatePolicy)
	refunds.Put("/policies/:id", refundHandler.UpdatePolicy)
	refunds.Delete("/policies/:id", refundHandler.DeletePolicy)
	refunds.Get("/", refundHandler.ListRefunds)
	refunds.Get("/by-invoice/:id", refundHandler.GetRefundsByInvoice)
	refunds.Get("/:id", refundHandler.GetRefund)
	refunds.Put("/:id/approve", refundHandler.ApproveRefund)
	refunds.Put("/:id/process", refundHandler.ProcessRefund)
	refunds.Put("/:id/complete", refundHandler.CompleteRefund)
	refunds.Put("/:id/reject", refundHandler.RejectRefund)
	invoices.Post("/:id/refund", refundHandler.InitiateRefund)

	go func() {
		if err := app.Listen(":" + strconv.Itoa(cfg.Server.Port)); err != nil {
			logger.Fatalf("invoice service listen: %v", err)
		}
	}()
	logger.Infof("invoice service listening on :%d", cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down invoice service...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Errorf("invoice service shutdown: %v", err)
	}
}