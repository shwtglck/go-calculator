# Калькулятор-сервис (учебный проект)

Два микросервиса на Go:

1. **API** — принимает HTTP-запросы, считает результат, обращается к storage-сервису
2. **Storage** — сохраняет и читает данные из PostgreSQL

Сейчас сервисы общаются по **HTTP**. Структура проекта подготовлена для дальнейшего внедрения **gRPC** и **Kafka**.

---

## Структура проекта

```
newstart/
├── cmd/
│   ├── api/
│   │   └── main.go                         # Запуск HTTP API
│   └── storage/
│       └── main.go                         # Запуск сервиса хранения
├── internal/
│   ├── api/
│   │   ├── handler/                        # Публичные HTTP-эндпоинты
│   │   ├── client/                         # Клиент к storage (HTTP / gRPC)
│   │   └── kafka/                          # Заготовка под Kafka producer
│   ├── storage/
│   │   ├── handler/                        # Внутренние HTTP-эндпоинты
│   │   ├── repository/                     # PostgreSQL
│   │   ├── transport/grpc/                 # Заготовка под gRPC server
│   │   └── kafka/                          # Заготовка под Kafka consumer
│   ├── calculator/                         # Математика
│   └── model/                              # Общая модель Calculation
├── proto/
│   └── storage/v1/storage.proto            # Контракт для будущего gRPC
├── migrations/
│   └── 001_init.sql
├── docker-compose.yml
├── go.mod
└── README.md
```

### Как это работает

```
Клиент → API (8080) → Storage (8081) → PostgreSQL
```

| Сервис | Порт | Ответственность |
|--------|------|-----------------|
| `cmd/api` | 8080 | POST `/calculate`, GET `/calculations` |
| `cmd/storage` | 8081 | POST/GET `/calculations`, работа с БД |

---

## API (публичный, порт 8080)

### POST `/calculate` — посчитать и сохранить

**Запрос:**
```json
{
  "a": 10,
  "b": 5,
  "operator": "+"
}
```

**Ответ:**
```json
{
  "id": 1,
  "a": 10,
  "b": 5,
  "operator": "+",
  "result": 15,
  "created_at": "2026-06-08T12:00:00Z"
}
```

### GET `/calculations` — получить историю

---

## Storage API (внутренний, порт 8081)

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/calculations` | Сохранить вычисление |
| GET | `/calculations` | Получить историю |
| GET | `/health` | Проверка состояния |

---

## Как запустить

### Шаг 1. База данных

```powershell
docker compose up -d
```

### Шаг 2. Storage-сервис (в первом терминале)

```powershell
go run ./cmd/storage
```

### Шаг 3. API-сервис (во втором терминале)

```powershell
go run ./cmd/api
```

### Шаг 4. Проверка

```powershell
curl -X POST http://localhost:8080/calculate -H "Content-Type: application/json" -d "{\"a\": 10, \"b\": 5, \"operator\": \"+\"}"
curl http://localhost:8080/calculations
```

---

## Переменные окружения

| Переменная | Сервис | По умолчанию | Описание |
|------------|--------|--------------|----------|
| `API_PORT` | api | `8080` | Порт API-сервиса |
| `STORAGE_PORT` | storage | `8081` | Порт storage-сервиса |
| `STORAGE_URL` | api | `http://localhost:8081` | Адрес storage-сервиса |
| `STORAGE_TRANSPORT` | api | `http` | Транспорт: `http` или `grpc` (пока только http) |
| `DATABASE_URL` | storage | `postgres://calculator:calculator@localhost:5432/calculator?sslmode=disable` | PostgreSQL |

---

## Что подготовлено для будущего

| Технология | Где лежит | Статус |
|------------|-----------|--------|
| gRPC контракт | `proto/storage/v1/storage.proto` | Описан, не подключён |
| gRPC client | `internal/api/client/grpc.go` | Заготовка |
| gRPC server | `internal/storage/transport/grpc/` | Заготовка |
| Kafka producer | `internal/api/kafka/` | Заготовка |
| Kafka consumer | `internal/storage/kafka/` | Заготовка |

---

## Что можно сказать ментору

«Проект разделён на два сервиса: `cmd/api` для HTTP API и `cmd/storage` для работы с PostgreSQL. API не ходит в базу напрямую — только через storage-сервис. Есть интерфейс клиента, proto-файл и заготовки под gRPC и Kafka.»
