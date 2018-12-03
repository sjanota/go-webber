package view

import (
	"fmt"
	"github.com/oxtoacart/bpool"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

const (
	templateBasePath   = "templates"
	templateLayoutPath = "layouts"
	templatePagesPath  = "pages"
	mainTmpl           = `{{define "main" }} {{ template "base" . }} {{ end }}`
)

type Renderer interface {
	Render(writer http.ResponseWriter, tplName string, data interface{}) error
	RenderIndex(writer http.ResponseWriter) error
}

type renderer struct {
	templates  map[string]*template.Template
	bufferPool *bpool.BufferPool
}

func New(rootPath string) (Renderer, error) {
	v := &renderer{}
	v.templates = make(map[string]*template.Template)

	layoutPattern := filepath.Join(rootPath, templateBasePath, templateLayoutPath, "*.gohtml")
	layoutFiles, err := filepath.Glob(layoutPattern)
	if err != nil {
		return nil, err
	}
	log.Printf("Using layouts: %v", layoutFiles)

	pagePattern := filepath.Join(rootPath, templateBasePath, templatePagesPath, "*.gohtml")
	pageFiles, err := filepath.Glob(pagePattern)
	if err != nil {
		return nil, err
	}
	log.Printf("Loading pages: %v", pageFiles)

	mainTemplate := template.New("main")
	mainTemplate, err = mainTemplate.Parse(mainTmpl)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range pageFiles {
		fileName := filepath.Base(file)
		files := append(layoutFiles, file)

		v.templates[fileName], err = mainTemplate.Clone()
		if err != nil {
			return nil, err
		}

		v.templates[fileName], err = v.templates[fileName].ParseFiles(files...)
		if err != nil {
			return nil, err
		}

		log.Printf("Page %s loaded", fileName)
	}

	v.bufferPool = bpool.NewBufferPool(64)
	return v, nil
}


func (r renderer) Render(writer http.ResponseWriter, name string, data interface{}) error {
	tmpl, ok := r.templates[name]
	if !ok {
		http.Error(writer, fmt.Sprintf("The template %s does not exist.", name),
			http.StatusInternalServerError)
	}

	buf := r.bufferPool.Get()
	defer r.bufferPool.Put(buf)

	err := tmpl.Execute(buf, data)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = buf.WriteTo(writer)
	return err
}

func (r renderer) RenderIndex(writer http.ResponseWriter) error {
	return r.Render(writer, "index.gohtml", nil)
}