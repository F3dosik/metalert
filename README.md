# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**


## Оптимизация памяти

### Профилирование

Профили сняты под нагрузкой 50000 запросов (hey, concurrency=50).

### Результат (pprof -top -diff_base=profiles/base.pprof profiles/result.pprof)

    -514kB 11.76% 925.32%     -514kB 11.76%  bufio.NewReaderSize (inline)
  513.50kB 11.75% 937.07%   513.50kB 11.75%  regexp/syntax.(*compiler).inst (inline)
     513kB 11.74% 948.81%      513kB 11.74%  bufio.NewWriterSize (inline)
 -512.56kB 11.73% 937.08%  -512.56kB 11.73%  compress/flate.newHuffmanEncoder (inline)
 -512.25kB 11.72% 925.35%  -512.25kB 11.72%  runtime.mallocgc
         0     0% 925.35%   516.01kB 11.81%  bufio.(*Writer).Flush
         0     0% 925.35%     -514kB 11.76%  bufio.NewReader (inline)
         0     0% 925.35%  1026.25kB 23.48%  compress/flate.(*Writer).Close (inline)
         0     0% 925.35%  1026.25kB 23.48%  compress/flate.(*compressor).close
         0     0% 925.35%  1026.25kB 23.48%  compress/flate.(*compressor).deflate
         0     0% 925.35%  5499.65kB 125.85%  compress/flate.(*compressor).init
         0     0% 925.35%  1026.25kB 23.48%  compress/flate.(*compressor).writeBlock
         0     0% 925.35%  1026.25kB 23.48%  compress/flate.(*huffmanBitWriter).indexTokens
         0     0% 925.35%  1026.25kB 23.48%  compress/flate.(*huffmanBitWriter).writeBlock
         0     0% 925.35%  -512.56kB 11.73%  compress/flate.newHuffmanBitWriter (inline)
         0     0% 925.35%  1026.25kB 23.48%  compress/gzip.(*Writer).Close
         0     0% 925.35% 38895.34kB 890.06%  compress/gzip.(*Writer).Write
         0     0% 925.35% 38895.34kB 890.06%  fmt.Fprint
         0     0% 925.35% 38895.34kB 890.06%  github.com/F3dosik/metalert.git/internal/handler.RespondText
         0     0% 925.35% 38895.34kB 890.06%  github.com/F3dosik/metalert.git/internal/handler.RespondTextOK (inline)
         0     0% 925.35% 38895.34kB 890.06%  github.com/F3dosik/metalert.git/internal/handler.updateJSON
         0     0% 925.35% 38895.34kB 890.06%  github.com/F3dosik/metalert.git/internal/middleware.(*loggingResponseWriter).Write
         0     0% 925.35%  1026.25kB 23.48%  github.com/F3dosik/metalert.git/internal/middleware/gzip.(*compressWriter).Close
         0     0% 925.35% 38895.34kB 890.06%  github.com/F3dosik/metalert.git/internal/middleware/gzip.(*compressWriter).Write
         0     0% 925.35% 39921.59kB 913.54%  github.com/F3dosik/metalert.git/internal/server.(*Server).routes.WithCompression.func3.1
         0     0% 925.35%  1026.25kB 23.48%  github.com/F3dosik/metalert.git/internal/server.(*Server).routes.WithCompression.func3.1.1
         0     0% 925.35% 38895.34kB 890.06%  github.com/F3dosik/metalert.git/internal/server.(*Server).routes.WithLogging.func4.1
         0     0% 925.35% 38895.34kB 890.06%  github.com/F3dosik/metalert.git/internal/server.(*Server).routes.func1.RequireJSON.1.1
         0     0% 925.35% 38895.34kB 890.06%  github.com/F3dosik/metalert.git/internal/server.(*Server).routes.func1.UpdateJSONHandler.2
         0     0% 925.35% 38895.34kB 890.06%  github.com/go-chi/chi/v5.(*ChainHandler).ServeHTTP
         0     0% 925.35% 38895.34kB 890.06%  github.com/go-chi/chi/v5.(*Mux).Mount.func1
         0     0% 925.35% 39921.59kB 913.54%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 925.35% 38895.34kB 890.06%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 925.35%   513.50kB 11.75%  github.com/go-resty/resty/v2.init
         0     0% 925.35%   516.01kB 11.81%  io.Copy (inline)
         0     0% 925.35%   516.01kB 11.81%  io.CopyN
         0     0% 925.35%   516.01kB 11.81%  io.copyBuffer
         0     0% 925.35%   516.01kB 11.81%  io.discard.ReadFrom
         0     0% 925.35%   516.01kB 11.81%  net/http.(*chunkWriter).Write
         0     0% 925.35%   516.01kB 11.81%  net/http.(*chunkWriter).writeHeader
         0     0% 925.35%      513kB 11.74%  net/http.(*conn).readRequest
         0     0% 925.35% 40436.60kB 925.33%  net/http.(*conn).serve
         0     0% 925.35%   516.01kB 11.81%  net/http.(*response).finishRequest
         0     0% 925.35% 39921.59kB 913.54%  net/http.HandlerFunc.ServeHTTP
         0     0% 925.35%     -514kB 11.76%  net/http.newBufioReader
         0     0% 925.35%      513kB 11.74%  net/http.newBufioWriterSize
         0     0% 925.35% 39921.59kB 913.54%  net/http.serverHandler.ServeHTTP
         0     0% 925.35%   513.50kB 11.75%  regexp.Compile (inline)
         0     0% 925.35%   513.50kB 11.75%  regexp.MustCompile
         0     0% 925.35%   513.50kB 11.75%  regexp.compile
         0     0% 925.35%   513.50kB 11.75%  regexp/syntax.Compile
         0     0% 925.35%   513.50kB 11.75%  runtime.doInit (inline)
         0     0% 925.35%   513.50kB 11.75%  runtime.doInit1
         0     0% 925.35%  -512.25kB 11.72%  runtime.gcBgMarkWorker
         0     0% 925.35%   513.50kB 11.75%  runtime.main
         0     0% 925.35%  -512.25kB 11.72%  runtime.newobject
         0     0% 925.35%   516.01kB 11.81%  sync.(*Pool).Get
