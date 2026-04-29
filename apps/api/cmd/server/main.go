package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/typewriter-blog/api/internal/config"
	"github.com/typewriter-blog/api/internal/database"
	"github.com/typewriter-blog/api/internal/handler"
	"github.com/typewriter-blog/api/internal/middleware"
	"github.com/typewriter-blog/api/internal/repository"
	"github.com/typewriter-blog/api/internal/service"
	"github.com/typewriter-blog/api/internal/storage"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}

func run() error {
	// ── Config ──────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// ── Database ────────────────────────────────────────────
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db, "migrations"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	// ── Repositories ────────────────────────────────────────
	blogRepo := repository.NewBlogRepository(db)
	adminRepo := repository.NewAdminRepository(db)

	// ── Services ────────────────────────────────────────────
	blogService := service.NewBlogService(blogRepo)
	authService := service.NewAuthService(adminRepo, cfg.JWTSecret)

	// Seed admin account
	if cfg.AdminUsername != "" && cfg.AdminPassword != "" {
		if err := authService.SeedAdmin(context.Background(), cfg.AdminUsername, cfg.AdminPassword); err != nil {
			return fmt.Errorf("seed admin: %w", err)
		}
		log.Println("Admin account ready")
	}

	// ── S3 Storage (optional) ───────────────────────────────
	var imageHandler *handler.ImageHandler
	if cfg.S3AccessKey != "" && cfg.S3SecretKey != "" {
		s3Client, err := storage.NewS3Client(cfg.S3Endpoint, cfg.S3Region, cfg.S3Bucket, cfg.S3AccessKey, cfg.S3SecretKey)
		if err != nil {
			log.Printf("WARN: S3 not configured: %v (image uploads disabled)", err)
		} else {
			imageHandler = handler.NewImageHandler(s3Client)
			log.Println("S3 storage connected")
		}
	}

	// ── Handlers ────────────────────────────────────────────
	blogHandler := handler.NewBlogHandler(blogService)
	authHandler := handler.NewAuthHandler(authService)

	// ── Router ──────────────────────────────────────────────
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   strings.Split(cfg.AllowedOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", handler.HealthCheck)

		// Auth
		r.Post("/auth/login", authHandler.Login)

		// Blogs — public routes (with optional auth to show private posts for admin)
		r.Group(func(r chi.Router) {
			r.Use(middleware.OptionalAuth(authService))
			r.Get("/blogs", blogHandler.List)
			r.Get("/blogs/{slug}", blogHandler.GetBySlug)
		})

		// Blogs — admin-only routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authService))
			r.Post("/blogs", blogHandler.Create)
			r.Put("/blogs/{id}", blogHandler.Update)
			r.Delete("/blogs/{id}", blogHandler.Delete)
		})

		// Image upload — admin-only
		if imageHandler != nil {
			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(authService))
				r.Post("/images/upload", imageHandler.Upload)
			})
		}
	})

	// ── Server ──────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	errCh := make(chan error, 1)
	go func() {
		log.Printf("Server listening on :%s", cfg.Port)
		errCh <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("Received signal %v, shutting down...", sig)
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	log.Println("Server stopped gracefully")
	return nil
}
