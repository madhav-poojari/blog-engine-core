package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	JWTSecret      string
	AllowedOrigins string

	// Admin credentials. The password is supplied as a bcrypt hash so the
	// plaintext never sits in the environment; `server hashpw` generates one.
	AdminUsername     string
	AdminPasswordHash string

	// Content repo.
	ContentRepoURL    string
	ContentRepoPath   string
	ContentRepoBranch string
	ContentSSHKeyPath string
	BotName           string
	BotEmail          string

	// Sync behaviour.
	WebhookSecret        string
	PollInterval         time.Duration
	DailySweepAt         string
	NormalizeOnEverySync bool
	IgnoredPathTags      []string

	// Object storage for images (optional).
	S3Endpoint      string
	S3Region        string
	S3Bucket        string
	S3AccessKey     string
	S3SecretKey     string
	S3PublicBaseURL string
}

func Load() (*Config, error) {
	// .env is optional — production uses real env vars
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),

		AdminUsername:     getEnv("ADMIN_USERNAME", "admin"),
		AdminPasswordHash: getEnv("ADMIN_PASSWORD_HASH", ""),

		ContentRepoURL:    getEnv("CONTENT_REPO_URL", ""),
		ContentRepoPath:   getEnv("CONTENT_REPO_PATH", "./content-repo"),
		ContentRepoBranch: getEnv("CONTENT_REPO_BRANCH", "main"),
		ContentSSHKeyPath: getEnv("CONTENT_SSH_KEY_PATH", ""),
		BotName:           getEnv("CONTENT_BOT_NAME", "blog-engine-bot"),
		BotEmail:          getEnv("CONTENT_BOT_EMAIL", "bot@blog-engine.local"),

		WebhookSecret:        getEnv("GITHUB_WEBHOOK_SECRET", ""),
		PollInterval:         getDuration("SYNC_POLL_INTERVAL", 5*time.Minute),
		DailySweepAt:         getEnv("SYNC_DAILY_SWEEP_AT", "03:00"),
		NormalizeOnEverySync: getBool("NORMALIZE_ON_EVERY_SYNC", false),
		IgnoredPathTags:      splitList(getEnv("IGNORED_PATH_TAGS", "uncategorised")),

		S3Endpoint:      getEnv("S3_ENDPOINT", ""),
		S3Region:        getEnv("S3_REGION", ""),
		S3Bucket:        getEnv("S3_BUCKET", ""),
		S3AccessKey:     getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:     getEnv("S3_SECRET_KEY", ""),
		S3PublicBaseURL: strings.TrimRight(getEnv("S3_PUBLIC_BASE_URL", ""), "/"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.AdminPasswordHash == "" {
		return nil, fmt.Errorf("ADMIN_PASSWORD_HASH is required (generate one with: server hashpw)")
	}
	if cfg.ContentRepoURL == "" {
		return nil, fmt.Errorf("CONTENT_REPO_URL is required")
	}
	if cfg.ContentRepoPath == "" {
		return nil, fmt.Errorf("CONTENT_REPO_PATH is required")
	}

	return cfg, nil
}

// IgnoredPathTagSet renders the ignore list as a lookup set.
func (c *Config) IgnoredPathTagSet() map[string]bool {
	set := make(map[string]bool, len(c.IgnoredPathTags))
	for _, t := range c.IgnoredPathTags {
		if t != "" {
			set[strings.ToLower(t)] = true
		}
	}
	return set
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	v, err := strconv.ParseBool(getEnv(key, ""))
	if err != nil {
		return fallback
	}
	return v
}

func getDuration(key string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(getEnv(key, ""))
	if err != nil {
		return fallback
	}
	return d
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
