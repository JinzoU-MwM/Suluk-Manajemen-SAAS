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

	"github.com/jamaah-in/v2/internal/aiocr/handler"
	"github.com/jamaah-in/v2/internal/aiocr/repository"
	"github.com/jamaah-in/v2/internal/aiocr/service"
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
	cfg.Database.DBName = "jamaah_aiocr"
	cfg.Server.Port = 50056
	if p := os.Getenv("AIOCR_SERVICE_PORT"); p != "" {
		cfg.Server.Port, _ = strconv.Atoi(p)
	}

	cfg.Validate()
	logger := sharedLogger.New(cfg.App.Env)
	sharedResponse.SetLogger(logger)
	logger.Infof("starting ai/ocr service on :%d", cfg.Server.Port)

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

	aiocrRepo := repository.NewAIOCRRepo(pool)
	analyzer := service.NewAnalyzer(cfg)
	aiocrService := service.NewAIOCRService(aiocrRepo, analyzer, logger).
		WithPolicy(service.NewPolicyExtractor(cfg))
	aiocrHandler := handler.NewAIOCRHandler(aiocrService)

	app := fiber.New(fiber.Config{
		AppName: "jamaah-aiocr-service",
		// Batch OCR can take a while; keep the server from severing the response
		// before a (now-concurrent) multi-file scan finishes. The proxy chain is
		// the real ceiling, but don't let Fiber be the first to time out at 60s.
		ReadTimeout:  180 * time.Second,
		WriteTimeout: 180 * time.Second,
		BodyLimit:    50 * 1024 * 1024,
	})
	app.Use(recover.New(), sharedMW.RequestID(), sharedMW.RequestLogger(logger))

	app.Get("/health", sharedHealth.Handler("aiocr",
		sharedHealth.Check{Name: "database", Ping: pool.Ping}))

	authMW := sharedMW.AuthMiddleware(jwtManager)

	scan := app.Group("/api/v1/scan", authMW, sharedMW.RequireStaff)
	scan.Post("/jobs", aiocrHandler.CreateScanJob)
	scan.Get("/jobs", aiocrHandler.ListScanJobs)
	scan.Get("/jobs/:id", aiocrHandler.GetScanJob)
	scan.Get("/results/:id", aiocrHandler.GetScanResult)
	scan.Get("/jobs/:jobId/results", aiocrHandler.GetScanResultsByJob)

	export := app.Group("/api/v1/export-templates", authMW, sharedMW.RequireStaff)
	export.Post("/", aiocrHandler.CreateExportTemplate)
	export.Get("/", aiocrHandler.ListExportTemplates)
	export.Delete("/:id", aiocrHandler.DeleteExportTemplate)

	ocr := app.Group("/api/v1/ocr", authMW, sharedMW.RequireStaff)
	ocr.Get("/status", aiocrHandler.GetStatus)

	processDocs := app.Group("/api/v1/process-documents", authMW, sharedMW.RequireStaff)
	processDocs.Post("/", aiocrHandler.ProcessDocuments)

	genExcel := app.Group("/api/v1/generate-excel", authMW, sharedMW.RequireStaff)
	genExcel.Post("/", aiocrHandler.GenerateExcel)

	// Service-to-service: auth-service reads an org's monthly scan count to surface
	// it as usage_count on subscription status. Guarded by X-Internal-Key inside
	// the handler, so it is intentionally NOT behind AuthMiddleware.
	app.Post("/api/v1/internal/scan-usage", aiocrHandler.ScanUsageInternal)

	go func() {
		if err := app.Listen(":" + strconv.Itoa(cfg.Server.Port)); err != nil {
			logger.Fatalf("ai/ocr service listen: %v", err)
		}
	}()
	logger.Infof("ai/ocr service listening on :%d", cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down ai/ocr service...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Errorf("ai-ocr service shutdown: %v", err)
	}
}
