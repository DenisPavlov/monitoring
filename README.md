- To run server run `go run cmd/server/main.go -d "host=localhost user=postgres password=postgres dbname=examples sslmode=disable"`
- To run agent run `go run cmd/agent/main.go`

### Default flags
- d "host=localhost user=postgres password=postgres dbname=examples sslmode=disable"
- f "storage.json"
- DATABASE_DSN=host=localhost user=postgres password=postgres dbname=examples sslmode=disable
- FILE_STORAGE_PATH=storage.json;STORE_INTERVAL=5

## todo
- добавить в log какая БД поднимается