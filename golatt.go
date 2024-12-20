package golatt

import (
	"context"
	"github.com/gorilla/mux"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// Golatt is a http server empowered by mux.Router
type Golatt struct {
	// Router used
	*mux.Router
	// Files containing templates used
	Files fs.FS
	// Templates to parse during a request
	Templates []string
	// DefaultSeoData contains all default seo data used by opengraph and twitter
	DefaultSeoData *SeoData
	// InitialSection is the initial section called to render templates.
	// It must be the section containing basic HTML5 structure
	//
	// Default: "base"
	InitialSection string
	// FormatTitle format titles to be more consistant.
	//
	// Default: returns the title without modification
	FormatTitle func(t string) string
	// AssetsDirectory is the folder containing assets
	//
	// Default: "dist"
	AssetsDirectory string
	// StaticDirectory is the folder containing static files
	//
	// Default: "public"
	StaticDirectory string
	// PageDirectory is the folder containing page templates
	//
	// Default: "page"
	PageDirectory string
	// DefaultDirectoryFS is the folder containing all templates
	//
	// Default: "templates"
	FsDirectory string
	// TemplateExtension is the extension of all templates
	//
	// Default: "gohtml"
	TemplateExtension string
	// NotFoundHandler handles 404 errors
	NotFoundHandler func(http.ResponseWriter, *http.Request)
}

// New creates a new Golatt instance with provided files (must be valid go templates files)
func New(files fs.FS) *Golatt {
	return &Golatt{
		Files:  files,
		Router: mux.NewRouter(),
		FormatTitle: func(t string) string {
			return t
		},
		Templates:         make([]string, 0),
		InitialSection:    "base",
		AssetsDirectory:   "dist",
		StaticDirectory:   "public",
		PageDirectory:     "page",
		FsDirectory:       "templates",
		TemplateExtension: "gohtml",
	}
}

// StartServer starts the http server listening on addr (e.g. ":8000", "127.0.0.1:80")
func (g *Golatt) StartServer(addr string) {
	g.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./"+g.StaticDirectory))))
	g.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./"+g.AssetsDirectory))))

	g.Router.NotFoundHandler = http.HandlerFunc(g.NotFoundHandler)

	srv := &http.Server{
		Handler:      g,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	slog.Info("Starting...")
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			slog.Error(err.Error())
		}
	}()

	slog.Info("Started")
	slog.Info("Listening on " + srv.Addr)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		panic(err)
	}
	slog.Info("Shutting down")
}
