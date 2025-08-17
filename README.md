# Pinstack Notification Service 🔔

**Pinstack Notification Service** — микросервис для управления уведомлениями пользователей в системе **Pinstack**.

## Основные функции:
- CRUD-операции для уведомлений (создание, чтение, обновление, удаление).
- Отправка и получение пользовательских уведомлений.
- Поддержка различных типов уведомлений (настраивается через тип сообщения).
- Получение событий от других сервисов через **Kafka** (например, relation-events).
- Взаимодействие с другими микросервисами через gRPC.

## Технологии:
- **Go** — основной язык разработки.
- **gRPC** — для межсервисной коммуникации.
- **Kafka** — асинхронная обработка событий и интеграция с другими сервисами.
- **PostgreSQL** — хранение уведомлений.
- **Redis** — кэширование и временное хранение данных.
- **Docker** — для контейнеризации.
- **Prometheus** — для сбора метрик и мониторинга.
- **Grafana** — для визуализации метрик.
- **Loki** — для централизованного сбора логов.

## Архитектура

Проект построен на основе **гексагональной архитектуры (Hexagonal Architecture)** с четким разделением слоев:

### Структура проекта
```
├── cmd/                    # Точки входа приложения
│   ├── server/             # gRPC сервер
│   └── migrate/            # Миграции БД
├── internal/
│   ├── domain/             # Доменный слой
│   │   ├── models/         # Доменные модели
│   │   └── ports/          # Интерфейсы (порты)
│   │       ├── input/      # Входящие порты (use cases)
│   │       └── output/     # Исходящие порты (репозитории, кэш, метрики)
│   ├── application/        # Слой приложения
│   │   └── service/        # Бизнес-логика и сервисы
│   └── infrastructure/     # Инфраструктурный слой
│       ├── inbound/        # Входящие адаптеры (gRPC, Kafka Consumer)
│       │   ├── grpc/       # gRPC обработчики
│       │   └── kafka/      # Kafka потребители
│       └── outbound/       # Исходящие адаптеры (PostgreSQL, Redis, Kafka Producer)
│           ├── repository/ # Репозитории для БД
│           ├── client/     # Клиенты для внешних сервисов
│           └── kafka/      # Kafka производители
├── migrations/             # SQL миграции
└── mocks/                 # Моки для тестирования
```

### Принципы архитектуры
- **Dependency Inversion**: Зависимости направлены к доменному слою
- **Clean Architecture**: Четкое разделение ответственности между слоями
- **Port & Adapter Pattern**: Интерфейсы определяются в domain, реализуются в infrastructure
- **Event-Driven Architecture**: Асинхронная обработка событий через Kafka
- **Testability**: Легкое модульное тестирование благодаря dependency injection

### Мониторинг и метрики
Сервис включает полную интеграцию с системой мониторинга:
- **Prometheus метрики**: Автоматический сбор метрик gRPC, базы данных, Kafka, кэша
- **Structured logging**: Интеграция с Loki для централизованного сбора логов
- **Health checks**: Проверки состояния всех компонентов
- **Performance monitoring**: Метрики времени ответа и throughput

## CI/CD Pipeline 🚀

### GitHub Actions
Проект использует GitHub Actions для автоматического тестирования при каждом push/PR.

**Этапы CI:**
1. **Unit Tests** — юнит-тесты с покрытием кода
2. **Integration Tests** — интеграционные тесты с полной инфраструктурой (включая Kafka)
3. **Auto Cleanup** — автоматическая очистка Docker ресурсов

### Makefile команды 📋

#### Команды разработки

### Настройка и запуск
```bash
# Создание необходимых сетей Docker
docker network create pinstack 2>/dev/null || true
docker network create pinstack-test 2>/dev/null || true

# Запуск легкой среды разработки (только Prometheus stack)
make start-dev-light

# Запуск полной среды разработки (с мониторингом)
make start-dev-full

# Остановка среды разработки
make stop-dev-full
```

### Мониторинг
```bash
# Запуск полного стека мониторинга (Prometheus, Grafana, Loki, ELK)
make start-monitoring

# Запуск только Prometheus stack (Prometheus, Grafana, Loki)
make start-prometheus-stack

# Запуск только ELK stack (Elasticsearch, Logstash, Kibana)
make start-elk-stack

# Остановка мониторинга
make stop-monitoring

# Проверка состояния мониторинга
make check-monitoring-health

# Просмотр логов мониторинга
make logs-prometheus
make logs-grafana
make logs-loki
make logs-elasticsearch
make logs-kibana
```

### Доступ к сервисам мониторинга
- **Prometheus**: http://localhost:9090 - метрики и мониторинг
- **Grafana**: http://localhost:3000 (admin/admin) - дашборды и визуализация
- **Loki**: http://localhost:3100 - сбор логов
- **Kibana**: http://localhost:5601 - анализ логов ELK
- **Elasticsearch**: http://localhost:9200 - поиск и хранение логов
- **PgAdmin**: http://localhost:5050 (admin@admin.com/admin) - управление БД
- **Kafka UI**: http://localhost:9091 - управление Kafka

### Основные команды разработки
```bash
# Проверка кода и тесты
make fmt                    # Форматирование кода (gofmt)
make lint                   # Проверка кода (go vet + golangci-lint)
make test-unit              # Юнит-тесты с покрытием
make test-integration       # Интеграционные тесты (с Docker + Kafka + Redis)
make test-all               # Все тесты: форматирование + линтер + юнит + интеграционные

# CI локально
make ci-local               # Полный CI процесс локально (имитация GitHub Actions)
```

