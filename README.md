# Golatt

Golatt is a new Go framework helping the creation of websites!

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
- SEO Data, templates to parse, initial section
- static and assets paths
- HandleSimpleTemplate
- Template
- Render
### Configuration
- change static and assets directories
- format title
