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

	"github.com/jamaah-in/v2/internal/package/handler"
	"github.com/jamaah-in/v2/internal/package/repository"
	"github.com/jamaah-in/v2/internal/package/service"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	sharedConfig "github.com/jamaah-in/v2/internal/shared/config"
	sharedDB "github.com/jamaah-in/v2/internal/shared/database"
	sharedLogger "github.com/jamaah-in/v2/internal/shared/logger"
	sharedHealth "github.com/jamaah-in/v2/internal/shared/health"
	sharedMW "github.com/jamaah-in/v2/internal/shared/middleware"
	sharedResponse "github.com/jamaah-in/v2/internal/shared/response"
)

func main() {
	cfg := sharedConfig.Load()
	cfg.Database.DBName = "jamaah_package"
	cfg.Server.Port = 50052
	if p := os.Getenv("PACKAGE_SERVICE_PORT"); p != "" {
		cfg.Server.Port, _ = strconv.Atoi(p)
	}

	cfg.Validate()
	logger := sharedLogger.New(cfg.App.Env)
	sharedResponse.SetLogger(logger)
	logger.Infof("starting package service on :%d", cfg.Server.Port)

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

	pkgRepo := repository.NewPackageRepo(pool)
	pkgService := service.NewPackageService(pkgRepo)
	pkgHandler := handler.NewPackageHandler(pkgService)

	app := fiber.New(fiber.Config{
		AppName:      "jamaah-package-service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	app.Use(recover.New(), sharedMW.RequestID(), sharedMW.RequestLogger(logger))

	app.Get("/health", sharedHealth.Handler("package",
		sharedHealth.Check{Name: "database", Ping: pool.Ping}))

	authMW := authMiddleware(jwtManager, logger)

	pkgs := app.Group("/api/v1/packages", authMW)
	pkgs.Post("/", pkgHandler.CreatePackage)
	pkgs.Get("/", pkgHandler.ListPackages)
	pkgs.Get("/:id", pkgHandler.GetPackage)
	pkgs.Put("/:id", pkgHandler.UpdatePackage)
	pkgs.Delete("/:id", pkgHandler.DeletePackage)
	pkgs.Patch("/:id/status", pkgHandler.UpdatePackageStatus)
	pkgs.Get("/:id/quota", pkgHandler.GetPackageQuota)
	pkgs.Get("/:id/profit", pkgHandler.GetProfitProjection)

	pkgs.Post("/:id/tiers", pkgHandler.CreatePricingTier)
	pkgs.Put("/:id/tiers/:tid", pkgHandler.UpdatePricingTier)
	pkgs.Delete("/:id/tiers/:tid", pkgHandler.DeletePricingTier)

	pkgs.Post("/:id/costs", pkgHandler.CreateCostComponent)
	pkgs.Put("/:id/costs/:cid", pkgHandler.UpdateCostComponent)
	pkgs.Delete("/:id/costs/:cid", pkgHandler.DeleteCostComponent)

	app.Get("/public/packages/:slug", pkgHandler.GetPublicPackage)

	go func() {
		if err := app.Listen(":" + strconv.Itoa(cfg.Server.Port)); err != nil {
			logger.Fatalf("package service listen: %v", err)
		}
	}()
	logger.Infof("package service listening on :%d", cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down package service...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Errorf("package service shutdown: %v", err)
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