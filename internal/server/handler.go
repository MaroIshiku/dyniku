package server

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/MaroIshiku/dyniku/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type handlers struct {
	ctx context.Context //nolint:containedctx
	// Objects
	db            Database
	runner        UpdateForcer
	indexTemplate *template.Template
	auth          *authService
	configPath    string
	dataDir       string
	publicIPLog   string
	buildInfo     models.BuildInformation
	// Mockable functions
	timeNow func() time.Time
}

//go:embed ui/*
var uiFS embed.FS

func newHandler(ctx context.Context, rootURL string,
	db Database, runner UpdateForcer, configPath, dataDir string,
	buildInfo models.BuildInformation,
) http.Handler {
	indexTemplate := template.Must(template.ParseFS(uiFS, "ui/index.html"))

	staticFolder, err := fs.Sub(uiFS, "ui/static")
	if err != nil {
		panic(err)
	}

	handlers := &handlers{
		ctx:           ctx,
		db:            db,
		indexTemplate: indexTemplate,
		auth:          newAuthService(dataDir, buildInfo),
		configPath:    configPath,
		dataDir:       dataDir,
		publicIPLog:   filepath.Join(dataDir, "public-ip.log"),
		buildInfo:     buildInfo,
		timeNow:       time.Now,
		runner:        runner,
	}

	router := chi.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	rootURL = strings.TrimSuffix(rootURL, "/")

	if rootURL != "" {
		router.Handle(rootURL, http.RedirectHandler(rootURL+"/", http.StatusPermanentRedirect))
	}
	router.Get("/healthz", handlers.healthz)
	router.Get("/readyz", handlers.readyz)
	if rootURL != "" {
		router.Get(rootURL+"/healthz", handlers.healthz)
		router.Get(rootURL+"/readyz", handlers.readyz)
	}
	router.Get(rootURL+"/", handlers.index)
	router.Get(rootURL+"/login", handlers.index)
	router.Get(rootURL+"/setup", handlers.setupPage)

	router.Get(rootURL+"/api/auth/state", handlers.authState)
	router.Post(rootURL+"/api/setup", handlers.setup)
	router.Post(rootURL+"/api/login", handlers.login)
	router.Post(rootURL+"/api/logout", handlers.logout)

	router.Group(func(protected chi.Router) {
		protected.Use(handlers.requireAuth)
		protected.Get(rootURL+"/update", handlers.update)
		protected.Post(rootURL+"/update", handlers.update)
		protected.Get(rootURL+"/api/me", handlers.me)
		protected.Put(rootURL+"/api/account", handlers.updateAccount)
		protected.Get(rootURL+"/api/admin", handlers.adminInfo)
		protected.Get(rootURL+"/api/status", handlers.status)
		protected.Get(rootURL+"/api/config", handlers.getConfig)
		protected.Put(rootURL+"/api/config", handlers.putConfig)
	})

	router.Handle(rootURL+"/static/*", http.StripPrefix(rootURL+"/static/", http.FileServerFS(staticFolder)))

	return router
}
