# Cargos

Backend платформы управления складом и закупками на Go. API Gateway и шесть самостоятельных сервисов: авторизация, склад, закупки, история расхода, рекомендации и уведомления.

## Запуск

Нужны Docker и Docker Compose v2:

```bash
cp .env.example .env
docker compose up --build -d --wait
```

- Swagger UI: **http://localhost:8080/docs/**
- OpenAPI 3.0.3: **http://localhost:8080/openapi.json**
- Готовность всей системы: **http://localhost:8080/readyz**

Если порт занят, задайте `GATEWAY_PORT` в `.env`. Swagger UI включён в бинарник Gateway и не требует внешнего CDN. Для первого скачивания образов и Go-зависимостей нужен интернет.

Проверка полного сценария с демо-данными:

```bash
make smoke
# Для другого адреса:
BASE_URL=http://localhost:18080 python3 scripts/smoke.py
```

Скрипт создаёт отдельный аккаунт, товар и поставщика, вносит остатки и расход, рассчитывает рекомендацию, принимает закупку, проверяет уведомление и авторизацию. Email выводится в терминал; пароль демо-аккаунта — `cargos-demo-password`.

Остановка с сохранением PostgreSQL-данных: `docker compose down`.

## Структура

В `services/` один Go-модуль. Каждый сервис имеет собственную точку запуска, бизнес-логику, адаптеры, миграции и Docker-образ.

```text
services/
├── api_gateway/
│   ├── cmd/api/main.go
│   ├── internal/adapters/primary/http/
│   ├── docs/                       # OpenAPI и локальный Swagger UI
│   └── service.go                  # сборка зависимостей Gateway
├── auth_service/
├── inventory_service/
├── procurement_service/
├── demand_service/
├── prediction_service/
├── notification_service/
├── internal/platform/              # инфраструктура без бизнес-моделей
│   ├── client/                     # межсервисный HTTP-транспорт
│   ├── config/
│   ├── httpx/                      # HTTP-ответы и middleware
│   ├── postgres/                   # подключения и запуск миграций
│   └── server/                     # жизненный цикл процесса
├── pkg/
│   ├── entity/                     # ID и временные метки
│   └── fault/                      # ошибки приложения
├── tests/integration/              # проверка взаимодействия сервисов
├── Dockerfile                      # сборка выбранного cmd/api
└── go.mod
```

Одинаковая структура бизнес-сервисов:

```text
inventory_service/
├── cmd/api/main.go
├── internal/
│   ├── core/
│   │   ├── domain/                 # сущности и правила предметной области
│   │   ├── ports/                  # интерфейсы зависимостей
│   │   └── services/               # сценарии приложения
│   └── adapters/
│       ├── primary/http/           # маршруты и HTTP-обработчики
│       └── secondary/postgres/     # реализация репозиториев
├── migrations/
└── service.go                      # подключение адаптеров к бизнес-логике
```

Клиенты других сервисов находятся в `adapters/secondary/httpclient/`; реализация паролей и JWT — в `auth_service/internal/adapters/secondary/security/`. Бизнес-логика зависит от интерфейсов в `ports`, HTTP-обработчики вызывают сценарии из `core/services`. В `core` нет HTTP, pgx и SQL. Сервисы не импортируют чужие внутренние пакеты.

## Функциональность

| Сервис | API | Ответственность |
| --- | --- | --- |
| Auth | `/api/v1/auth/*` | Регистрация, вход, refresh, logout, текущий аккаунт |
| Inventory | `/api/v1/products`, `/api/v1/movements` | Каталог, архивирование, остатки и журнал движений |
| Procurement | `/api/v1/suppliers`, `/api/v1/orders` | Поставщики, закупки и приёмка |
| Demand | `/api/v1/demand` | История расхода для прогнозирования |
| Prediction | `/api/v1/predictions` | Расчёт рекомендаций и их история |
| Notification | `/api/v1/notifications` | Inbox и отметки прочтения |

В Swagger сначала выполните `auth/register` и `auth/login`, затем вставьте `access_token` в **Authorize**. Access действует 15 минут, refresh — 7 дней. Refresh меняет пару токенов, logout немедленно отзывает текущую сессию. Данные изолированы по аккаунтам.

Основные правила:

