## flexfind - AI Job Aggregator & CV Tailoring Platform

### 1. Назначение системы

Платформа предназначена для:

- агрегации вакансий из нескольких источников;
    
- нормализации и дедупликации вакансий;
    
- поиска публично доступных контактов HR / Recruiter / Hiring Manager / Team Lead;
    
- хранения профиля и базового CV кандидата;
    
- оценки соответствия кандидата вакансии;
    
- AI-анализа требований вакансии;
    
- адаптации CV под конкретную вакансию без добавления ложного опыта или навыков;
    
- отслеживания вакансий и статусов откликов.
    

Основная backend-платформа разрабатывается на **Go** с использованием микросервисной архитектуры.

---

# 2. Бизнес-логика и User Flow

## 2.1. Получение вакансий

1. `Source/Scraper Service` получает вакансии:
    
    - через официальные API job-board платформ;
        
    - RSS/API партнёров;
        
    - разрешённый scraping публичных страниц.
        
2. Полученные данные преобразуются во внутренний формат `Vacancy`.
    
3. Выполняются:
    
    - валидация;
        
    - очистка HTML;
        
    - нормализация компании и должности;
        
    - дедупликация.
        
4. Новая вакансия сохраняется.
    
5. В message broker публикуется событие:
    

```text
vacancy.created
```

---

## 2.2. AI-анализ вакансии

После `vacancy.created` AI Service получает вакансию и извлекает:

- hard skills;
    
- soft skills;
    
- обязательные требования;
    
- optional requirements;
    
- seniority;
    
- технологии;
    
- язык;
    
- требуемый опыт;
    
- обязанности;
    
- ключевые слова.
    

Результат сохраняется как структурированные данные.

Пример:

```json
{
  "position": "Go Backend Developer",
  "seniority": "junior",
  "required_skills": [
    "Go",
    "PostgreSQL",
    "REST"
  ],
  "preferred_skills": [
    "Kafka",
    "Kubernetes"
  ]
}
```

После анализа:

```text
vacancy.analyzed
```

---

## 2.3. Поиск контактов

`Contact Service` пытается определить:

- Recruiter;
    
- HR;
    
- Hiring Manager;
    
- Team Lead.
    

Источники:

- публичные данные job-board;
    
- официальный сайт компании;
    
- публичные профессиональные профили;
    
- разрешённые сторонние enrichment API.
    

Каждый контакт должен содержать:

```text
name
position
company
public_contact
source
confidence_score
```

Контакт нельзя считать подтверждённым без источника.

---

## 2.4. Matching

`Matching Service` получает:

- профиль кандидата;
    
- структурированный CV;
    
- структурированные требования вакансии.
    

Результат:

```text
match_score
matched_skills
missing_skills
preferred_skills
experience_match
language_match
explanation
```

Например:

```text
Match: 78%

Matched:
- Go
- PostgreSQL
- REST
- Docker

Missing:
- Kafka
- Kubernetes
```

---

## 2.5. Адаптация CV

Пользователь выбирает:

```text
Vacancy → Adapt CV
```

Создаётся асинхронная задача.

AI получает:

```text
Original CV
+
Candidate Profile
+
Vacancy Requirements
```

AI имеет право:

- изменить формулировки;
    
- изменить порядок навыков;
    
- выделить наиболее релевантный опыт;
    
- изменить summary;
    
- адаптировать bullet points;
    
- добавить релевантные ключевые слова, если они подтверждаются исходным CV.
    

AI запрещено:

- придумывать компании;
    
- придумывать коммерческий опыт;
    
- изменять даты работы;
    
- добавлять отсутствующие навыки как существующие;
    
- придумывать достижения.
    

Результат:

```text
Tailored CV
```

Пользователь должен увидеть изменения относительно оригинала до экспорта.

---

# 3. Архитектура

```text
                    Client
                      │
                  HTTPS/REST
                      │
                API Gateway
                      │
       ┌──────────────┼───────────────┐
       │              │               │
      gRPC           gRPC            gRPC
       │              │               │
       ▼              ▼               ▼
    Profile        Vacancy           CV
    Service        Service         Service
                      │
                      │
                Event Broker
              Kafka / RabbitMQ
                      │
       ┌──────────────┼───────────────┐
       ▼              ▼               ▼
    Scraper          AI           Matching
    Service        Service         Service
                                      │
                      ┌───────────────┤
                      ▼               ▼
                   Contact       Notification
                   Service          Service
```

---

# 4. Разделение на микросервисы

## 4.1. API Gateway

Ответственность:

- единая публичная API-точка;
    
- authentication middleware;
    
