# Golatt

Golatt is a new Go framework helping the creation of websites!
It is an integration of Go templating in gorilla/mux.

It is lightweight and heavily customizable.

## Usage
### Basic usage
Install the framework with
```bash
go get -u github.com/anhgelus/golatt@v0.2.0
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
You also have to create the directory `templates/page`.
You can modify these names with a config, take a look at the next section to learn how to do it.

Then, you have to create two folders:
- `public` which is the folder containing static files (image, font...)
- `dist` which is the folder containing compiled stuff not in Go (css, javascript...)
These names can also be modified with a config.

Your project directory must look like this:
```
📁 dist/
    | ...
📁 templates/
    | 📁 page/
        | ...
    | ...
📁 public/
    | ...
🗎 .gitignore
🗎 main.go
🗎 go.mod
🗎 go.sum
```

#### Setting up general information
Golatt supports out-of-the-box opengraph.
To use it, you have to set some global SEO information after creating your Golatt instance.
To do it, just create a new `golatt.SeoData` instance.
You can now fill in all required information, i.e. domain, image and description.
Image must be a relative path inside your `public` directory (e.g., if your image is `public/foo/bar.webp`, the path
must be `foo/bar.webp`).
You must put this inside `Golatt.DefaultSeoData`.
```go
g.DefaultSeoData = &golatt.SeoData{
    Domain: "example.org", // as required by opengraph specification
    Image: "foo/bar.webp",
    Description: "An amazing example website!",
}
```

Then, you have to create your own templates.
These templates will be parsed at every request.
It can be components, html templates or what else needed to be loaded at each request.

Let's create a html template. 
In the new file `templates/base/base.gohtml`, you can fill it with your html template, e.g.
```html
{{define "base"}}
    <!DOCTYPE html>
    <html lang="en" prefix="og: https://ogp.me/ns#">
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
You must simply add the relative path of each template into the slice `Golatt.Templates`.
You must NOT register your page body (the template defining the section `body`).
```go
g.Templates = append(g.Templates, "templates/base/*.gohtml")
```
If you changed the name of the folder, you have to change the relative path too!
#### Static paths and assets paths 
Golatt splits files in two categories:
- `static`
- `assets`

Static files are not compiled files by you like images, fonts or whatever.
These are usually placed in the folder `public` and are available with the prefix `/static/`.
You can generate the URL of the file `public/foo/bar.jpg` with `getStaticPath foo/bar.jpg` inside your templates.

Assets files are compiled files by you like css (from scss, less), javascript or whatever.
These are usually placed in the folder `dist` and are available with the prefix `/assets/`.
You can generate the URL of the file `assets/foo/bar.js` with `getAssetPath foo/bar.jpg` inside your templates.

If you want to get a path inside your program, you can use `golatt.GetStaticPath` and `golatt.GetAssetPath`.
#### Simple template
You can handle simple template very easily with `Golatt.HandleSimpleTemplate(name string, title string)`.
- `name` is the name the template desired containing the body section, e.g. `foo/bar` for `templates/page/foo/bar.gohtml`.
These templates must be put in `templates/page` and have the extension `.gohtml`!
(Can be modified.)
- `title` is the title of your page.

You can handle a simple request with this, e.g.
```go
g.HandleSimpleTemplate("hello", "Hello")
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

You can also call `Golatt.NewTemplate` to create a new template.

After, you can call `Template.Handle()` to handle an HTTP request.

```go
t := golatt.Template{
    Golatt: g,
    Name: "index",
    Title: "Home",
    Image: "index.webp",
    Description: "Home of my website!",
    URL: "/",
}
// or
g.NewTemplate("index", "/", "Home", "index.webp", "Home of my website!", nil)

t.Handle()
```
#### Custom render
If you need more customization, you can use `Golatt.Render(w http.ResponseWriter, name string, data *TemplateData)`.
- `w` is the `http.ResponseWriter` to write the response 
- `name` is the name of the template (like the previous name given) 
- `data` are the data to pass to the template, but these are more complicated than the data given in `golatt.Template`.

For example:
```go
g.HandleFunc("/foo", g.Render(w, "foo/index", &golatt.TemplateData{
	Title: "Foo",
	SEO:   &golatt.SeoData{
		URL: "/foo",
		Description: "Foo page!",
		Image: "foo.jpg",
    },
	Data:  nil,
}))
```
#### Errors 404
You can handle errors 404 by setting `Golatt.NotFoundHandler`, e.g.
```go
t := golatt.Template{
    Golatt: g,
    Name: "not_found",
    Title: "Home",
    Image: "",
    Description: "Error 404",
    URL: "",
}
g.NotFoundHandler = t.Handle()
```
### Configuration
You can change default static and assets directories by modifying `AssetsDirectory` and `StaticDirectory` of your `Golatt`
instance.
```go
g.AssetsDirectory = "assets" // default: "dist"
g.StaticDirectory = "static" // default: "public"
```

You can also format each page's title by setting `Golatt.FormatTitle`.
It takes a string representing the page's title, and it returns the new title.
```go
g.FormatTitle = func(t string) string {
	return t + " - Example Website"
} // default: no modification of the title
```

It is also possible to edit the directory containing all your pages' template by modifying `Golatt.PageDirectory`.
```go
// new location is directory/foo/bar (where directory is the default directory in FS, i.e. templates by default)
g.PageDirectory = "foo/bar" // default: "page"
```

You can change the default directory of the filesystem by modifying `Golatt.FsDirectory`.
This value must be the same as the path of the embed directories.
```go
package main

import (
	"embed"

	"github.com/anhgelus/golatt"
)

//go:embed foo/bar
var templates embed.FS

func main() {
	g := golatt.New(templates)
	// sets the default directory to the path of the go:embed FS
	g.FsDirectory = "foo/bar" // default: "templates"
}
```

You can also use another extension for the templates file. 
Modify `Golatt.TemplateExtension` to change it.
```go
// all your template files must have this extension
g.TemplateExtension = "html" // default: "gohtml"
```
## Technologies

- Go 1.23
- [gorilla/mux](https://github.com/gorilla/mux)
