### Default flags
- d "host=localhost user=postgres password=postgres dbname=examples sslmode=disable"
- f "storage.json"
- DATABASE_DSN=host=localhost user=postgres password=postgres dbname=examples sslmode=disable
- FILE_STORAGE_PATH=storage.json;STORE_INTERVAL=5

## todo
- писать тесты на `updateMetricHandler`
- писать тесты на `getJsonMetricHandler`
- тесты для `pingDbHandler` с моками БД
- написать тесты на сохранение в файл
- переделать структуру проекта под картинку
- починить получение всех метрик в виде html
- доделать инкремент 13

## Замечания по спринту 3
- В лекции про контексты опрерируется каналами, но каналов еще небыло.
