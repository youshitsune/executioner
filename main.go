package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"

	"math/rand"

	"github.com/labstack/echo/v4"
)

type Template struct {
	Templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.Templates.ExecuteTemplate(w, name, data)
}

func newTemplate(templates *template.Template) *Template {
	return &Template{templates}
}

func newTemplateRenderer(e *echo.Echo, paths ...string) {
	tmpl := &template.Template{}
	for i := range paths {
		template.Must(tmpl.ParseGlob(paths[i]))
	}

	t := newTemplate(tmpl)
	e.Renderer = t
}

func nameDir() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, 10)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s) + "/"
}

func execute(dir string) {
	fmt.Println("Executing...")
	os.RemoveAll(dir)
	fmt.Println("Execution finished")
}

func main() {
	e := echo.New()
	e.Static("/static", "static/")
	newTemplateRenderer(e, "templates/*.html")
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", nil)
	})
	e.POST("/upload", func(c echo.Context) error {
		req, _ := c.MultipartForm()
		files := req.File["files"]
		path := nameDir()
		os.Mkdir(path, 0755)
		for _, file := range files {
			src, err := file.Open()
			if err != nil {
				return err
			}
			defer src.Close()
			dst, err := os.Create(path + file.Filename)
			if err != nil {
				return err
			}
			defer dst.Close()

			if _, err = io.Copy(dst, src); err != nil {
				return err
			}

		}
		execute(path)
		return c.NoContent(http.StatusOK)
	})
	e.Logger.Fatal(e.Start(":8080"))
}
