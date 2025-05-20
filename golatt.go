package golatt

import (
	"context"
	"crypto/tls"
	"crypto/x509/pkix"
	"embed"
	"git.sr.ht/~sotirisp/go-gemini"
	"git.sr.ht/~sotirisp/go-gemini/certificate"
	"github.com/gorilla/mux"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Server interface {
	StartServer(addr string) error
	StopServer() error
}

// Golatt is the main struct of your application
type Golatt struct {
	// HttpMux contains the HTTP server
	HttpMux *HTTP
	// GeminiMux contains the Gemini server
	GeminiMux *Gemini
	// SupportsGemini is true if the gemini protocol is supported by Golatt
	SupportsGemini bool
}

func (g *Golatt) UseGemini() *Golatt {
	g.SupportsGemini = true
	return g
}

// HTTP is an http server empowered by mux.Router
type HTTP struct {
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
	// AssetsFS is the filesystem containing all the assets
	AssetsFS fs.FS
	// StaticFS is the filesystem containing all the static files
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
	srv             *http.Server
}

func (h *HTTP) StartServer(addr string) error {
	h.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServerFS(h.StaticFS)))
	h.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServerFS(h.AssetsFS)))

	h.Router.NotFoundHandler = http.HandlerFunc(h.NotFoundHandler)

	h.srv = &http.Server{
		Handler:      h,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	if err := h.srv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (h *HTTP) StopServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err := h.srv.Shutdown(ctx)
	h.srv = nil
	return err
}

// Gemini is a gemini server empowered by gemini.Mux
type Gemini struct {
	// Gemini mux router
	*gemini.Mux
	// Certs is the certificate.Store
	Certs *certificate.Store
	// CertsDuration is the duration of Certs
	CertsDuration time.Duration
	// StaticFS is the filesystem containing all the static files
	StaticFS fs.FS
	// Domains contain all domains served by Gemini
	Domains []string
	srv     *gemini.Server
}

func (g *Gemini) StartServer(addr string) error {
	g.Handle("/static/", gemini.StripPrefix("/static/", gemini.FileServer(g.StaticFS)))

	g.Certs.CreateCertificate = func(scope string) (tls.Certificate, error) {
		options := certificate.CreateOptions{
			Subject: pkix.Name{
				CommonName: scope,
			},
			DNSNames: []string{scope},
			Duration: g.CertsDuration,
		}
		slog.Info("Creating certificate", "scope", scope, "duration", g.CertsDuration)
		return certificate.Create(options)
	}
	for _, d := range g.Domains {
		g.Certs.Register(d)
	}

	g.srv = &gemini.Server{
		Addr:           addr,
		WriteTimeout:   15 * time.Second,
		ReadTimeout:    15 * time.Second,
		Handler:        g,
		GetCertificate: g.Certs.Get,
	}

	if err := g.srv.ListenAndServe(context.Background()); err != nil {
		return err
	}
	return nil
}

func (g *Gemini) StopServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err := g.srv.Shutdown(ctx)
	g.srv = nil
	return err
}

// New creates a new Golatt instance with provided files (must be valid go templates files)
//
// If you are providing an embed.FS, check UsableEmbedFS before passing this argument
func New(files fs.FS, static fs.FS, assets fs.FS) *Golatt {
	gem := Gemini{
		Mux:           &gemini.Mux{},
		Certs:         &certificate.Store{},
		CertsDuration: 5 * 365 * 24 * time.Hour, // 5 years
		StaticFS:      static,
	}
	htp := HTTP{
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
	return &Golatt{
		HttpMux:   &htp,
		GeminiMux: &gem,
	}
}

// StartServer starts the http server listening on addr (e.g. ":8000", "127.0.0.1:80")
func (g *Golatt) StartServer(addr string) {
	slog.Info("Starting server(s)...")
	go func() {
		if err := g.HttpMux.StartServer(addr); err != nil {
			slog.Error(err.Error())
		}
	}()
	if g.SupportsGemini {
		go func() {
			if err := g.GeminiMux.StartServer(addr); err != nil {
				slog.Error(err.Error())
			}
		}()
	}

	slog.Info("Started")
	slog.Info("Listening on " + addr)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	slog.Info("Shutting down")
	wg := sync.WaitGroup{}
	go func() {
		wg.Add(1)
		if err := g.HttpMux.StopServer(); err != nil {
			panic(err)
		}
		wg.Done()
	}()
	if g.SupportsGemini {
		go func() {
			wg.Add(1)
			if err := g.GeminiMux.StartServer(addr); err != nil {
				panic(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

// httpEmbedFS is an implementation of fs.FS, fs.ReadDirFS and fs.ReadFileFS helping to manage embed.FS for Server
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