- authorization;
    
- rate limiting;
    
- routing;
    
- correlation/request ID;
    
- aggregation простых клиентских запросов.
    

Технологии:

```text
Go
Gin
REST/JSON
```

API Gateway не должен содержать бизнес-логику.

---

## 4.2. Auth Service

Функции:

- регистрация;
    
- login;
    
- refresh tokens;
    
- logout;
    
- password reset;
    
- управление сессиями.
    

Хранилище:

```text
PostgreSQL
Redis
```

---

## 4.3. Profile Service

Хранит:

- профиль кандидата;
    
- skills;
    
- preferred positions;
    
- location;
    
- experience;
    
- education;
    
- job preferences.
    

Интерфейс внутренних сервисов:

```text
gRPC
```

БД:

```text
PostgreSQL + sqlc
```

---

## 4.4. Vacancy Service

Отвечает за:

- хранение вакансий;
    
- компании;
    
- статус вакансии;
    
- source metadata;
    
- поиск;
    
- фильтрацию;
    
- дедупликацию.
    

БД:

```text
PostgreSQL + sqlc
```

Дополнительно:

```text
OpenSearch
```

может быть добавлен после MVP для полнотекстового поиска.

---

## 4.5. Scraper / Source Service

Отдельный сервис интеграции с источниками вакансий.

Каждый источник реализуется через adapter:

```go
type VacancySource interface {
    Fetch(ctx context.Context) ([]Vacancy, error)
}
```

Пример:

```text
HHAdapter
LinkedInAdapter
CompanyCareerAdapter
OtherJobBoardAdapter
```

Приоритет:

1. официальный API;
    
2. разрешённые feeds/integrations;
    
3. scraping только там, где это допустимо условиями источника и применимым законодательством.
    

Scraper не должен напрямую вызывать AI.

Результат публикуется через broker.

---

## 4.6. Contact Service

Отвечает за:

- поиск hiring contacts;
    
- нормализацию контактов;
    
- объединение дубликатов;
    
- сохранение источника;
    
- confidence score.
    

Пример:

```text
Company
   ↓
Domain discovery
   ↓
Public sources / enrichment API
   ↓
Potential HR / Team Lead
   ↓
Contact validation
```

---

## 4.7. CV Service

Отвечает за:

- загрузку CV;
    
- хранение версий;
    
- извлечение текста;
    
- структурированное представление CV;
    
- создание tailored versions;
    
- экспорт.
    

Поддерживаемые форматы:

```text
PDF
DOCX
```

Для каждого адаптированного CV необходимо сохранять:

```text
original_cv_id
vacancy_id
generated_version
created_at
AI_model
generation_metadata
```

---

## 4.8. AI Service

Централизованная точка взаимодействия с LLM.

Функции:

- анализ вакансий;
    
- extraction skills;
    
- CV parsing;
    
- CV tailoring;
    
- объяснение matching;
    
- structured output validation.
    

Внутренний API:

```text
gRPC
```

Тяжёлые задачи запускаются через message broker.

Пример:

```text
cv.tailoring.requested
        ↓
AI Service
        ↓
cv.tailoring.completed
```

AI Service должен быть независим от конкретного LLM-провайдера.

Интерфейс:

```go
type LLMProvider interface {
    Generate(ctx context.Context, req Request) (Response, error)
}
```

---

## 4.9. Matching Service

Отвечает за:

```text
Candidate ↔ Vacancy
```

Первый MVP не должен полагаться только на LLM.

Score формируется из детерминированных факторов:

```text
skills
experience
seniority
location
language
job preferences
```

LLM используется преимущественно для:

- semantic extraction;
    
- explanation;
    
- normalization.
    

---

## 4.10. Notification Service

События:

```text
new_matching_vacancy
cv_ready
contact_found
application_update
```

Каналы в будущем:

```text
Email
Telegram
Push
```

---

# 5. Межсервисное взаимодействие

## REST

Использовать для:

```text
Client → API Gateway
```

---

## gRPC

Использовать для синхронных внутренних запросов.

Например:

```text
Matching Service
     │
     ├── gRPC → Profile Service
     │
     └── gRPC → Vacancy Service
```

---

## Kafka / RabbitMQ

Использовать для асинхронной обработки.

Основные события:

```text
vacancy.discovered
vacancy.created
vacancy.updated
vacancy.analyzed

cv.uploaded
cv.parsed

cv.tailoring.requested
cv.tailoring.completed
cv.tailoring.failed

contact.search.requested
contact.found

matching.requested
matching.completed
```

### Broker

Для начальной версии допустимо использовать:

