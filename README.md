# Notification Service (Golang) - Структура проекта

```
email-service-go/
├── cmd/
│   └── server/
│       └── main.go                      # Entry point
│
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── email.go                 # Email entity
│   │   │   └── attachment.go            # Attachment entity
│   │   ├── repository/
│   │   │   └── email_repository.go      # Repository interface
│   │   └── service/
│   │       ├── email_provider.go        # Email provider interface
│   │       └── queue_service.go         # Queue service interface
│   │
│   ├── application/
│   │   ├── usecase/
│   │   │   ├── send_email.go           # SendEmail use case
│   │   │   └── get_email_status.go     # GetEmailStatus use case
│   │   └── dto/
│   │       ├── send_email_dto.go       # Request DTO
│   │       └── email_response_dto.go    # Response DTO
│   │
│   ├── infrastructure/
│   │   ├── persistence/
│   │   │   ├── postgres/
│   │   │   │   ├── connection.go       # DB connection pool
│   │   │   │   ├── email_repository.go # SQL implementation
│   │   │   │   └── migrations.go       # SQL migrations
│   │   │   └── redis/
│   │   │       └── cache.go            # Redis cache (optional)
│   │   ├── email/
│   │   │   ├── smtp_provider.go        # SMTP implementation
│   │   │   └── mock_provider.go        # Mock for testing
│   │   ├── queue/
│   │   │   ├── rabbitmq.go             # RabbitMQ client
│   │   │   └── consumer.go             # Message consumer
│   │   └── storage/
│   │       ├── s3_storage.go           # S3 file storage
│   │       └── local_storage.go        # Local file storage
│   │
│   ├── presentation/
│   │   ├── http/
│   │   │   ├── handler/
│   │   │   │   ├── email_handler.go   # HTTP handlers
│   │   │   │   └── health_handler.go  # Health check
│   │   │   ├── middleware/
│   │   │   │   ├── error_handler.go   # Error handling
│   │   │   │   ├── logger.go          # Request logging
│   │   │   │   ├── cors.go            # CORS middleware
│   │   │   │   └── rate_limiter.go    # Rate limiting
│   │   │   └── router/
│   │   │       └── router.go          # Routes setup
│   │   └── queue/
│   │       └── claim_consumer.go      # Claim event consumer
│   │
│   └── pkg/
│       ├── config/
│       │   └── config.go              # Configuration loader
│       ├── logger/
│       │   └── logger.go              # Structured logging
│       ├── validator/
│       │   └── validator.go           # Input validation
│       └── errors/
│           └── errors.go              # Custom errors
│
│
│
├── configs/
│   ├── config.yaml
│   └── config.local.yaml
│
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## Технологический стек

### Core
- **Go**: 1.21+
- **HTTP Framework**: Chi (или Gin/Echo)
- **Database Driver**: pgx/v5
- **Message Queue**: amqp091-go
- **Email**: gomail v2

### Infrastructure
- **PostgreSQL**: 16
- **RabbitMQ**: 3.12
- **Redis**: 7 (optional, для кеширования)

### Tools
- **Migration**: golang-migrate
- **Config**: viper
- **Logging**: zap/zerolog
- **Validation**: go-playground/validator
- **Testing**: testify

### Monitoring (optional)
- **Metrics**: prometheus client
- **Tracing**: opentelemetry
- **Profiling**: pprof

## Команды для запуска

```bash
# Установка зависимостей
make deps

# Запуск сервиса
make run

# Build
make build

# Docker
make docker-build
make docker-run
```

## Преимущества Go реализации

1. **Performance**: В 5-10 раз быстрее Node.js
2. **Concurrency**: Встроенные goroutines для параллельной обработки
3. **Memory**: Меньше потребление памяти
4. **Deployment**: Один бинарный файл, не нужен runtime
5. **Type Safety**: Строгая типизация на этапе компиляции
6. **Error Handling**: Явная обработка ошибок
7. **Standard Library**: Богатая стандартная библиотека

## Ключевые отличия от Node.js версии

- ✅ Нативные SQL запросы вместо Prisma ORM
- ✅ Структурное логирование (zap)
- ✅ Контекст для таймаутов и отмены операций
- ✅ Goroutines для конкурентной обработки
- ✅ Channels для коммуникации между компонентами
- ✅ Более явная обработка ошибок
- ✅ Compile-time проверки типов
