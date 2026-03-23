package handler

import (
	"html/template"
	"net/http"

	"github.com/F3dosik/metalert/internal/repository"
	"github.com/F3dosik/metalert/internal/templates"
	"github.com/F3dosik/metalert/pkg/models"
)

func MainHandler(storage repository.MetricsStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mainPage(w, r, storage)
	}
}

func mainPage(w http.ResponseWriter, r *http.Request, storage repository.MetricsStorage) {
	tmpl, err := template.ParseFS(templates.TemplatesFS, "index.html")
	if err != nil {
		http.Error(w, errTemplateParsing.Error()+err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	var metrics []models.Metric
	metrics, err = storage.GetAllMetrics(ctx)
	if err != nil {
		http.Error(w, "failed to load metrics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, metrics)
	if err != nil {
		http.Error(w, errTemplateRenderind.Error()+err.Error(), http.StatusInternalServerError)
		return
	}
}
