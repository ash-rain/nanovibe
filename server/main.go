package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"vibecodepc/server/config"
	"vibecodepc/server/db"
	"vibecodepc/server/routes"
	"vibecodepc/server/services"
)

//go:embed public
var staticFiles embed.FS

const version = "0.1.0"

func main() {
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("vibecodepc %s\n", version)
		os.Exit(0)
	}

	cfg := config.Load()

	// Initialise SQLite database.
	if err := db.Init(cfg.DataDir); err != nil {
		log.Fatalf("Failed to initialise database: %v", err)
	}
	log.Printf("Database initialised at %s", cfg.DataDir)

	// Start Cloudflare quick tunnel in the background (non-fatal if cloudflared not installed).
	go func() {
		time.Sleep(2 * time.Second) // Let the server start first
		if err := services.StartQuickTunnel(); err != nil {
			log.Printf("Cloudflare tunnel not started: %v", err)
		} else {
			log.Printf("Cloudflare tunnel started")
		}
	}()

	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(securityHeaders)

	// CORS — only applied in development
	if cfg.IsDevelopment() {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
	}

	// Register all API and WebSocket routes.
	routes.RegisterSetupRoutes(r)
	routes.RegisterProjectRoutes(r)
	routes.RegisterGitHubRoutes(r)
	routes.RegisterSettingsRoutes(r)
	routes.RegisterAgentRoutes(r)
	routes.RegisterMetricsRoutes(r)
	routes.RegisterTerminalRoutes(r)

	// SPA catch-all: serve embedded client assets for all non-API, non-WS, non-auth paths.
	// This must come after all API routes.
	r.NotFound(spaHandler())

	addr := cfg.Host + ":" + cfg.Port
	log.Printf("VibeCodePC %s listening on http://%s (env: %s)", version, addr, cfg.AppEnv)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 0, // 0 = no timeout (needed for SSE)
		IdleTimeout:  120 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// securityHeaders adds standard security HTTP headers.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// spaHandler returns an http.HandlerFunc that serves the embedded SPA.
// API, WebSocket, and auth paths are excluded — they must be registered first.
func spaHandler() http.HandlerFunc {
	// Unwrap the embed.FS to get a sub-filesystem rooted at "public".
	publicFS, err := fs.Sub(staticFiles, "public")
	if err != nil {
		log.Fatalf("Failed to create sub-FS for public: %v", err)
	}
	fileServer := http.FileServer(http.FS(publicFS))

	return func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path

		// Never serve API, WS, or auth routes through the SPA handler.
		if strings.HasPrefix(urlPath, "/api/") ||
			strings.HasPrefix(urlPath, "/ws/") ||
			strings.HasPrefix(urlPath, "/auth/") {
			http.NotFound(w, r)
			return
		}

		// Clean the path and check if the file exists in the embedded FS.
		cleanPath := path.Clean(urlPath)
		if cleanPath == "/" {
			cleanPath = "/index.html"
		}

		// Try to serve the file as-is.
		f, err := publicFS.Open(strings.TrimPrefix(cleanPath, "/"))
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fall back to index.html for SPA client-side routing.
		r.URL.Path = "/index.html"
		fileServer.ServeHTTP(w, r)
	}
}