fedos@BigFriend:~/Programming/Yandex/metalert/profiles$ 
         0     0% 925.35%   516.01kB 11.81%  net/http.(*chunkWriter).writeHeader
         0     0% 925.35%      513kB 11.74%  net/http.(*conn).readRequest
         0     0% 925.35% 40436.60kB 925.33%  net/http.(*conn).serve
         0     0% 925.35%   516.01kB 11.81%  net/http.(*response).finishRequest
         0     0% 925.35% 39921.59kB 913.54%  net/http.HandlerFunc.ServeHTTP
         0     0% 925.35%     -514kB 11.76%  net/http.newBufioReader
         0     0% 925.35%      513kB 11.74%  net/http.newBufioWriterSize
         0     0% 925.35% 39921.59kB 913.54%  net/http.serverHandler.ServeHTTP
         0     0% 925.35%   513.50kB 11.75%  regexp.Compile (inline)
         0     0% 925.35%   513.50kB 11.75%  regexp.MustCompile
         0     0% 925.35%   513.50kB 11.75%  regexp.compile
         0     0% 925.35%   513.50kB 11.75%  regexp/syntax.Compile
         0     0% 925.35%   513.50kB 11.75%  runtime.doInit (inline)
         0     0% 925.35%   513.50kB 11.75%  runtime.doInit1
         0     0% 925.35%  -512.25kB 11.72%  runtime.gcBgMarkWorker
         0     0% 925.35%   513.50kB 11.75%  runtime.main
         0     0% 925.35%  -512.25kB 11.72%  runtime.newobject
         0     0% 925.35%   516.01kB 11.81%  sync.(*Pool).Get

### Что изменено

- `compress/gzip.Writer` — переиспользование через `sync.Pool` вместо создания на каждый запрос
- `compress/gzip.Reader` — аналогично через `sync.Pool`

### Эффект

| Метрика | До | После |
|---------|-----|-------|
| bufio.NewReaderSize | +514kB | -514kB |
| compress/flate.newHuffmanEncoder | +512kB | -512kB |
| runtime.mallocgc | baseline | -512kB |

Отрицательные значения подтверждают снижение аллокаций.