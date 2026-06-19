# Go Calculator — микросервисный калькулятор

Учебный проект: HTTP-калькулятор на Go с двумя микросервисами, PostgreSQL и Kafka.

## Архитектура

```
                    POST /calculate
Клиент ──────────────────────────────► API Service (:8080)
                                           │
                                           ├─ calculator.Calculate()
                                           └─ Kafka Producer
                                                  │
                                                  ▼
                                           topic: calculations
                                                  │
                                                  ▼
                                        Storage Service (:8081)
                                           │
                                           ├─ Kafka Consumer
                                           └─ PostgreSQL (:5432)

                    GET /calculations
Клиент ──────────────────────────────► API Service ──HTTP──► Storage Service ──► PostgreSQL
```

| Компонент | Порт | Назначение |
|-----------|------|------------|
| API Service | 8080 | Публичный HTTP API |
| Storage Service | 8081 | Внутренний HTTP API + Kafka consumer |
| PostgreSQL | 5432 | Хранение вычислений |
| Kafka | 9092 | Асинхронная передача событий |
| Kafka UI | 8082 | Веб-интерфейс для просмотра топиков |

### Потоки данных

- **POST `/calculate`** — API считает результат и отправляет событие в Kafka. Storage consumer сохраняет запись в PostgreSQL.
- **GET `/calculations`** — API запрашивает историю у Storage по HTTP.

---

## Структура проекта

```
go-calculator/
├── cmd/
│   ├── api/main.go              # Точка входа API-сервиса
│   └── storage/main.go          # Точка входа Storage-сервиса
├── internal/
│   ├── api/
│   │   ├── handler/             # HTTP-обработчики (/calculate, /calculations)
│   │   ├── client/              # HTTP/gRPC клиент к Storage (gRPC — заготовка)
│   │   └── kafka/               # Kafka Producer
│   ├── storage/
│   │   ├── handler/             # HTTP-обработчики Storage
│   │   ├── repository/          # PostgreSQL
│   │   ├── kafka/               # Kafka Consumer
│   │   └── transport/grpc/      # gRPC server (заготовка)
│   ├── calculator/              # Бизнес-логика: математика
│   └── model/                   # Общая модель Calculation
├── proto/storage/v1/            # Protobuf-контракт для будущего gRPC
├── migrations/                  # SQL-миграции
├── docker-compose.yml           # PostgreSQL + Kafka + Kafka UI
├── go.mod
└── README.md
```

---

## API

### POST `/calculate`

```json
{ "a": 10, "b": 5, "operator": "+" }
```

Ответ:

```json
{ "a": 10, "b": 5, "operator": "+", "result": 15 }
```

### GET `/calculations`

Возвращает историю вычислений из PostgreSQL.

---

## Быстрый старт

### 1. Инфраструктура (PostgreSQL + Kafka)

```powershell
docker compose up -d
```

### 2. Storage-сервис

```powershell
go run ./cmd/storage
```

### 3. API-сервис

```powershell
go run ./cmd/api
```

### 4. Проверка

```powershell
curl -X POST http://localhost:8080/calculate -H "Content-Type: application/json" -d "{\"a\": 10, \"b\": 5, \"operator\": \"+\"}"
curl http://localhost:8080/calculations
```

Kafka UI: http://localhost:8082

---

## Переменные окружения

| Переменная | Сервис | По умолчанию |
|------------|--------|--------------|
| `API_PORT` | api | `8080` |
| `STORAGE_PORT` | storage | `8081` |
| `STORAGE_URL` | api | `http://localhost:8081` |
| `KAFKA_BROKER` | api | `localhost:9092` |
| `DATABASE_URL` | storage | `postgres://calculator:calculator@localhost:5432/calculator?sslmode=disable` |

---

## Стек

| Технология | Статус |
|------------|--------|
| Go | ✅ |
| HTTP (net/http) | ✅ |
| PostgreSQL (pgx) | ✅ |
| Kafka (segmentio/kafka-go) | ✅ |
| Docker Compose | ✅ |
| gRPC | 🔜 заготовка (proto + stubs) |

---

## Сборка

```powershell
go build ./...
```
