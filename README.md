- To run server run `go run cmd/server/main.go -d "host=localhost user=postgres password=postgres dbname=examples sslmode=disable"`
- To run agent run `go run cmd/agent/main.go`

### Default flags
- d "host=localhost user=postgres password=postgres dbname=examples sslmode=disable"
- f "storage.json"
- DATABASE_DSN=host=localhost user=postgres password=postgres dbname=examples sslmode=disable
- FILE_STORAGE_PATH=storage.json;STORE_INTERVAL=5

## профилирование
- собрать профиль по памяти - `curl http://127.0.0.1:8082/debug/pprof/heap?seconds=300 > profiles/base1.prof`
- анализ профиля в браузере `go tool pprof -http=":9090" profiles/base.prof`
- смотреть разницу в профилях - `pprof -top -diff_base=profiles/base.pprof profiles/result.pprof`

Было выполнено профилирование сервера при стандартной работе агента. Профиль собирался в течении 5-ти минут работы сервера.
Было произведено 3 замера. При анализе профиля не было выявлено каких-либо утечек и мест для оптимизации.
Файлы с профилем расположены в `/profiles`.