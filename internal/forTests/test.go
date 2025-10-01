package main

import (
	"fmt"
	"strings"
)

func main() {
	s := "http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>"
	s = strings.TrimPrefix(s, "/update/")
	strings.Split(s, "/")
	fmt.Println(s)
}
