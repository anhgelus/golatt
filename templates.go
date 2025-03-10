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
	if g.Templates == nil || len(g.Templates) == 0 {
		panic("templates are not initialized")
	}
	for _, p := range g.Templates {
		if t == nil {
			t = template.New(p)
		} else {
			t = t.New(p)
		}
	}
	t = t.Funcs(template.FuncMap{
		"getStaticPath": func(path string) string {
			return GetStaticPath(path)
		},
		"getAssetPath": func(path string) string {
			return GetAssetPath(path)
		},
	})
	if g.TemplateFuncMap == nil {
		return template.Must(t.ParseFS(g.Files, g.Templates...))
	}
	return template.Must(t.Funcs(g.TemplateFuncMap).ParseFS(g.Files, g.Templates...))
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
	return g.PageDirectory + "/" + path + "." + g.TemplateExtension
}

// GetStaticPath returns the path of a static file (image, font)
func GetStaticPath(path string) string {
	return "/static/" + path
}

// GetAssetPath returns the path of an asset (js, css)
func GetAssetPath(path string) string {
	return "/assets/" + path
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

// NewTemplate creates a new template.
// You can directly handle it with Template.Handle
func (g *Golatt) NewTemplate(name string, url string, title string, image string, description string, data interface{}) *Template {
	return &Template{
		Golatt:      g,
		Name:        name,
		Title:       title,
		Data:        data,
		Image:       image,
		Description: description,
		URL:         url,
	}
}

// Handle a http request
func (t *Template) Handle() {
	url := t.URL
	if url == "" {
		url = "/" + t.Name
	}
	t.Golatt.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		seo := &SeoData{
			URL:         url,
			Description: t.Description,
		}
		if t.Image != "" {
			seo.Image = GetStaticPath(t.Image)
		}
		t.Golatt.Render(w, t.Name, &TemplateData{
			Title: t.Title,
			SEO:   seo,
			Data:  t.Data,
		})
	})
}

// HandleSimpleTemplate handles a http request for a simple Template (only name and title are present)
func (g *Golatt) HandleSimpleTemplate(name string, title string) {
	t := Template{
		Golatt: g,
		Name:   name,
		Title:  title,
		Data:   nil,
	}
	t.Handle()
}
