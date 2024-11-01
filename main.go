package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
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

func newDir() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, 10)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s) + "/"
}

func execute(dir string, cmd string) ([]byte, error) {
	os.Chdir(dir)
	res, err := exec.Command(strings.Split(cmd, " ")[0], strings.Split(cmd, " ")[1:]...).Output()
	if err != nil {
		return []byte{}, err
	}
	os.Chdir("..")
	os.RemoveAll(dir)
	return res, nil
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
		path := newDir()
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
		cmd := c.FormValue("cmd")
		res, err := execute(path, cmd)
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, "output", map[string]interface{}{"Text": fmt.Sprintf("%s", res)})
	})
	e.Logger.Fatal(e.Start(":8080"))
}
