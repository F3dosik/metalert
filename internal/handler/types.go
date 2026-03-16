package handler

import "errors"

var (
	errInvalidContentType = errors.New("ошибка: Content-Type должен быть text/plain")
	errTemplateParsing    = errors.New("ошибка парсинга шаблона")
	errTemplateRenderind  = errors.New("ошибка рендеринга шаблона HTML")
)

const (
	ContentTypePlainText = "text/plain; charset=utf-8"
	ContentType          = "text/plain"
)
