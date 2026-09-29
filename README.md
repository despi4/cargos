# Cargos

Cargos - это backend логистической платформы

### Technical Stack\

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

### Structure

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