package golatt

import (
	"html/template"
	"log/slog"
	"net/http"
)

// SeoData contains seo data used by opengraph and twitter
type SeoData struct {
	// Title of the page (always replaced by TemplateData's title)
	Title string
	// URL of the page
	URL string
	// Image used in embeds
	Image string
	// Description of the page
	Description string
	// Domain of the website (always replaced by Golatt's DefaultSeoData)
	Domain string
}

// TemplateData is passed to the template during the render
type TemplateData struct {
	// Title of the page
	Title string
	// SEO data
	SEO *SeoData
	// Data is custom data passed to the template
	Data interface{}
}

func (g *Golatt) setupTemplates() *template.Template {
	var t *template.Template
	for _, p := range g.Templates {
		t = template.Must(template.ParseFS(g.Files, p))
	}
	return t
}

func (g *Golatt) mergeData(d *TemplateData) {
	d.Title = g.FormatTitle(d.Title)
	if g.DefaultSeoData == nil {
		return
	}
	s := d.SEO
	s.Domain = g.DefaultSeoData.Domain
	s.Title = d.Title
	if s.Image == "" {
		s.Image = g.DefaultSeoData.Image
	}
	if s.Description == "" {
		s.Description = g.DefaultSeoData.Description
	}
}

// Render the template available at templates/page/name.gohtml with the data provided
func (g *Golatt) Render(w http.ResponseWriter, name string, data *TemplateData) {
	g.mergeData(data)
	t := g.setupTemplates()
	template.Must(t.ParseFS(g.Files, g.getFile(name)))
	err := t.ExecuteTemplate(w, g.InitialSection, data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("error while rendering template", "err", err.Error())
	}
}

func (g *Golatt) getFile(path string) string {
	return g.FsDirectory + "/" + g.PageDirectory + "/" + path + "." + g.TemplateExtension
}

func getStaticPath(path string) string {
	return "/static/" + path
}

func getAssetsPath(path string) string {
	return "/assets/" + path
}

// GetStaticPath returns the path of a static file (image, font)
func (d *TemplateData) GetStaticPath(path string) string {
	return getStaticPath(path)
}

// GetAssetsPath returns the path of an assets (js, css)
func (d *TemplateData) GetAssetsPath(path string) string {
	return getAssetsPath(path)
}

// Template represents a generic template
type Template struct {
	// Golatt used
	Golatt *Golatt
	// Name of the template (check Golatt.Render)
	Name string
	// Title of the template
	Title string
	// Data to pass
	Data interface{}
	// Image to use in the SEO
	Image string
	// Description to use in the SEO
	Description string
	// URL of the template
	URL string
}

// Handle a http request
func (t *Template) Handle() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		seo := &SeoData{
			URL:         t.URL,
			Description: t.Description,
		}
		if seo.URL == "" {
			seo.URL = "/" + t.Name
		}
		if t.Image != "" {
			seo.Image = getStaticPath(t.Image)
		}
		t.Golatt.Render(w, t.Name, &TemplateData{
			Title: t.Title,
			SEO:   seo,
			Data:  t.Data,
		})
	}
}

// HandleSimpleTemplate handles a http request for a simple Template (only name and title are present)
func (g *Golatt) HandleSimpleTemplate(name string, title string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t := Template{
			Golatt:      g,
			Name:        name,
			Title:       title,
			Data:        nil,
			Image:       "",
			Description: "",
		}
		t.Handle()
	}
}
