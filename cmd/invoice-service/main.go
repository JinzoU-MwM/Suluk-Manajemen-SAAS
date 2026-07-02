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
	"github.com/jamaah-in/v2/internal/shared/events"
	sharedHealth "github.com/jamaah-in/v2/internal/shared/health"
	"github.com/jamaah-in/v2/internal/shared/outbox"
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

	cfg.Validate("INTERNAL_API_KEY", "PAKASIR_API_KEY", "PAKASIR_PROJECT_SLUG")
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
	invoiceService := service.NewInvoiceService(invoiceRepo).WithPayments(service.PaymentDeps{
		Pakasir:     cfg.Pakasir,
		InternalKey: cfg.Internal.APIKey,
		AuthAddr:    os.Getenv("AUTH_SERVICE_ADDR"),
		AiocrAddr:   os.Getenv("AIOCR_SERVICE_ADDR"),
		PublicURL:   cfg.App.PublicURL,
	})
	invoiceHandler := handler.NewInvoiceHandler(invoiceService)

	refundSvc := service.NewRefundService(invoiceRepo).WithPackageAddr(os.Getenv("PACKAGE_SERVICE_ADDR"))
	refundHandler := handler.NewRefundHandler(refundSvc)

	// Integration Bus: start the outbox relay so domain events reach accounting.
	// If NATS is down we keep serving; events queue in the outbox until a relay
	// (next start) drains them.
	if bus, berr := events.Connect(cfg.NATS.Addr, logger); berr != nil {
		logger.Errorf("event bus unavailable (outbox relay disabled): %v", berr)
	} else {
		defer bus.Close()
		relay := outbox.NewRelay(outbox.NewStore(pool), bus, logger, "invoice")
		go relay.Start(ctx, 2*time.Second)
	}

	app := fiber.New(fiber.Config{
		AppName:      "jamaah-invoice-service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	app.Use(recover.New(), sharedMW.RequestID(), sharedMW.RequestLogger(logger))

	app.Get("/health", sharedHealth.Handler("invoice",
		sharedHealth.Check{Name: "database", Ping: pool.Ping}))

	// Internal-only: apply a non-cash credit (tabungan conversion) to an invoice.
	// Guarded by X-Internal-Key inside the handler (no JWT). MUST be registered
	// BEFORE the authenticated /api/v1/invoices group, otherwise that group's
	// AuthMiddleware (mounted on the prefix) intercepts it and returns 401.
	app.Post("/api/v1/invoices/internal/settle", invoiceHandler.SettleInternal)
	// Internal-only: cascades (gagal berangkat, kloter cancel, removed from
	// package) refund exactly what was paid, bypassing the policy percentage
	// cap meant for staff-initiated requests. Same guard/placement as above.
	app.Post("/api/v1/invoices/internal/refund", refundHandler.InitiateRefundInternal)

	authMW := sharedMW.AuthMiddleware(jwtManager)
	// Money writes are restricted to owner/admin/finance; reads stay open to any
	// authenticated role (cs needs to view invoices/balances).
	finRole := sharedMW.RequireRole("owner", "admin", "finance")

	invoices := app.Group("/api/v1/invoices", authMW, sharedMW.RequireStaff)
	invoices.Post("/", finRole, invoiceHandler.CreateInvoice)
	invoices.Get("/", invoiceHandler.ListInvoices)
	invoices.Get("/summary", invoiceHandler.GetSummary)
	invoices.Get("/revenue/monthly", invoiceHandler.GetMonthlyRevenue)
	invoices.Get("/revenue/by-package", invoiceHandler.GetPackageRevenueAll)
	invoices.Get("/balances", invoiceHandler.GetBalances)
	invoices.Get("/package/:pkgId/revenue", invoiceHandler.GetPackageRevenue)
	invoices.Get("/package/:pkgId", invoiceHandler.ListByPackage)
	invoices.Get("/number/:number", invoiceHandler.GetInvoiceByNumber)
	invoices.Get("/jamaah/:jamaahId", invoiceHandler.GetInvoicesByJamaah)
	invoices.Get("/:id", invoiceHandler.GetInvoice)
	invoices.Put("/:id", finRole, invoiceHandler.UpdateInvoice)
	invoices.Patch("/:id/cancel", finRole, invoiceHandler.CancelInvoice)

	invoices.Post("/:id/schedules", finRole, invoiceHandler.CreatePaymentSchedules)
	invoices.Get("/:id/schedules", invoiceHandler.GetPaymentSchedules)

	invoices.Post("/:id/payments", finRole, invoiceHandler.RecordPayment)
	invoices.Get("/:id/payments", invoiceHandler.GetPayments)
	invoices.Get("/:id/pdf", invoiceHandler.ExportInvoicePDF)

	// Public Pakasir webhook (server-to-server, no JWT). Registered before the
	// authenticated payment group so it is not gated by AuthMiddleware.
	app.Post("/api/v1/payment/webhook", invoiceHandler.PakasirWebhook)
	// Public signed subscription-invoice PDF (linked from the confirmation email).
	// Protected by an HMAC sig query param rather than a JWT.
	app.Get("/api/v1/payment/invoice/:orderID", invoiceHandler.SubscriptionInvoicePDF)

	payment := app.Group("/api/v1/payment", authMW, sharedMW.RequireStaff)
	payment.Post("/create-order", invoiceHandler.CreatePaymentOrder)
	payment.Post("/topup-order", invoiceHandler.CreateTopupOrder)
	payment.Get("/status/:id", invoiceHandler.CheckPaymentStatus)

	// Kasir POS — cash drawer sessions (buka/tutup kas).
	sessions := app.Group("/api/v1/cash-sessions", authMW, sharedMW.RequireStaff)
	sessions.Get("/", invoiceHandler.ListCashSessions)
	sessions.Get("/active", invoiceHandler.GetActiveCashSession)
	sessions.Post("/", finRole, invoiceHandler.OpenCashSession)
	sessions.Post("/:id/close", finRole, invoiceHandler.CloseCashSession)

	refunds := app.Group("/api/v1/refunds", authMW, sharedMW.RequireStaff)
	refunds.Get("/policies", refundHandler.ListPolicies)
	refunds.Post("/policies", finRole, refundHandler.CreatePolicy)
	refunds.Put("/policies/:id", finRole, refundHandler.UpdatePolicy)
	refunds.Delete("/policies/:id", finRole, refundHandler.DeletePolicy)
	refunds.Get("/", refundHandler.ListRefunds)
	refunds.Get("/by-invoice/:id", refundHandler.GetRefundsByInvoice)
	refunds.Get("/:id", refundHandler.GetRefund)
	refunds.Put("/:id/approve", finRole, refundHandler.ApproveRefund)
	refunds.Put("/:id/process", finRole, refundHandler.ProcessRefund)
	refunds.Put("/:id/complete", finRole, refundHandler.CompleteRefund)
	refunds.Put("/:id/reject", finRole, refundHandler.RejectRefund)
	invoices.Post("/:id/refund", finRole, refundHandler.InitiateRefund)
	invoices.Get("/:id/refund-policy", refundHandler.GetApplicablePolicy)

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