```text
RabbitMQ
```

Если основная цель проекта — изучение event-driven architecture.

При росте количества событий, replay, event streaming и большого числа consumers предпочтительно перейти на:

```text
Kafka
```

---

# 6. Functional Requirements — Core Features

## Пользователь

Система должна позволять:

- зарегистрироваться;
    
- авторизоваться;
    
- создать профиль;
    
- указать skills;
    
- загрузить CV;
    
- хранить несколько CV.
    

---

## Вакансии

Пользователь должен иметь возможность:

- просматривать вакансии;
    
- искать по названию;
    
- фильтровать по технологии;
    
- фильтровать по компании;
    
- фильтровать по seniority;
    
- фильтровать по location;
    
- сохранять вакансию;
    
- скрывать вакансию;
    
- видеть источник вакансии.
    

---

## AI-анализ

Для вакансии отображать:

- extracted skills;
    
- mandatory requirements;
    
- optional requirements;
    
- seniority;
    
- краткое AI-summary;
    
- match score;
    
- отсутствующие skills.
    

---

## Contacts

Для компании система должна иметь возможность отображать:

```text
Name
Role
Public contact/profile
Source
Confidence
```

---

## CV

Система должна позволять:

- загрузить базовое CV;
    
- распарсить CV;
    
- просмотреть структуру;
    
- выбрать вакансию;
    
- запустить адаптацию;
    
- сравнить original / tailored;
    
- вручную изменить результат;
    
- сохранить новую версию;
    
- экспортировать CV.
    

---

## Applications

Минимальные статусы:

```text
SAVED
PLANNED
APPLIED
INTERVIEW
OFFER
REJECTED
ARCHIVED
```

---

# 7. Внешние интеграции

## Job Sources

Приоритетно:

- официальные API;
    
- partner APIs;
    
- публично разрешённые feeds.
    

Для источников без API должен существовать отдельный adapter.

Не следует проектировать систему в расчёте на обход технических ограничений конкретной платформы.

---

## LLM

AI Service должен поддерживать несколько providers:

```text
OpenAI-compatible API
Anthropic
Gemini
Local LLM
```

Через общий abstraction layer.

Необходимо учитывать:

- token limits;
    
- timeout;
    
- retries;
    
- provider errors;
    
- rate limits;
    
- стоимость запроса.
    

---

## OCR

OCR необходим только для:

- scanned PDF;
    
- изображений CV;
    
- PDF без text layer.
    

Возможные варианты:

```text
Tesseract
Cloud OCR API
```

Перед OCR необходимо сначала попытаться извлечь embedded text.

---

## Contact Enrichment

При необходимости возможно подключение легальных enrichment/search API.

Данные должны сохранять:

```text
source
retrieved_at
confidence
```

---

# 8. Базы данных

## PostgreSQL

Основная БД.

Используется для:

```text
users
profiles
vacancies
companies
contacts
applications
CV metadata
AI analysis results
matching results
```

Каждый микросервис должен владеть своей схемой/БД.

Другой сервис не должен напрямую обращаться к его таблицам.

---

## sqlc

Использовать для доступа Go → PostgreSQL.

Преимущества:

- SQL-first;
    
- generated type-safe Go code;
    
- отсутствие тяжёлого ORM;
    
- явные queries.
    

---

## Redis

Использовать для:

- cache;
    
- sessions;
    
- rate limiting;
    
- distributed locks;
    
- temporary processing state;
    
- idempotency keys.
    

Redis не является primary storage.

---

## Object Storage

CV и другие файлы не хранить как `bytea` в PostgreSQL.

Использовать:

```text
S3-compatible storage
```

Для локальной разработки:

```text
MinIO
```

---

## MongoDB

Для MVP **не требуется**.

Полуструктурированные результаты AI можно хранить через:

```text
PostgreSQL JSONB
```

MongoDB добавлять только при появлении конкретной необходимости.

---

# 9. Инфраструктура

## Local Development

```text
Docker Compose
```

Должен поднимать:

```text
services
PostgreSQL
Redis
RabbitMQ/Kafka
MinIO
```

---

## Containerization

Каждый сервис:

```text
Dockerfile
healthcheck
graceful shutdown
configuration via environment
```

---

## Observability

Обязательно:

```text
structured logging
request_id
trace_id
metrics
health checks
```

Рекомендуемый стек:

```text
OpenTelemetry
Prometheus
Grafana
```

Позже:

```text
distributed tracing
```

---

# 10. Надёжность

Все consumers должны быть:

```text
idempotent
```

Повторное получение:

```text
vacancy.created
```

