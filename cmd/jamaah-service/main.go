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

	"github.com/jamaah-in/v2/internal/jamaah/handler"
	"github.com/jamaah-in/v2/internal/jamaah/repository"
	"github.com/jamaah-in/v2/internal/jamaah/service"
	sharedAuth "github.com/jamaah-in/v2/internal/shared/auth"
	sharedConfig "github.com/jamaah-in/v2/internal/shared/config"
	sharedDB "github.com/jamaah-in/v2/internal/shared/database"
	"github.com/jamaah-in/v2/internal/shared/events"
	sharedHealth "github.com/jamaah-in/v2/internal/shared/health"
	sharedLogger "github.com/jamaah-in/v2/internal/shared/logger"
	sharedMW "github.com/jamaah-in/v2/internal/shared/middleware"
	sharedNotify "github.com/jamaah-in/v2/internal/shared/notify"
	"github.com/jamaah-in/v2/internal/shared/outbox"
	sharedResponse "github.com/jamaah-in/v2/internal/shared/response"
)

func main() {
	cfg := sharedConfig.Load()
	cfg.Database.DBName = "jamaah_crm"
	cfg.Server.Port = 50053
	if p := os.Getenv("JAMAAH_SERVICE_PORT"); p != "" {
		cfg.Server.Port, _ = strconv.Atoi(p)
	}

	cfg.Validate()
	logger := sharedLogger.New(cfg.App.Env)
	sharedResponse.SetLogger(logger)
	logger.Infof("starting jamaah service on :%d", cfg.Server.Port)

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

	jamaahRepo := repository.NewJamaahRepo(pool)
	jamaahService := service.NewJamaahService(jamaahRepo, os.Getenv("INVOICE_SERVICE_ADDR"), os.Getenv("AUTH_SERVICE_ADDR"), os.Getenv("PACKAGE_SERVICE_ADDR"), os.Getenv("AGENT_SERVICE_ADDR")).
		WithNotify(sharedNotify.New(os.Getenv("AUTH_SERVICE_ADDR"), os.Getenv("INTERNAL_API_KEY"))).
		WithLogger(logger)
	jamaahHandler := handler.NewJamaahHandler(jamaahService)

	// Connect the event bus and start the lead-scoring consumer. If NATS is
	// unreachable we keep serving the CRM API; scores still refresh on in-process
	// mutations and the lazy refresh in ListCRM, just not on payment events.
	var bus *events.Bus
	bus, err = events.Connect(cfg.NATS.Addr, logger)
	if err != nil {
		logger.Errorf("event bus unavailable (scoring consumer disabled): %v", err)
	} else {
		defer bus.Close()
		if serr := jamaahService.StartConsumer(ctx, bus); serr != nil {
			logger.Errorf("start scoring consumer: %v", serr)
		} else {
			logger.Info("jamaah lead-scoring consumer started")
		}
		// Relay visa.* events from the transactional outbox.
		go outbox.NewRelay(outbox.NewStore(pool), bus, logger, "jamaah").Start(ctx, 2*time.Second)
	}

	// Daily passport/visa expiry reminders + auto-expire of lapsed visas.
	jamaahService.StartLifecycleReminders(ctx, 24*time.Hour)

	app := fiber.New(fiber.Config{
		AppName:      "jamaah-crm-service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	app.Use(recover.New(), sharedMW.RequestID(), sharedMW.RequestLogger(logger))

	app.Get("/health", sharedHealth.Handler("jamaah-crm",
		sharedHealth.Check{Name: "database", Ping: pool.Ping}))

	authMW := sharedMW.AuthMiddleware(jwtManager)

	// Staff CRM surface — external agents are confined to /b2b, so block them here
	// (jamaah PII must not be reachable by an agent token via the legacy routes).
	jamaah := app.Group("/api/v1/jamaah", authMW, sharedMW.RequireStaff)
	jamaah.Post("/", jamaahHandler.CreateProfile)
	jamaah.Get("/", jamaahHandler.ListProfiles)
	jamaah.Get("/crm", jamaahHandler.ListCRM)
	// CRM analytics + admin recompute — registered before the "/:id" param route
	// so "pipeline"/"recompute-scores" aren't captured as a jamaah id.
	jamaah.Get("/crm/pipeline", sharedMW.RequireRole("owner", "admin", "cs", "finance"), jamaahHandler.GetPipelineFunnel)
	jamaah.Post("/crm/recompute-scores", sharedMW.RequireRole("owner", "admin"), jamaahHandler.RecomputeScores)
	jamaah.Get("/visa", jamaahHandler.ListVisas) // visa board — before "/:id"
	jamaah.Get("/search/nik/:nik", jamaahHandler.FindByNIK)
	jamaah.Get("/search/paspor/:paspor", jamaahHandler.FindByPaspor)
	jamaah.Get("/dashboard/alerts", jamaahHandler.DashboardAlerts)
	jamaah.Get("/follow-ups", jamaahHandler.ListFollowUps) // literal — before "/:id"
	jamaah.Get("/:id", jamaahHandler.GetProfile)
	jamaah.Put("/:id", jamaahHandler.UpdateProfile)
	jamaah.Delete("/:id", jamaahHandler.DeleteProfile)

	jamaah.Post("/:id/register", jamaahHandler.RegisterToPackage)
	jamaah.Get("/:id/registrations/:pkgId", jamaahHandler.GetRegistration)
	jamaah.Patch("/:id/registrations/:pkgId/status", jamaahHandler.UpdatePipelineStatus)
	jamaah.Patch("/:id/registrations/:pkgId/mahram", jamaahHandler.SetMahram)
	jamaah.Delete("/:id/registrations/:pkgId", jamaahHandler.RemoveFromPackage)

	jamaah.Post("/:id/notes", jamaahHandler.AddNote)
	jamaah.Get("/:id/notes", jamaahHandler.ListNotes)

	jamaah.Post("/:id/follow-ups", jamaahHandler.AddFollowUp)
	jamaah.Patch("/follow-ups/:followUpId/complete", jamaahHandler.CompleteFollowUp)

	jamaah.Post("/:id/documents", jamaahHandler.UploadDocument)
	jamaah.Get("/:id/documents", jamaahHandler.ListDocuments)
	jamaah.Patch("/:id/documents/:docId/status", jamaahHandler.UpdateDocumentStatus)

	jamaah.Get("/:id/visa", jamaahHandler.GetVisa)
	jamaah.Post("/:id/visa", jamaahHandler.UpsertVisa)
	jamaah.Patch("/:id/visa/status", jamaahHandler.TransitionVisa)
	jamaah.Get("/:id/visa/history", jamaahHandler.GetVisaHistory)

	jamaah.Get("/by-package/:pkgId", jamaahHandler.ListByPackage)

	// B2B agent portal: an agent's own + downline leads (scoped to the token's agent).
	b2b := app.Group("/api/v1/b2b", authMW, sharedMW.RequireAgentScope)
	b2b.Get("/leads", jamaahHandler.B2BMyLeads)

	analytics := app.Group("/api/v1/analytics", authMW)
	analytics.Get("/dashboard", jamaahHandler.GetAnalyticsDashboard)

	itineraries := app.Group("/api/v1/itineraries", authMW)
	itineraries.Get("/:groupId", jamaahHandler.GetItinerary)
	itineraries.Post("/:groupId", jamaahHandler.CreateItinerary)
	itineraries.Put("/:groupId/:itemId", jamaahHandler.UpdateItinerary)
	itineraries.Delete("/:groupId/:itemId", jamaahHandler.DeleteItinerary)

	documents := app.Group("/api/v1/documents", authMW)
	documents.Get("/:groupId/:type", jamaahHandler.GetDocumentUrl)

	rooming := app.Group("/api/v1/rooming", authMW)
	rooming.Get("/group/:groupId", jamaahHandler.ListRooms)
	rooming.Post("/group/:groupId", jamaahHandler.CreateRoom)
	rooming.Delete("/:roomId", jamaahHandler.DeleteRoom)
	rooming.Post("/auto/:groupId", jamaahHandler.AutoRooming)
	rooming.Delete("/auto/:groupId", jamaahHandler.ClearAutoRooming)
	rooming.Post("/assign", jamaahHandler.AssignMemberToRoom)
	rooming.Post("/unassign/:memberId", jamaahHandler.UnassignMember)

	shared := app.Group("/api/v1/shared", authMW)
	shared.Post("/groups/:groupId/share", jamaahHandler.ShareGroup)
	shared.Delete("/groups/:groupId/share", jamaahHandler.RevokeShare)

	sharedPub := app.Group("/api/v1/shared")
	sharedPub.Post("/manifest/:token", jamaahHandler.GetSharedManifest)

	groups := app.Group("/api/v1/groups", authMW)
	groups.Get("/", jamaahHandler.ListGroups)
	groups.Post("/", jamaahHandler.CreateGroup)
	groups.Get("/:groupId", jamaahHandler.GetGroup)
	groups.Put("/:groupId", jamaahHandler.UpdateGroup)
	groups.Delete("/:groupId", jamaahHandler.DeleteGroup)
	groups.Post("/:groupId/members", jamaahHandler.AddGroupMembers)
	groups.Put("/:groupId/members/:memberId", jamaahHandler.UpdateGroupMember)
	groups.Delete("/:groupId/members/:memberId", jamaahHandler.DeleteGroupMember)

	// Registration — public endpoints (no auth)
	regPublic := app.Group("/api/v1/registration/public")
	regPublic.Get("/:token", jamaahHandler.PublicRegistrationInfo)
	regPublic.Post("/:token", jamaahHandler.PublicRegistrationSubmit)

	// Registration — admin endpoints (auth required)
	registration := app.Group("/api/v1/registration", authMW)
	registration.Post("/generate", jamaahHandler.GenerateLink)
	registration.Get("/link/:groupId", jamaahHandler.GetActiveLink)
	registration.Delete("/link/:groupId", jamaahHandler.RevokeLink)
	registration.Get("/pending/:groupId", jamaahHandler.ListPendingRegistrations)
	registration.Post("/pending/:pendingId/approve", jamaahHandler.ApproveRegistration)
	registration.Post("/pending/:pendingId/reject", jamaahHandler.RejectRegistration)

	go func() {
		if err := app.Listen(":" + strconv.Itoa(cfg.Server.Port)); err != nil {
			logger.Fatalf("jamaah service listen: %v", err)
		}
	}()
	logger.Infof("jamaah service listening on :%d", cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down jamaah service...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Errorf("jamaah service shutdown: %v", err)
	}
}
