package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	sharedConfig "github.com/jamaah-in/v2/internal/shared/config"
	sharedLogger "github.com/jamaah-in/v2/internal/shared/logger"
	sharedMW "github.com/jamaah-in/v2/internal/shared/middleware"
)

var httpClient *http.Client

func main() {
	cfg := sharedConfig.Load()
	gwPort := 8080
	if p := os.Getenv("GATEWAY_PORT"); p != "" {
		gwPort, _ = strconv.Atoi(p)
	}
	cfg.Server.Port = gwPort

	logger := sharedLogger.New(cfg.App.Env)
	logger.Infof("starting API gateway on :%d", cfg.Server.Port)

	httpClient = &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	app := fiber.New(fiber.Config{
		AppName:      "jamaah-api-gateway",
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	})

	app.Use(recover.New(), sharedMW.RequestID(), sharedMW.RequestLogger(logger))
	allowedOrigins := getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:8005,https://test.jni.my.id")
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "jamaah-api-gateway", "version": "2.1.0"})
	})

	services := map[string]string{
		"auth":      getEnv("AUTH_SERVICE_ADDR", "localhost:50051"),
		"package":   getEnv("PACKAGE_SERVICE_ADDR", "localhost:50052"),
		"jamaah":    getEnv("JAMAAH_SERVICE_ADDR", "localhost:50053"),
		"invoice":   getEnv("INVOICE_SERVICE_ADDR", "localhost:50054"),
		"finance":   getEnv("FINANCE_SERVICE_ADDR", "localhost:50055"),
		"aiocr":     getEnv("AIOCR_SERVICE_ADDR", "localhost:50056"),
		"vendor":    getEnv("VENDOR_SERVICE_ADDR", "localhost:50057"),
		"contract":  getEnv("CONTRACT_SERVICE_ADDR", "localhost:50058"),
		"inventory": getEnv("INVENTORY_SERVICE_ADDR", "localhost:50059"),
		"payroll":   getEnv("PAYROLL_SERVICE_ADDR", "localhost:50060"),
		"agent":     getEnv("AGENT_SERVICE_ADDR", "localhost:50061"),
	}

	api := app.Group("/api/v1")

	// Auth service: auth, orgs, invite, and notifications
	setupProxy(api, "/auth", services["auth"])
	setupProxy(api, "/orgs", services["auth"])
	setupProxy(api, "/invite", services["auth"])

	// Package service: auth-protected routes
	setupProxy(api, "/packages", services["package"])

	// Public package page (no auth required)
	setupPublicProxy(app, "/public/packages", services["package"])

	// Jamaah/CRM service
	setupProxy(api, "/jamaah", services["jamaah"])
	setupProxy(api, "/registration", services["jamaah"])
	setupProxy(api, "/analytics", services["jamaah"])

	// Invoice service
	setupProxy(api, "/invoices", services["invoice"])
	setupProxy(api, "/refunds", services["invoice"])

	// Finance service
	setupProxy(api, "/finance", services["finance"])
	setupProxy(api, "/dashboard", services["finance"])

	// Vendor & Biaya Operasional service
	setupProxy(api, "/vendors", services["vendor"])

	// Contract service
	setupProxy(api, "/contracts", services["contract"])
	setupPublicProxy(app, "/public/contracts", services["contract"])

	// Inventory service
	setupProxy(api, "/inventory", services["inventory"])

	// Payroll service
	setupProxy(api, "/payroll", services["payroll"])

	// Agent & Commission service
	setupProxy(api, "/agents", services["agent"])
	setupProxy(api, "/commissions", services["agent"])

	// AI/OCR service: scan jobs/results + export templates
	setupProxy(api, "/scan", services["aiocr"])
	setupProxy(api, "/export-templates", services["aiocr"])

	// Group & Rooming operations
	setupProxy(api, "/groups", services["jamaah"])
	setupProxy(api, "/rooming", services["jamaah"])
	setupProxy(api, "/shared", services["jamaah"])

	// Team, Subscription, Notifications, Support
	setupProxy(api, "/team", services["auth"])
	setupProxy(api, "/subscription", services["auth"])
	setupProxy(api, "/notifications", services["auth"])
	setupProxy(api, "/tickets", services["auth"])

	// Documents & Itineraries
	setupProxy(api, "/documents", services["jamaah"])
	setupProxy(api, "/itineraries", services["jamaah"])

	// Super Admin routes (delegated to ai-ocr for AI cache, auth for users)
	setupProxy(api, "/super-admin", services["aiocr"])

	// Payment & OCR processing
	setupProxy(api, "/payment", services["invoice"])
	setupProxy(api, "/ocr", services["aiocr"])
	setupProxy(api, "/process-documents", services["aiocr"])
	setupProxy(api, "/generate-excel", services["aiocr"])

	go func() {
		if err := app.Listen(":" + strconv.Itoa(cfg.Server.Port)); err != nil {
			logger.Fatalf("gateway listen: %v", err)
		}
	}()
	logger.Infof("API gateway listening on :%d", cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down API gateway...")
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.Shutdown(); err != nil {
		logger.Errorf("gateway shutdown: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func setupProxy(api fiber.Router, prefix string, targetAddr string) {
	handler := createProxyHandler(targetAddr)
	allMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	for _, method := range allMethods {
		path := prefix + "/*"
		switch method {
		case "GET":
			api.Get(path, handler)
		case "POST":
			api.Post(path, handler)
		case "PUT":
			api.Put(path, handler)
		case "PATCH":
			api.Patch(path, handler)
		case "DELETE":
			api.Delete(path, handler)
		}
	}
}

func setupPublicProxy(app *fiber.App, prefix string, targetAddr string) {
	handler := createProxyHandler(targetAddr)
	app.Get(prefix+"/*", handler)
	app.Get(prefix, handler)
	app.Post(prefix+"/*", handler)
	app.Post(prefix, handler)
}

func createProxyHandler(targetAddr string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://" + targetAddr + c.OriginalURL()

		var body io.Reader
		if len(c.Body()) > 0 {
			body = strings.NewReader(string(c.Body()))
		}

		req, err := http.NewRequest(c.Method(), targetURL, body)
		if err != nil {
			return c.Status(502).JSON(fiber.Map{"success": false, "error": "failed to create proxy request"})
		}

		// Forward all relevant headers
		c.Request().Header.VisitAll(func(key, val []byte) {
			k := string(key)
			if k != "Host" && k != "Connection" && k != "Transfer-Encoding" {
				req.Header.Set(k, string(val))
			}
		})
		req.Header.Set("X-Forwarded-For", c.IP())
		req.Header.Set("X-Forwarded-Proto", "http")
		req.Header.Set("X-Real-Ip", c.IP())
		req.Close = true

		resp, err := httpClient.Do(req)
		if err != nil {
			return c.Status(502).JSON(fiber.Map{"success": false, "error": "service unavailable: " + err.Error()})
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(502).JSON(fiber.Map{"success": false, "error": "failed to read response"})
		}

		for k, v := range resp.Header {
			for _, vv := range v {
				c.Set(k, vv)
			}
		}
		return c.Status(resp.StatusCode).Send(respBody)
	}
}