не должно создавать дубликаты.

Для сообщений необходимо предусмотреть:

- retry;
    
- exponential backoff;
    
- dead-letter queue;
    
- idempotency key.
    

---

# 11. Edge Cases и технические риски

## 11.1. Блокировки источников

Риски:

- изменение HTML;
    
- CAPTCHA;
    
- блокировка IP;
    
- rate limiting;
    
- изменение API;
    
- закрытие endpoint;
    
- ограничения Terms of Service.
    

Архитектура источников должна быть adapter-based, чтобы потеря одного источника не ломала систему.

---

## 11.2. Дубликаты вакансий

Одна вакансия может находиться одновременно на нескольких платформах.

Необходимо вычислять fingerprint:

```text
normalized_company
+
normalized_title
+
location
+
description similarity
```

Должна поддерживаться связь:

```text
Vacancy
   ↓
Multiple Sources
```

---

## 11.3. Устаревшие вакансии

Необходимо хранить:

```text
first_seen_at
last_seen_at
published_at
expired_at
status
```

Если источник больше не содержит вакансию, она не должна немедленно удаляться.

---

## 11.4. AI Hallucinations

Критический риск для CV generation.

LLM не является источником истины.

До и после генерации необходимо контролировать:

```text
companies
employment dates
technologies
experience duration
education
certificates
numbers/metrics
```

AI должен использовать исключительно подтверждённую информацию из профиля и исходного CV.

---

## 11.5. AI Rate Limits

Не отправлять LLM-запросы синхронно внутри долгого HTTP request.

Использовать:

```text
API
 ↓
Job created
 ↓
Broker
 ↓
AI Worker
 ↓
Result
```

Frontend получает:

```text
202 Accepted
job_id
```

И затем проверяет status либо получает событие через WebSocket/SSE.

---

## 11.6. Ошибка AI Provider

Необходимо:

```text
timeout
retry
fallback provider
dead-letter queue
```

Запрос должен иметь состояния:

```text
PENDING
PROCESSING
COMPLETED
FAILED
```

---

## 11.7. Большие CV

Ограничить:

```text
file size
page count
supported formats
```

Необходимо проверять MIME type и не доверять только расширению файла.

---

## 11.8. PII

CV содержит персональные данные.

Необходимо:

- ограничивать доступ по ownership;
    
- использовать HTTPS;
    
- не логировать полный CV;
    
- не логировать access tokens;
    
- ограничить хранение raw LLM prompts;
    
- поддерживать удаление пользовательских данных.
    

---

# 12. MVP Scope

Первая версия должна включать только:

```text
Auth Service
Profile Service
Vacancy Service
Source Service
CV Service
AI Service
Matching Service
API Gateway
RabbitMQ/Kafka
PostgreSQL
Redis
MinIO
```

MVP Flow:

```text
Получить вакансию
      ↓
Сохранить
      ↓
AI анализ вакансии
      ↓
Пользователь загружает CV
      ↓
CV parsing
      ↓
Matching
      ↓
Пользователь выбирает вакансию
      ↓
AI tailoring
      ↓
Проверка изменений
      ↓
Экспорт CV
```

Contact Service можно реализовать вторым этапом после стабильной работы core-flow.

---

# 13. Рекомендуемая структура репозитория

На этапе обучения допустим monorepo:

```text
hireflow/
│
├── services/
│   ├── gateway/
│   ├── auth/
│   ├── profile/
│   ├── vacancy/
│   ├── scraper/
│   ├── cv/
│   ├── ai/
│   ├── matching/
│   └── contact/
│
├── proto/
│
├── pkg/
│
├── deployments/
│   └── docker-compose.yml
│
├── migrations/
│
└── docs/
```

Общий `pkg` должен содержать только действительно инфраструктурные компоненты.

Бизнес-модели между сервисами не должны импортироваться напрямую.

---

# 14. Definition of Done для первой технической версии

Первая backend-версия считается готовой, когда пользователь способен пройти цепочку:

```text
Registration
→ Profile
→ Upload CV
→ Parse CV
→ View Vacancy
→ Vacancy AI Analysis
→ Candidate Matching
→ Generate Tailored CV
→ Review Changes
→ Export
```

При этом:

- сервисы запускаются через Docker Compose;
    
- внутреннее синхронное взаимодействие работает через gRPC;
    
- минимум один workflow работает асинхронно через broker;
    
- PostgreSQL используется через sqlc;
    
- Redis используется минимум для cache/idempotency;
    
- AI-задачи имеют retry/error handling;
    
- события обрабатываются идемпотентно;
    
- имеются structured logs и health endpoints.