#### Управление инфраструктурой:
```bash
# Настройка репозитория
make setup-system-tests              # Клонирует/обновляет pinstack-system-tests репозиторий
make setup-monitoring               # Клонирует/обновляет pinstack-monitoring-service репозиторий

# Запуск инфраструктуры
make start-notification-infrastructure  # Поднимает все Docker контейнеры для тестов
make check-services                    # Проверяет готовность всех сервисов

# Интеграционные тесты
make test-notification-integration     # Запускает только интеграционные тесты
make quick-test-local                 # Быстрый запуск тестов с локальным notification-service
make quick-test                       # Быстрый запуск тестов без пересборки контейнеров

# Остановка и очистка
make stop-notification-infrastructure  # Останавливает все тестовые контейнеры
make clean-notification-infrastructure # Полная очистка (контейнеры + volumes + образы)
make clean                           # Полная очистка проекта + Docker
```

#### Логи и отладка:
```bash
# Просмотр логов сервисов
make logs-notification      # Логи Notification Service
make logs-notification-db   # Логи Notification Database
make logs-kafka            # Логи Kafka
make logs-user             # Логи User Service
make logs-auth             # Логи Auth Service  
make logs-gateway          # Логи API Gateway
make logs-db               # Логи User Database
make logs-auth-db          # Логи Auth Database
make logs-redis            # Логи Redis

# Redis утилиты для отладки
make redis-cli             # Подключение к Redis CLI
make redis-info            # Информация о Redis
make redis-keys            # Показать все ключи в Redis
make redis-flush           # Очистить все данные Redis (с подтверждением)

# Kafka утилиты для отладки
make kafka-topics          # Список всех топиков Kafka
make kafka-create-topic    # Создать новый топик
make kafka-describe-topic  # Описание топика
make kafka-consumer        # Подключиться к потребителю топика
make kafka-producer        # Подключиться к производителю топика
make kafka-delete-topic    # Удалить топик (с подтверждением)

# Экстренная очистка
make clean-docker-force    # Удаляет ВСЕ Docker ресурсы (с подтверждением)
```

### Зависимости для интеграционных тестов 🐳

Для интеграционных тестов автоматически поднимаются контейнеры:
- **notification-db-test** — PostgreSQL для Notification Service
- **notification-migrator-test** — миграции Notification Service
- **notification-service-test** — сам Notification Service
- **kafka-test** — Apache Kafka для асинхронной обработки
- **kafka-topics-init-test** — инициализация топиков Kafka
- **redis** — Redis для кэширования и временного хранения
- **user-db-test** — PostgreSQL для User Service
- **user-migrator-test** — миграции User Service
- **user-service-test** — User Service (для валидации пользователей)
- **auth-db-test** — PostgreSQL для Auth Service
- **auth-migrator-test** — миграции Auth Service
- **auth-service-test** — Auth Service
- **api-gateway-test** — API Gateway

> 📍 **Требования:** Docker, docker-compose  
> 🚀 **Все сервисы собираются автоматически из Git репозиториев**  
> 🔄 **Репозитории `pinstack-system-tests` и `pinstack-monitoring-service` клонируются автоматически при запуске**
> ⚡ **Kafka и Redis готовы к работе с предустановленными настройками**

### Быстрый старт разработки ⚡

```bash
# 1. Проверить код
make fmt lint

# 2. Запустить юнит-тесты
make test-unit

# 3. Запустить интеграционные тесты (включая Kafka и Redis)
make test-integration

# 4. Или всё сразу
make ci-local

# 5. Запуск полной среды разработки с мониторингом
make start-dev-full

# 6. Очистка после работы
make clean
```

### Комбинированные команды для разработки 🚀

```bash
# Полная среда разработки с мониторингом
make start-dev-full     # Запускает все сервисы + мониторинг
# Доступ к сервисам:
# 🔗 API Gateway: http://localhost:8080
# 👤 User Service: http://localhost:8081  
# 🔐 Auth Service: http://localhost:8082
# 📨 Notification Service: http://localhost:8083
# 📊 Prometheus: http://localhost:9090
# 📈 Grafana: http://localhost:3000 (admin/admin)
# 🔍 Loki: http://localhost:3100
# 📋 Kibana: http://localhost:5601

# Легкая среда разработки (без ELK stack)
make start-dev-light    # Запускает сервисы + Prometheus stack

# Остановка всей среды
make stop-dev-full      # Останавливает все сервисы и мониторинг

# Полная очистка
make clean-dev-full     # Очищает все ресурсы разработки
```

### Особенности 🔧

- **Отключение кеша тестов:** все тесты запускаются с флагом `-count=1`
- **Фокус на Notification Service:** интеграционные тесты тестируют только Notification endpoints
- **Kafka интеграция:** тесты включают проверку асинхронной обработки через Kafka
- **Redis кэширование:** интеграция с Redis для временного хранения данных
- **Автоочистка:** CI автоматически удаляет все Docker ресурсы после себя
- **Параллельность:** в CI юнит и интеграционные тесты запускаются последовательно
- **User Service валидация:** интеграция с User Service для проверки существования пользователей
- **Полный мониторинг:** интеграция с Prometheus, Grafana, Loki и ELK stack
- **Docker сети:** автоматическое создание сетей `pinstack` и `pinstack-test`

> ✅ Сервис готов к использованию.