- Товар создаётся с нулевым остатком. Изменения выполняются движениями с уникальной `reference`; отрицательное количество списывает товар. Одинаковый повтор не изменяет остаток повторно. Отрицательный остаток запрещён.
- `PUT` заменяет метаданные товара или поставщика; SKU неизменяем. `DELETE` архивирует запись с сохранением истории.
- Demand хранит статистику и не списывает склад автоматически. Исторические продажи можно импортировать отдельно от текущих движений.
- Закупка проходит `draft → ordered → receiving → received`. Из `draft` и `ordered` доступна отмена. Приёмка сохраняет намерение до обращения к складу; незавершённые операции повторяются каждые 2 секунды. Уникальная ссылка закупки предотвращает двойное зачисление. Существующую закупку можно принять и после архивирования товара.
- Количества — целые, остаток и отдельная операция ограничены 1 000 000 000. Цена задаётся в минимальных денежных единицах; валюта — код из трёх букв.
- Прогноз использует средний расход за окно (по умолчанию 30 дней), минимальный и целевой остатки, срок поставки и уже заказанные товары. Поставщик задаётся явно либо выбирается самый быстрый активный; без поставщиков используется срок 7 дней.
- Уведомления доставляются в inbox через сохраняемую очередь с повторными попытками и исключением дубликатов. Email/SMS в этой версии не отправляются.

Списки поддерживают `limit` (1–500, по умолчанию 100) и `offset`. Подробные запросы, ответы и ошибки описаны в Swagger.

## Хранение и инфраструктура

Стек: Go 1.26.1, `net/http`, PostgreSQL 17, pgx; для репозитория пользователей сохранён sqlc. Межсервисные вызовы идут по HTTP. Redis, RabbitMQ и gRPC для этой версии не требуются.

Каждый сервис владеет своей схемой PostgreSQL и предметными таблицами: `users`, `sessions`, `products`, `movements`, `suppliers`, `orders`, `demand_records`, `predictions`, `notifications`. JSONB используется для вложенного списка рекомендаций. SQL и миграции принадлежат соответствующему сервису.

Миграции запускаются при старте, отмечаются в `schema_migrations` и защищены advisory lock. Если остались данные предыдущей реализации в `records`, миграции переносят их в предметные таблицы, сохраняя исходные записи. Таблицы в `public` автоматически не переносятся и не удаляются.

Наружу опубликован только Gateway. Внутренние запросы требуют ключ и идентификатор пользователя; клиентские внутренние заголовки Gateway удаляет. `/healthz` проверяет процесс, `/readyz` — зависимости, `/metrics` отдаёт счётчик запросов Prometheus. Поддерживаются логи с ID запроса и корректное завершение по SIGTERM.

## Разработка

```bash
make test       # unit-тесты; интеграционные пропускаются без БД
make vet
make fmt
make openapi    # обновить api_gateway/docs/openapi.json
```

Для интеграционных тестов используйте отдельную PostgreSQL-базу:

```bash
TEST_DATABASE_URL='postgres://cargos:cargos@localhost:5432/cargos_test?sslmode=disable' make integration
```

Тесты проверяют полный сценарий, конкурентные списания, идемпотентность, изоляцию аккаунтов, токены, восстановление приёмки и доставки, перенос старых данных. CI выполняет `go vet` и `go test -race ./...` с PostgreSQL.

Для локального запуска одного сервиса:

```bash
cd services
export DATABASE_URL='postgres://cargos:cargos@localhost:5432/cargos_auth?sslmode=disable'
export INTERNAL_KEY='local-internal-key-at-least-32-bytes-long'
export JWT_SECRET='local-jwt-secret-at-least-32-bytes-long'
HTTP_ADDR=:8081 go run ./auth_service/cmd/api
```

Остальные точки запуска: `./inventory_service/cmd/api` на 8082, `./procurement_service/cmd/api` на 8083, `./demand_service/cmd/api` на 8084, `./prediction_service/cmd/api` на 8085, `./notification_service/cmd/api` на 8086 и `./api_gateway/cmd/api` на 8080. Адреса зависимостей переопределяются переменными `AUTH_URL`, `INVENTORY_URL`, `PROCUREMENT_URL`, `DEMAND_URL`, `PREDICTION_URL`, `NOTIFICATION_URL`.

Перед внешним развёртыванием замените ключи и пароль PostgreSQL в `.env`, настройте TLS. В этой учебной версии нет ролей организаций, сброса пароля и подтверждения email; схемы используют одну учётную запись БД.
