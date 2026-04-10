package handler

import (
	"html/template"
	"net/http"

	"github.com/F3dosik/metalert/internal/service"
	"github.com/F3dosik/metalert/internal/templates"
)

func MainHandler(svc service.MetricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mainPage(w, r, svc)
	}
}

func mainPage(w http.ResponseWriter, r *http.Request, svc service.MetricsService) {
	tmpl, err := template.ParseFS(templates.TemplatesFS, "index.html")
	if err != nil {
		http.Error(w, errTemplateParsing.Error()+err.Error(), http.StatusInternalServerError)
		return
	}

	metrics, err := svc.GetAll(r.Context())
	if err != nil {
		http.Error(w, "failed to load metrics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tmpl.Execute(w, metrics); err != nil {
		http.Error(w, errTemplateRenderind.Error()+err.Error(), http.StatusInternalServerError)
		return
	}
}
