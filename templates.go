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

func (h *HTTP) setupTemplates() *template.Template {
	var t *template.Template
	if h.Templates == nil || len(h.Templates) == 0 {
		panic("templates are not initialized")
	}
	for _, p := range h.Templates {
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
	if h.TemplateFuncMap == nil {
		return template.Must(t.ParseFS(h.Files, h.Templates...))
	}
	return template.Must(t.Funcs(h.TemplateFuncMap).ParseFS(h.Files, h.Templates...))
}

func (h *HTTP) mergeData(d *TemplateData) {
	d.Title = h.FormatTitle(d.Title)
	if h.DefaultSeoData == nil {
		return
	}
	s := d.SEO
	s.Domain = h.DefaultSeoData.Domain
	s.Title = d.Title
	if s.Image == "" {
		s.Image = h.DefaultSeoData.Image
	}
	if s.Description == "" {
		s.Description = h.DefaultSeoData.Description
	}
}

// Render the template available at templates/page/name.gohtml with the data provided
func (h *HTTP) Render(w http.ResponseWriter, name string, data *TemplateData) {
	h.mergeData(data)
	t := h.setupTemplates()
	template.Must(t.ParseFS(h.Files, h.getFile(name)))
	err := t.ExecuteTemplate(w, h.InitialSection, data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("error while rendering template", "err", err.Error())
	}
}

func (h *HTTP) getFile(path string) string {
	return h.PageDirectory + "/" + path + "." + h.TemplateExtension
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
	// HTTP used
	Http *HTTP
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
func (h *HTTP) NewTemplate(name string, url string, title string, image string, description string, data interface{}) *Template {
	return &Template{
		Http:        h,
		Name:        name,
		Title:       title,
		Data:        data,
		Image:       image,
		Description: description,
		URL:         url,
	}
}

// Handle an http request
func (t *Template) Handle() {
	url := t.URL
	if url == "" {
		url = "/" + t.Name
	}
	t.Http.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		seo := &SeoData{
			URL:         url,
			Description: t.Description,
		}
		if t.Image != "" {
			seo.Image = GetStaticPath(t.Image)
		}
		t.Http.Render(w, t.Name, &TemplateData{
			Title: t.Title,
			SEO:   seo,
			Data:  t.Data,
		})
	})
}

// HandleSimpleTemplate handles an http request for a simple Template (only name and title are present)
func (h *HTTP) HandleSimpleTemplate(name string, title string) {
	t := Template{
		Http:  h,
		Name:  name,
		Title: title,
		Data:  nil,
	}
	t.Handle()
}
