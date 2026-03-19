// Package handler содержит HTTP-хендлеры сервера метрик.
//
// Хендлеры принимают запросы от агентов и клиентов, валидируют входные данные,
// обновляют хранилище метрик и возвращают ответы в текстовом или JSON-формате.
//
// Основные эндпоинты:
//
//   - POST /update/{metType}/{metName}/{metValue} — обновление метрики через URL-параметры
//   - POST /update/ — обновление одной метрики через JSON
//   - POST /updates/ — пакетное обновление метрик через JSON
//   - GET  /value/{metType}/{metName} — получение значения метрики через URL
//   - POST /value/ — получение значения метрики через JSON
//   - GET  /ping — проверка соединения с базой данных
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RespondJSON записывает в ответ JSON-тело с заданным HTTP-статусом.
// Устанавливает заголовок Content-Type: application/json.
func RespondJSON(rw http.ResponseWriter, status int, data any) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(status)
	json.NewEncoder(rw).Encode(data)
}

// RespondJSONOK записывает успешный JSON-ответ со статусом 200 OK.
func RespondJSONOK(rw http.ResponseWriter, data any) {
	RespondJSON(rw, http.StatusOK, data)
}

// RespondText записывает в ответ текстовое сообщение с заданным HTTP-статусом.
// Устанавливает заголовок Content-Type: text/plain.
func RespondText(rw http.ResponseWriter, status int, message string) {
	rw.Header().Set("Content-Type", ContentTypePlainText)
	rw.WriteHeader(status)
	fmt.Fprint(rw, message)
}

// RespondTextOK записывает успешный текстовый ответ со статусом 200 OK.
func RespondTextOK(rw http.ResponseWriter, message string) {
	RespondText(rw, http.StatusOK, message)
}
