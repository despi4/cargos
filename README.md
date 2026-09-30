# Cargos

Cargos - это backend логистической платформы

### Technical Stack

```
Go
Gin
PostgreSQL
sqlc
Redis
gRPC
RabbitMQ
Docker
microservices
observability
retry / idempotency
```

### Project Pipeline

```
Client (Web/App)
  ↓
API Gateway (Gin)
  ↓
┌──────────────────────┐
│ Auth Service         │ Управление правами закупщиков
├──────────────────────┤
│ Inventory Service    │ Учет остатков (сколько на складе)
├──────────────────────┤
│ Procurement Service  │ Справочник поставщиков и сроков доставки
├──────────────────────┤
│ Demand Service       │ Сбор статистики продаж/расхода
├──────────────────────┤
│ Prediction Service   │ "Мозг": расчет рекомендаций для закупки
├──────────────────────┤
│ Notification Service │ Уведомления ("Пора заказать зимние шины")
└──────────────────────┘
        ↓
     RabbitMQ
```

### Structure of Services

**Hexganol Architecture**
```
    internal/
    ├── core/                       # Чистая бизнес-логика              
    │   ├── domain/                 # Бизнес модели                                 
    │   ├── ports/                  # Interfaces                                         
    │   └── services/               # Реализация бизнес-логики                                 
    │                                                                                          
    └── adapters/    
        ├── primary/                # Вход в сервис                               
        │   ├── http/               # Gin контроллеры, роуты, DTO запросов/ответов
        │   └── grpc/               # gRPC хэндлеры
        │
        └── secondary/              # ТО, К ЧЕМУ ОБРАЩАЕМСЯ МЫ (инфраструктура)
            ├── postgres/           # Репозиторий БД (sqlc, pgx)
            └── redis/              # Репозиторий сессий (go-redis)
```
