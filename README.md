# Golatt

Golatt is a new Go framework helping the creation of websites!
It is an integration of Go templating in gorilla/mux.

It is lightweight and heavily customizable.

## Usage
### Basic usage
Install the framework with
```bash
go get -u github.com/anhgelus/golatt
```

Create a new directory called `templates` and embed it with go:embed, e.g.
```go
//go:embed templates
var templates embed.FS
```
This directory will contain all your Go templates.

Create a new `Golatt` instance with `golatt.New(fs.FS)`, e.g.
```go
g := golatt.New(templates)
```

Then you can use this instance to handle http queries, e.g.
```go
g.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Yeah!"))
})
```

And you can start the server with `g.StartServer(string)`, e.g.
```go
g.StartServer(":8000")
```

Full example file:
```go
package main

import (
	"embed"

	"github.com/anhgelus/golatt"
)

//go:embed templates
var templates embed.FS

func main() {
	g := golatt.New(templates)
	g.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Yeah!"))
	})
	g.StartServer(":8000")
}
```
### Templates
Golatt is designed to add Go templates in gorilla/mux.

To use it, you must name the folder containing all your templates `templates` and place it at the root of your project.
You can modify this name with a config, take a look at the next section to learn how to do it.

Then, you have to create two folders:
- `public` which is the folder containing static files (image, font...)
- `dist` which is the folder containing compiled stuff not in Go (css, javascript...)
These names can also be modified with a config.

Your project directory must look like this:
```
dist/
templates/
public/
.gitignore
main.go
go.mod
go.sum
```

#### Setting up general information
Golatt supports out-of-the-box opengraph.
To use it, you have to set some global SEO information after creating your Golatt instance.
To do it, just create a new `golatt.SeoData` instance.
You can now fill in all required informations, i.e. domain, image and description.
Image must be a relative path inside your `public` directory (e.g., if your image is `public/foo/bar.webp`, the path must be `foo/bar.webp`).
```go
seo := golatt.SeoData{
    Domain: "example.org", // as required by opengraph specification
    Image: "foo/bar.webp",
    Description: "An amazing example website!",
}
```

Then, you have to create your own templates.
These templates will be parsed at every request.
It can be components, html templates or what else needed to be loaded at each request.

Let's create an html template. 
In the new file `templates/base/base.gohtml`, you can fill it with your html template, e.g.
```gohtml
{{define "base"}}
    <!DOCTYPE html>
    <html lang="fr" prefix="og: https://ogp.me/ns#">
    <head>
        <meta charset="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>{{ .Title }}</title>
    </head>
    <body>
    {{template "body" .}}
    </body>
    </html>
{{end}}
```
The template "body" will be replaced by your page.
`.Title` refers to the title of the page.

In this example, this initial section is `base`: our templates will be loaded with this entry.
You can change this by setting `Golatt.InitialSection` to any other value.

Finally, we have to register these templates in Golatt.
You must simply add the relative path of each templates into the slice `Golatt.Templates`.
You must NOT register your page body (the template defining the section `body`).
```go
g.Templates = append(g.Templates, "templates/base/base.gohtml")
```
If you changed the name of the folder, you have to change the relative path too!
#### Static paths and assets paths 
Golatt splits files in two categories:
- `static`
- `assets`

Static files are not compiled files by you like images, fonts or whatever.
These are usally placed in the folder `public` and are available with the prefix `/static/`.
You can generate the URL of the file `public/foo/bar.jpg` with `.GetStaticPath foo/bar.jpg` inside your templates.

Assets files are compiled files by you like css (from scss, less), javascript or whatever.
These are usually placed in the folder `dist` and are available with the prefix `/assets/`.
You can generate the URL of the file `assets/foo/bar.js` with `.GetAssetsPath foo/bar.jpg` inside your templates.
#### Simple template
You can handle simple template very easily with `Golatt.HandleSimpleTemplate(name string, title string)`.
- `name` is the name the template desired containing the body section, e.g. `foo/bar` for `templates/page/foo/bar.gohtml`.
These templates must be put in `templates/page` and have the extension `.gohtml`!
(Can be modified.)
- `title` is the title of your page.

You can handle a simple request with this, e.g.
```go
g.HandleFunc("/hello", g.HandleSimpleTemplate("hello", "Hello"))
```
:warning: The URL generated could possibly be the wrong one!
The generated URL will be `"/"+name`.
Here it's `/hello` which is correct, but if you call this with `"index"` to handle `/`, the URL will not be working!
#### Advanced template
Advanced template is like the simple one but with more options.
To use it, you must create a new instance of `golatt.Template`.
- `Golatt` is the current instance of Golatt.
- `Name` and `Title` are the same parameters of `Golatt.HandleSimpleTemplate`.
- `Image` is the image to use for opengraph (see the definition of DefaultSeoData for more information).
- `Description` is the description of the page.
- `Data` is any kind of data that you want to pass to your template. 
This is accessible via `.Data` inside the template.
- `URL` is the URL of the page.
If not set, it will be generated automatically like in simple template.

After, you can call `Template.Handle()` to handle a HTTP request.

```go
t := golatt.Template{
    Golatt: g,
    Name: "index",
    Title: "Home",
    Image: "index.webp",
    Description: "Home of my website!",
    URL: "/",
}
g.HandleFunc("/", t.Handle())
```
#### Custom render
If you need more customization, you can use `Golatt.Render(w http.ResponseWriter, name string, data *TemplateData)`.
- `w` is the `http.ResponseWriter` to write the response 
- `name` is the name of the template (like the previous name given) 
- `data` are the data to pass to the template, but these are more complicated than the data given in `golatt.Template`.

For example:
```
g.HandleFunc("/foo", t.Golatt.Render(w, "foo/index", &TemplateData{
	Title: t.Title,
	SEO:   &golatt.SeoData{
        URL: "/foo",
        Description: "Foo page!",
        Image: "foo.jpg",
    },
	Data:  t.Data,
}))
```
### Configuration
- change static and assets directories
- format title
- change page directory
- change default directory in FS
- change gohtml extension
