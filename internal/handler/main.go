package handler

import (
	"html/template"
	"net/http"

	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/F3dosik/metalert.git/internal/templates"
)

func MainHandler(storage *repository.MemMetricsStorage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		mainPage(rw, storage)
	}
}

func mainPage(rw http.ResponseWriter, storage *repository.MemMetricsStorage) {
	tmpl, err := template.ParseFS(templates.TemplatesFS, "index.html")
	if err != nil {
		http.Error(rw, errTemplateParsing.Error()+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(rw, storage)
	if err != nil {
		http.Error(rw, errTemplateRenderind.Error()+err.Error(), http.StatusInternalServerError)
		return
	}
}
