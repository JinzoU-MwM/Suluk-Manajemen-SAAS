package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/jamaah-in/v2/internal/auth/handler"
	"github.com/jamaah-in/v2/internal/auth/repository"
	"github.com/jamaah-in/v2/internal/auth/service"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	sharedConfig "github.com/jamaah-in/v2/internal/shared/config"
	sharedDB "github.com/jamaah-in/v2/internal/shared/database"
	sharedEmail "github.com/jamaah-in/v2/internal/shared/email"
	sharedHealth "github.com/jamaah-in/v2/internal/shared/health"
	sharedLogger "github.com/jamaah-in/v2/internal/shared/logger"
	sharedMW "github.com/jamaah-in/v2/internal/shared/middleware"
	sharedRedis "github.com/jamaah-in/v2/internal/shared/redis"
	sharedResponse "github.com/jamaah-in/v2/internal/shared/response"
)

func main() {
	cfg := sharedConfig.Load()
	cfg.Database.DBName = "jamaah_auth"
	cfg.Server.Port = 50051
	if p := os.Getenv("AUTH_SERVICE_PORT"); p != "" {
		fmt.Sscan(p, &cfg.Server.Port)
	}

	cfg.Validate()
	logger := sharedLogger.New(cfg.App.Env)
	sharedResponse.SetLogger(logger)
	logger.Infof("starting auth service on :%d", cfg.Server.Port)

	ctx := context.Background()

	pool, err := sharedDB.Connect(ctx, cfg.Database.DSN())
	if err != nil {
		logger.Fatalf("connect to database: %v", err)
	}
	defer sharedDB.Close(pool)
	logger.Info("connected to database")

	rdb, err := sharedRedis.New(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Fatalf("connect to redis: %v", err)
	}
	defer rdb.Close()
	logger.Info("connected to redis")

	accessTTL := sharedAuth.ParseDuration(cfg.JWT.AccessTTL)
	refreshTTL := sharedAuth.ParseDuration(cfg.JWT.RefreshTTL)

	var jwtManager *sharedAuth.JWTManager
	if _, err := os.Stat(cfg.JWT.PrivateKeyPath); err == nil {
		jwtManager, err = sharedAuth.NewJWTManager(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath, accessTTL, refreshTTL)
		if err != nil {
			logger.Fatalf("init jwt manager: %v", err)
		}
		logger.Info("JWT manager initialized with RSA keys")
	} else if cfg.App.Env == "production" {
		logger.Fatal("JWT keys not found; refusing to start without auth in production")
	} else {
		logger.Warn("JWT keys not found, running without JWT validation (dev only)")
		jwtManager = nil
	}

	authRepo := repository.NewAuthRepo(pool)
	authRepo.StartCleanupScheduler(ctx)
	jamaahAddr := os.Getenv("JAMAAH_SERVICE_ADDR")
	invoiceAddr := os.Getenv("INVOICE_SERVICE_ADDR")
	authService := service.NewAuthService(authRepo, jwtManager, rdb, jamaahAddr, invoiceAddr).
		WithEmail(sharedEmail.New(sharedEmail.Config{
			From:         cfg.Email.From,
			SMTPHost:     cfg.Email.SMTPHost,
			SMTPPort:     cfg.Email.SMTPPort,
			SMTPUser:     cfg.Email.SMTPUser,
			SMTPPass:     cfg.Email.SMTPPass,
			ResendAPIKey: cfg.Email.ResendAPIKey,
		}))
	authHandler := handler.NewAuthHandler(authService)

	app := fiber.New(fiber.Config{
		AppName:      "jamaah-auth-service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Use(recover.New(), sharedMW.RequestID(), sharedMW.RequestLogger(logger))

	app.Get("/health", sharedHealth.Handler("auth",
		sharedHealth.Check{Name: "database", Ping: pool.Ping}))

	// Rate limit unauthenticated auth endpoints to slow brute-force/abuse.
	// Key on the real client IP (X-Forwarded-For from the gateway) so the limit is
	// per-client, not collapsed onto the gateway's IP.
	authLimiter := limiter.New(limiter.Config{
		Max:        20,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			if xff := c.Get("X-Forwarded-For"); xff != "" {
				return xff
			}
			return c.IP()
		},
	})

	authPublic := app.Group("/api/v1/auth", authLimiter)
	authPublic.Post("/register", authHandler.Register)
	authPublic.Post("/login", authHandler.Login)
	authPublic.Post("/refresh", authHandler.RefreshToken)

	authPrivate := app.Group("/api/v1/auth", sharedMW.AuthMiddleware(jwtManager))
	authPrivate.Post("/logout", authHandler.Logout)
	authPrivate.Get("/me", authHandler.GetMe)
	authPrivate.Put("/me", authHandler.UpdateMe)
	authPrivate.Post("/change-password", authHandler.ChangePassword)
	authPrivate.Get("/activity", authHandler.GetActivity)
	authPrivate.Delete("/account", authHandler.DeleteAccount)
	authPrivate.Post("/send-phone-otp", authHandler.SendPhoneOtp)
	authPrivate.Post("/verify-phone", authHandler.VerifyPhone)

	authPublic.Post("/verify-email", authHandler.VerifyEmail)
	authPublic.Post("/resend-otp", authHandler.ResendOtp)
	authPublic.Post("/forgot-password", authHandler.ForgotPassword)
	authPublic.Post("/reset-password", authHandler.ResetPassword)

	// Team/branch management is restricted to owner/admin.
	adminRole := sharedMW.RequireRole("owner", "admin")

	orgs := app.Group("/api/v1/orgs", sharedMW.AuthMiddleware(jwtManager))
	orgs.Post("/", authHandler.CreateOrganization)
	orgs.Get("/", authHandler.GetOrganization)
	orgs.Put("/", adminRole, authHandler.UpdateOrganization)
	orgs.Get("/members", authHandler.ListTeamMembers)
	orgs.Get("/users", authHandler.ListUsersByOrg)
	orgs.Post("/members", adminRole, authHandler.AddTeamMember)
	orgs.Delete("/members/:userId", adminRole, authHandler.RemoveTeamMember)
	orgs.Put("/members/:userId/role", adminRole, authHandler.UpdateMemberRole)
	orgs.Post("/invite", adminRole, authHandler.InviteMember)
	orgs.Post("/branches", adminRole, authHandler.CreateBranch)
	orgs.Get("/branches", authHandler.ListBranches)
	orgs.Get("/dashboard/consolidated", authHandler.GetConsolidatedDashboard)

	subscription := app.Group("/api/v1/subscription", sharedMW.AuthMiddleware(jwtManager))
	subscription.Get("/status", authHandler.GetSubscriptionStatus)
	subscription.Post("/upgrade", authHandler.UpgradeToPro)
	subscription.Get("/trial-status", authHandler.GetTrialStatus)
	subscription.Post("/activate-trial", authHandler.ActivateTrial)
	subscription.Get("/pricing", authHandler.GetPricing)

	// Service-to-service: payment webhook activates a paid plan. Guarded by
	// X-Internal-Key inside the handler, so it is intentionally not behind AuthMiddleware.
	app.Post("/api/v1/internal/subscription/activate", authHandler.ActivatePlanInternal)
	// Service-to-service: other services push in-app notifications on key events.
	app.Post("/api/v1/internal/notifications", authHandler.CreateNotificationInternal)

	notifications := app.Group("/api/v1/notifications", sharedMW.AuthMiddleware(jwtManager))
	notifications.Get("/", authHandler.ListNotifications)
	notifications.Put("/:id/read", authHandler.MarkNotificationRead)
	notifications.Put("/read-all", authHandler.MarkAllNotificationsRead)

	tickets := app.Group("/api/v1/tickets", sharedMW.AuthMiddleware(jwtManager))
	tickets.Get("/", authHandler.ListTickets)
	tickets.Post("/", authHandler.CreateTicket)
	tickets.Get("/:id/messages", authHandler.GetTicketMessages)
	tickets.Post("/:id/messages", authHandler.AddTicketMessage)

	team := app.Group("/api/v1/team", sharedMW.AuthMiddleware(jwtManager))
	team.Get("/", authHandler.GetOrganization)
	team.Post("/create", authHandler.CreateOrganization)
	team.Post("/invite", adminRole, authHandler.InviteMember)
	team.Patch("/members/:userId", adminRole, authHandler.UpdateMemberRole)
	team.Delete("/members/:userId", adminRole, authHandler.RemoveTeamMember)
	team.Post("/join/:token", authHandler.AcceptInvite)
	team.Delete("/invites/:inviteId", authHandler.CancelInvite)

	app.Post("/api/v1/invite/accept", sharedMW.AuthMiddleware(jwtManager), authHandler.AcceptInvite)

	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
			logger.Fatalf("fiber listen: %v", err)
		}
	}()
	logger.Infof("auth service listening on :%d", cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down auth service...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Errorf("fiber shutdown: %v", err)
	}
}
