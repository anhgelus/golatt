package golatt

import (
	"context"
	"embed"
	"github.com/gorilla/mux"
	"html/template"
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
	// AssetsFS is the filesystem containing all the assets. It must start with AssetsName
	AssetsFS fs.FS
	// StaticFS is the filesystem containing all the static files. It must start with StaticName
	StaticFS fs.FS
	// PageDirectory is the folder containing page templates
	//
	// Default: "page"
	PageDirectory string
	// TemplateExtension is the extension of all templates
	//
	// Default: "gohtml"
	TemplateExtension string
	// NotFoundHandler handles 404 errors
	NotFoundHandler func(http.ResponseWriter, *http.Request)
	// TemplateFuncMap is a map of custom functions usable in templates
	TemplateFuncMap template.FuncMap
}

// New creates a new Golatt instance with provided files (must be valid go templates files)
//
// If you are providing an embed.FS, check UsableEmbedFS before passing this argument
func New(files fs.FS, static fs.FS, assets fs.FS) *Golatt {
	return &Golatt{
		Files:  files,
		Router: mux.NewRouter(),
		FormatTitle: func(t string) string {
			return t
		},
		Templates:         make([]string, 0),
		InitialSection:    "base",
		AssetsFS:          assets,
		StaticFS:          static,
		PageDirectory:     "page",
		TemplateExtension: "gohtml",
		NotFoundHandler:   http.NotFound,
		TemplateFuncMap:   template.FuncMap{},
	}
}

// StartServer starts the http server listening on addr (e.g. ":8000", "127.0.0.1:80")
func (g *Golatt) StartServer(addr string) {
	g.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServerFS(g.StaticFS)))
	g.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServerFS(g.AssetsFS)))

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

// httpEmbedFS is an implementation of fs.FS, fs.ReadDirFS and fs.ReadFileFS helping to manage embed.FS for http server
type httpEmbedFS struct {
	prefix string
	embed.FS
}

func (h *httpEmbedFS) Open(name string) (fs.File, error) {
	return h.FS.Open(h.prefix + "/" + name)
}

func (h *httpEmbedFS) ReadFile(name string) ([]byte, error) {
	return h.FS.ReadFile(h.prefix + "/" + name)
}

func (h *httpEmbedFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return h.FS.ReadDir(h.prefix + "/" + name)
}

// UsableEmbedFS converts embed.FS into usable fs.FS by Golatt
//
// folder may not finish or start with a slash (/)
func UsableEmbedFS(folder string, em embed.FS) fs.FS {
	return &httpEmbedFS{
		prefix: folder,
		FS:     em,
	}
}
