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

	"github.com/jamaah-in/v2/internal/agent/handler"
	"github.com/jamaah-in/v2/internal/agent/repository"
	"github.com/jamaah-in/v2/internal/agent/service"
)

func main() {
	cfg := sharedConfig.Load()
	cfg.Database.DBName = "jamaah_agent"
	cfg.Server.Port = 50061
	if p := os.Getenv("AGENT_SERVICE_PORT"); p != "" {
		cfg.Server.Port, _ = strconv.Atoi(p)
	}

	cfg.Validate()
	logger := sharedLogger.New(cfg.App.Env)
	sharedResponse.SetLogger(logger)
	logger.Infof("starting agent service on :%d", cfg.Server.Port)

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

	agentRepo := repository.NewAgentRepo(pool)
	agentSvc := service.NewAgentService(agentRepo)

	if bus, berr := events.Connect(cfg.NATS.Addr, logger); berr != nil {
		logger.Errorf("event bus unavailable (outbox relay disabled): %v", berr)
	} else {
		defer bus.Close()
		go outbox.NewRelay(outbox.NewStore(pool), bus, logger, "agent").Start(ctx, 2*time.Second)
	}
	agentHandler := handler.NewAgentHandler(agentSvc)

	app := fiber.New(fiber.Config{
		AppName:      "jamaah-agent-service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	app.Use(recover.New(), sharedMW.RequestID(), sharedMW.RequestLogger(logger))

	app.Get("/health", sharedHealth.Handler("agent",
		sharedHealth.Check{Name: "database", Ping: pool.Ping}))

	authMW := sharedMW.AuthMiddleware(jwtManager)

	// Staff-only management surface — external agents (role "agent") are confined
	// to /b2b. Writes are further restricted to owner/admin (they create GL
	// liabilities via commission.accrued).
	adminOnly := sharedMW.RequireRole("owner", "admin")
	api := app.Group("/api/v1/agents", authMW, sharedMW.RequireStaff)
	api.Get("/", agentHandler.ListAgents)
	api.Post("/", adminOnly, agentHandler.CreateAgent)
	// Tier config — registered before "/:id" so "tiers" isn't read as an id.
	api.Get("/tiers", agentHandler.GetTiers)
	api.Put("/tiers", adminOnly, agentHandler.SetTiers)
	api.Get("/:id", agentHandler.GetAgent)
	api.Put("/:id", adminOnly, agentHandler.UpdateAgent)
	api.Get("/:id/downline", agentHandler.GetDownline)
	api.Get("/:id/upline", agentHandler.GetUpline)

	comm := app.Group("/api/v1/commissions", authMW, sharedMW.RequireStaff)
	comm.Get("/", agentHandler.ListCommissions)
	comm.Post("/", adminOnly, agentHandler.CreateCommission)
	comm.Put("/:id/pay", adminOnly, agentHandler.PayCommission)
	comm.Get("/agent/:id", agentHandler.GetAgentCommissions)

	// B2B external-agent portal: every route is scoped to the signed-in agent's
	// own subtree (RequireAgentScope guarantees a linked agent id).
	b2b := app.Group("/api/v1/b2b", authMW, sharedMW.RequireAgentScope)
	b2b.Get("/me", agentHandler.B2BMe)
	b2b.Get("/dashboard", agentHandler.B2BDashboard)
	b2b.Get("/downline", agentHandler.B2BDownline)
	b2b.Get("/commissions", agentHandler.B2BCommissions)

	go func() {
		if err := app.Listen(":" + strconv.Itoa(cfg.Server.Port)); err != nil {
			logger.Fatalf("agent service listen: %v", err)
		}
	}()
	logger.Infof("agent service listening on :%d", cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down agent service...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Errorf("agent service shutdown: %v", err)
	}
}
