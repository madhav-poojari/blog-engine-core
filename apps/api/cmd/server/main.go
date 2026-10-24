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
	"github.com/typewriter-blog/api/internal/content"
	"github.com/typewriter-blog/api/internal/gitrepo"
	"github.com/typewriter-blog/api/internal/handler"
	"github.com/typewriter-blog/api/internal/middleware"
	"github.com/typewriter-blog/api/internal/service"
	"github.com/typewriter-blog/api/internal/storage"
	websync "github.com/typewriter-blog/api/internal/sync"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// `server hashpw [password]` generates the bcrypt hash for
	// ADMIN_PASSWORD_HASH, so the plaintext never has to live in the env.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "hashpw":
			if err := hashpw(os.Args[2:]); err != nil {
				log.Fatalf("ERROR: %v", err)
			}
			return
		case "check":
			// Index a directory and print what the API would serve from it.
			if err := check(os.Args[2:]); err != nil {
				log.Fatalf("ERROR: %v", err)
			}
			return
		}
	}

	if err := run(); err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}

func hashpw(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: server hashpw <password>")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(args[0]), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	fmt.Printf("ADMIN_PASSWORD_HASH=%s\n", hash)
	return nil
}

func run() error {
	// ── Config ──────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Fail at boot rather than at 3am during the first sweep.
	if err := gitrepo.CheckGitAvailable(); err != nil {
		return err
	}

	// ── Content repo + index ────────────────────────────────
	repo := gitrepo.New(gitrepo.Config{
		URL:        cfg.ContentRepoURL,
		Path:       cfg.ContentRepoPath,
		Branch:     cfg.ContentRepoBranch,
		SSHKeyPath: cfg.ContentSSHKeyPath,
		BotName:    cfg.BotName,
		BotEmail:   cfg.BotEmail,
	})

	builder := &content.Builder{
		Renderer: content.NewRenderer(),
		Options:  content.Options{IgnoredPathTags: cfg.IgnoredPathTagSet()},
	}

	worker := websync.New(repo, builder, websync.Options{
		PollInterval:         cfg.PollInterval,
		DailySweepAt:         cfg.DailySweepAt,
		NormalizeOnEverySync: cfg.NormalizeOnEverySync,
	})

	// The worker runs for the life of the process. It performs the first sync
	// itself, and a failure there is reported as degraded rather than fatal:
	// a GitHub outage must not turn into a crash loop.
	syncCtx, stopSync := context.WithCancel(context.Background())
	defer stopSync()
	go worker.Run(syncCtx)

	// ── Services ────────────────────────────────────────────
	authService := service.NewAuthService(cfg.AdminUsername, cfg.AdminPasswordHash, cfg.JWTSecret)

	// ── S3 Storage (optional) ───────────────────────────────
	var imageHandler *handler.ImageHandler
	if cfg.S3AccessKey != "" && cfg.S3SecretKey != "" {
		s3Client, err := storage.NewS3Client(cfg.S3Endpoint, cfg.S3Region, cfg.S3Bucket,
			cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3PublicBaseURL)
		if err != nil {
			log.Printf("WARN: S3 not configured: %v (image uploads disabled)", err)
		} else {
			imageHandler = handler.NewImageHandler(s3Client)
			log.Println("S3 storage connected")
		}
	}

	// ── Handlers ────────────────────────────────────────────
	blogHandler := handler.NewBlogHandler(worker)
	authHandler := handler.NewAuthHandler(authService)
	syncHandler := handler.NewSyncHandler(worker, cfg.WebhookSecret, cfg.ContentRepoBranch)
	healthHandler := handler.NewHealthHandler(worker)

	// ── Router ──────────────────────────────────────────────
	r := chi.NewRouter()

	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   strings.Split(cfg.AllowedOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-GitHub-Event", "X-Hub-Signature-256"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", healthHandler.Check)

		r.Post("/auth/login", authHandler.Login)

		// GitHub webhook. Authenticated by HMAC, not JWT.
		r.Post("/sync", syncHandler.Webhook)

		// Reads — optional auth, so an admin token reveals private posts.
		r.Group(func(r chi.Router) {
			r.Use(middleware.OptionalAuth(authService))
			r.Get("/blogs", blogHandler.List)
			r.Get("/blogs/{slug}", blogHandler.GetBySlug)
			r.Get("/tags", blogHandler.Tags)
		})

		// Admin-only.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authService))
			r.Post("/admin/sync", syncHandler.SyncNow)
			r.Post("/admin/normalize", syncHandler.NormalizeNow)
			r.Get("/admin/sync/status", syncHandler.Status)

			if imageHandler != nil {
				r.Post("/images/upload", imageHandler.Upload)
			}
		})

		// Posts live in git now; these remain only to give a clear answer.
		r.Post("/blogs", handler.WriteNotSupported)
		r.Put("/blogs/{id}", handler.WriteNotSupported)
		r.Delete("/blogs/{id}", handler.WriteNotSupported)
	})

	// ── Server ──────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

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
