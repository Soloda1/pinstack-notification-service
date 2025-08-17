.PHONY: test test-unit test-integration test-notification-integration clean build run docker-build setup-system-tests setup-monitoring start-monitoring start-prometheus-stack start-elk-stack stop-monitoring clean-monitoring check-monitoring-health logs-prometheus logs-grafana logs-loki logs-elasticsearch logs-kibana start-dev-full stop-dev-full clean-dev-full start-dev-light

BINARY_NAME=notification-service
DOCKER_IMAGE=pinstack-notification-service:latest
GO_VERSION=1.24.2
SYSTEM_TESTS_DIR=../pinstack-system-tests
SYSTEM_TESTS_REPO=https://github.com/Soloda1/pinstack-system-tests.git
MONITORING_DIR=../pinstack-monitoring-service
MONITORING_REPO=https://github.com/Soloda1/pinstack-monitoring-service.git

# Проверка версии Go
check-go-version:
	@echo "🔍 Проверка версии Go..."
	@go version | grep -q "go$(GO_VERSION)" || (echo "❌ Требуется Go $(GO_VERSION)" && exit 1)
	@echo "✅ Go $(GO_VERSION) найден"

# Настройка monitoring репозитория
setup-monitoring:
	@echo "🔄 Проверка monitoring репозитория..."
	@if [ ! -d "$(MONITORING_DIR)" ]; then \
		echo "📥 Клонирование pinstack-monitoring-service..."; \
		git clone $(MONITORING_REPO) $(MONITORING_DIR); \
	else \
		echo "🔄 Обновление pinstack-monitoring-service..."; \
		cd $(MONITORING_DIR) && git pull origin main; \
	fi
	@echo "✅ Monitoring готов"

# Настройка system tests репозитория
setup-system-tests:
	@echo "🔄 Проверка system tests репозитория..."
	@if [ ! -d "$(SYSTEM_TESTS_DIR)" ]; then \
		echo "📥 Клонирование pinstack-system-tests..."; \
		git clone $(SYSTEM_TESTS_REPO) $(SYSTEM_TESTS_DIR); \
	else \
		echo "🔄 Обновление pinstack-system-tests..."; \
		cd $(SYSTEM_TESTS_DIR) && git pull origin main; \
	fi
	@echo "✅ System tests готовы"

# Форматирование и проверки
fmt: check-go-version
	gofmt -s -w .
	go fmt ./...

lint: check-go-version
	go vet ./...
	golangci-lint run

# Юнит тесты
test-unit: check-go-version
	go test -v -count=1 -race -coverprofile=coverage.txt ./...

# Запуск полной инфраструктуры для интеграционных тестов из существующего docker-compose
start-notification-infrastructure: setup-system-tests
	@echo "🚀 Запуск полной инфраструктуры для интеграционных тестов..."
	@echo "🔍 Проверка и создание сетей..."
	@docker network create pinstack 2>/dev/null || true
	@docker network create pinstack-test 2>/dev/null || true
	cd $(SYSTEM_TESTS_DIR) && \
	NOTIFICATION_SERVICE_CONTEXT=../pinstack-notification-service docker compose -f docker-compose.test.yml up -d \
		user-db-test \
		user-migrator-test \
		user-service-test \
		auth-db-test \
		auth-migrator-test \
		auth-service-test \
		api-gateway-test \
		notification-db-test \
		notification-migrator-test \
		notification-service-test \
		kafka-test \
		kafka-topics-init-test \
		redis
	@echo "⏳ Ожидание готовности сервисов..."
	@sleep 30

# Проверка готовности сервисов
check-services:
	@echo "🔍 Проверка готовности сервисов..."
	@docker exec pinstack-user-db-test pg_isready -U postgres || (echo "❌ User база данных не готова" && exit 1)
	@docker exec pinstack-auth-db-test pg_isready -U postgres || (echo "❌ Auth база данных не готова" && exit 1)
	@docker exec pinstack-notification-db-test pg_isready -U postgres || (echo "❌ Notification база данных не готова" && exit 1)
	@timeout 30 bash -c 'until docker exec pinstack-redis-test redis-cli ping | grep -q PONG; do echo "⏳ Ожидание Redis..."; sleep 2; done' || (echo "❌ Redis не готов" && exit 1)
	@timeout 120 bash -c 'until docker exec pinstack-kafka-test kafka-topics --bootstrap-server localhost:9092 --list > /dev/null 2>&1; do echo "⏳ Ожидание Kafka..."; sleep 5; done' || (echo "❌ Kafka не готов" && exit 1)
	@echo "✅ Базы данных, Redis и Kafka готовы"
	@echo "=== User Service logs ==="
	@docker logs pinstack-user-service-test --tail=10
	@echo "=== Auth Service logs ==="
	@docker logs pinstack-auth-service-test --tail=10
	@echo "=== Notification Service logs ==="
	@docker logs pinstack-notification-service-test --tail=10
	@echo "=== API Gateway logs ==="
	@docker logs pinstack-api-gateway-test --tail=10
	@echo "=== Kafka logs ==="
	@docker logs pinstack-kafka-test --tail=10
	@echo "=== Redis logs ==="
	@docker logs pinstack-redis-test --tail=5

# Интеграционные тесты только для notification service
test-notification-integration: check-services
	@echo "🧪 Запуск интеграционных тестов для Notification Service..."
	cd $(SYSTEM_TESTS_DIR) && \
	go test -v -count=1 -timeout=10m ./internal/scenarios/integration/gateway_notification/...

# Остановка всех контейнеров
stop-notification-infrastructure:
	@echo "🛑 Остановка всей инфраструктуры..."
	cd $(SYSTEM_TESTS_DIR) && \
	docker compose -f docker-compose.test.yml stop \
		api-gateway-test \
		notification-service-test \
		notification-migrator-test \
		notification-db-test \
		kafka-test \
		kafka-topics-init-test \
		auth-service-test \
		auth-migrator-test \
		auth-db-test \
		user-service-test \
		user-migrator-test \
		user-db-test \
		redis
	cd $(SYSTEM_TESTS_DIR) && \
	docker compose -f docker-compose.test.yml rm -f \
		api-gateway-test \
		notification-service-test \
		notification-migrator-test \
		notification-db-test \
		kafka-test \
		kafka-topics-init-test \
		auth-service-test \
		auth-migrator-test \
		auth-db-test \
		user-service-test \
		user-migrator-test \
		user-db-test \
		redis

# Полная очистка (включая volumes)
clean-notification-infrastructure:
	@echo "🧹 Полная очистка всей инфраструктуры..."
	cd $(SYSTEM_TESTS_DIR) && \
	docker compose -f docker-compose.test.yml down -v
	@echo "🧹 Очистка Docker контейнеров, образов и volumes..."
	docker container prune -f
	docker image prune -a -f
	docker volume prune -f
	docker network prune -f
	@echo "✅ Полная очистка завершена"

# Полные интеграционные тесты (с очисткой)
test-integration: start-notification-infrastructure test-notification-integration stop-notification-infrastructure

# Все тесты
test-all: fmt lint test-unit test-integration

# Логи сервисов
logs-notification:
	cd $(SYSTEM_TESTS_DIR) && \
	docker compose -f docker-compose.test.yml logs -f notification-service-test

logs-notification-db:
	cd $(SYSTEM_TESTS_DIR) && \
	docker compose -f docker-compose.test.yml logs -f notification-db-test

logs-kafka:
	cd $(SYSTEM_TESTS_DIR) && \
	docker compose -f docker-compose.test.yml logs -f kafka-test

logs-user:
	cd $(SYSTEM_TESTS_DIR) && \
	docker compose -f docker-compose.test.yml logs -f user-service-test

logs-auth:
	cd $(SYSTEM_TESTS_DIR) && \
	docker compose -f docker-compose.test.yml logs -f auth-service-test

logs-gateway:
	cd $(SYSTEM_TESTS_DIR) && \
	docker compose -f docker-compose.test.yml logs -f api-gateway-test

logs-db:
	cd $(SYSTEM_TESTS_DIR) && \
	docker compose -f docker-compose.test.yml logs -f user-db-test

logs-auth-db:
	cd $(SYSTEM_TESTS_DIR) && \
	docker compose -f docker-compose.test.yml logs -f auth-db-test

logs-redis:
	cd $(SYSTEM_TESTS_DIR) && \
	docker compose -f docker-compose.test.yml logs -f redis

# Redis утилиты для отладки
redis-cli:
	@echo "🔍 Подключение к Redis CLI..."
	docker exec -it pinstack-redis-test redis-cli

redis-info:
	@echo "📊 Информация о Redis..."
	docker exec pinstack-redis-test redis-cli info

redis-keys:
	@echo "🔑 Все ключи в Redis..."
	docker exec pinstack-redis-test redis-cli keys "*"

redis-flush:
	@echo "🧹 Очистка всех данных Redis..."
	@read -p "Очистить все данные Redis? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	docker exec pinstack-redis-test redis-cli flushall
	@echo "✅ Redis очищен"

# Kafka утилиты для отладки
kafka-topics:
	@echo "📋 Список топиков Kafka..."
	docker exec pinstack-kafka-test kafka-topics --bootstrap-server localhost:9092 --list

kafka-create-topic:
	@read -p "Введите имя топика: " topic && \
	docker exec pinstack-kafka-test kafka-topics --bootstrap-server localhost:9092 --create --topic $$topic --partitions 1 --replication-factor 1
	@echo "✅ Топик создан"

kafka-describe-topic:
	@read -p "Введите имя топика: " topic && \
	docker exec pinstack-kafka-test kafka-topics --bootstrap-server localhost:9092 --describe --topic $$topic

kafka-consumer:
	@read -p "Введите имя топика: " topic && \
	echo "🔍 Подключение к потребителю Kafka для топика $$topic (Ctrl+C для выхода)..." && \
	docker exec -it pinstack-kafka-test kafka-console-consumer --bootstrap-server localhost:9092 --topic $$topic --from-beginning

kafka-producer:
	@read -p "Введите имя топика: " topic && \
	echo "📤 Подключение к производителю Kafka для топика $$topic (введите сообщения, Ctrl+C для выхода)..." && \
	docker exec -it pinstack-kafka-test kafka-console-producer --bootstrap-server localhost:9092 --topic $$topic

kafka-delete-topic:
	@read -p "Введите имя топика для удаления: " topic && \
	read -p "Удалить топик $$topic? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1 && \
	docker exec pinstack-kafka-test kafka-topics --bootstrap-server localhost:9092 --delete --topic $$topic
	@echo "✅ Топик удален"

# Быстрый тест с локальным notification-service
quick-test-local: setup-system-tests
	@echo "⚡ Быстрый запуск тестов с локальным notification-service..."
	@echo "🔍 Проверка и создание сетей..."
	@docker network create pinstack 2>/dev/null || true
	@docker network create pinstack-test 2>/dev/null || true
	cd $(SYSTEM_TESTS_DIR) && \
	NOTIFICATION_SERVICE_CONTEXT=../pinstack-notification-service docker compose -f docker-compose.test.yml up -d \
		user-db-test user-migrator-test user-service-test \
		auth-db-test auth-migrator-test auth-service-test \
		api-gateway-test notification-db-test notification-migrator-test notification-service-test \
		kafka-test kafka-topics-init-test redis
	@echo "⏳ Ожидание готовности сервисов..."
	@sleep 30
	@timeout 30 bash -c 'until docker exec pinstack-redis-test redis-cli ping | grep -q PONG; do echo "⏳ Ожидание Redis..."; sleep 2; done'
	@timeout 120 bash -c 'until docker exec pinstack-kafka-test kafka-topics --bootstrap-server localhost:9092 --list > /dev/null 2>&1; do echo "⏳ Ожидание Kafka..."; sleep 5; done'
	cd $(SYSTEM_TESTS_DIR) && \
	go test -v -count=1 -timeout=5m ./internal/scenarios/integration/gateway_notification/...
	$(MAKE) stop-notification-infrastructure

# Очистка
clean: clean-notification-infrastructure
	go clean
	rm -f $(BINARY_NAME)
	@echo "🧹 Финальная очистка Docker системы..."
	docker system prune -a -f --volumes
	@echo "✅ Вся очистка завершена"

# Экстренная полная очистка Docker (если что-то пошло не так)
clean-docker-force:
	@echo "🚨 ЭКСТРЕННАЯ ПОЛНАЯ ОЧИСТКА DOCKER..."
	@echo "⚠️  Это удалит ВСЕ Docker контейнеры, образы, volumes и сети!"
	@read -p "Продолжить? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	docker stop $$(docker ps -aq) 2>/dev/null || true
	docker rm $$(docker ps -aq) 2>/dev/null || true
	docker rmi $$(docker images -q) 2>/dev/null || true
	docker volume rm $$(docker volume ls -q) 2>/dev/null || true
	docker network rm $$(docker network ls -q) 2>/dev/null || true
	docker system prune -a -f --volumes
	@echo "💥 Экстренная очистка завершена"

# CI локально (имитация GitHub Actions)
ci-local: test-all
	@echo "🎉 Локальный CI завершен успешно!"

# Быстрый тест (только запуск без пересборки)
quick-test: start-notification-infrastructure
	@echo "⚡ Быстрый запуск тестов без пересборки..."
	cd $(SYSTEM_TESTS_DIR) && \
	go test -v -count=1 -timeout=5m ./internal/scenarios/integration/gateway_notification/...
	$(MAKE) stop-notification-infrastructure

######################
# Monitoring Stack   #
######################

# Запуск полного monitoring stack
start-monitoring: setup-monitoring
	@echo "📊 Запуск monitoring stack..."
	@echo "🔍 Проверка и создание сетей..."
	@docker network create pinstack 2>/dev/null || true
	@docker network create pinstack-test 2>/dev/null || true
	cd $(MONITORING_DIR) && \
	docker compose up -d
	@echo "⏳ Ожидание готовности monitoring сервисов..."
	@sleep 15
	@echo "✅ Monitoring stack запущен:"
	@echo "  📊 Prometheus: http://localhost:9090"
	@echo "  📈 Grafana: http://localhost:3000 (admin/admin)"
	@echo "  🔍 Loki: http://localhost:3100"
	@echo "  📋 Kibana: http://localhost:5601"
	@echo "  💾 Elasticsearch: http://localhost:9200"
	@echo "  🐧 PgAdmin: http://localhost:5050 (admin@admin.com/admin)"
	@echo "  🐛 Kafka UI: http://localhost:9091"

# Запуск только Prometheus stack (Prometheus + Grafana + Loki)
start-prometheus-stack: setup-monitoring
	@echo "📊 Запуск Prometheus stack..."
	@echo "🔍 Проверка и создание сетей..."
	@docker network create pinstack 2>/dev/null || true
	@docker network create pinstack-test 2>/dev/null || true
	cd $(MONITORING_DIR) && \
	docker compose up -d prometheus grafana loki promtail
	@echo "⏳ Ожидание готовности Prometheus stack..."
	@sleep 10
	@echo "✅ Prometheus stack запущен:"
	@echo "  📊 Prometheus: http://localhost:9090"
	@echo "  📈 Grafana: http://localhost:3000 (admin/admin)"
	@echo "  🔍 Loki: http://localhost:3100"

# Запуск только ELK stack
start-elk-stack: setup-monitoring
	@echo "📊 Запуск ELK stack..."
	@echo "🔍 Проверка и создание сетей..."
	@docker network create pinstack 2>/dev/null || true
	@docker network create pinstack-test 2>/dev/null || true
	cd $(MONITORING_DIR) && \
	docker compose up -d elasticsearch logstash kibana filebeat
	@echo "⏳ Ожидание готовности ELK stack..."
	@sleep 30
	@echo "✅ ELK stack запущен:"
	@echo "  📋 Kibana: http://localhost:5601"
	@echo "  💾 Elasticsearch: http://localhost:9200"

# Проверка состояния monitoring сервисов
check-monitoring-health:
	@echo "🔍 Проверка состояния monitoring сервисов..."
	@echo "Prometheus:" && curl -s http://localhost:9090/-/healthy | head -1 || echo "❌ Prometheus недоступен"
	@echo "Grafana:" && curl -s http://localhost:3000/api/health | head -1 || echo "❌ Grafana недоступна"
	@echo "Loki:" && curl -s http://localhost:3100/ready | head -1 || echo "❌ Loki недоступен"
	@echo "Elasticsearch:" && curl -s http://localhost:9200/_cluster/health | head -1 || echo "❌ Elasticsearch недоступен"
	@echo "Kibana:" && curl -s http://localhost:5601/api/status | head -1 || echo "❌ Kibana недоступна"

# Остановка monitoring stack
stop-monitoring:
	@echo "🛑 Остановка monitoring stack..."
	@if [ -d "$(MONITORING_DIR)" ]; then \
		cd $(MONITORING_DIR) && docker compose stop; \
	else \
		echo "⚠️  Monitoring директория не найдена"; \
	fi

# Полная очистка monitoring stack
clean-monitoring:
	@echo "🧹 Очистка monitoring stack..."
	@if [ -d "$(MONITORING_DIR)" ]; then \
		cd $(MONITORING_DIR) && docker compose down -v; \
		echo "🧹 Очистка monitoring volumes..."; \
		docker volume rm pinstack-monitoring-service_elasticsearch_data 2>/dev/null || true; \
		docker volume rm pinstack-monitoring-service_filebeat_data 2>/dev/null || true; \
		echo "✅ Monitoring stack очищен"; \
	else \
		echo "⚠️  Monitoring директория не найдена"; \
	fi

# Логи monitoring сервисов
logs-prometheus:
	@if [ -d "$(MONITORING_DIR)" ]; then \
		cd $(MONITORING_DIR) && docker compose logs -f prometheus; \
	else \
		echo "⚠️  Monitoring директория не найдена"; \
	fi

logs-grafana:
	@if [ -d "$(MONITORING_DIR)" ]; then \
		cd $(MONITORING_DIR) && docker compose logs -f grafana; \
	else \
		echo "⚠️  Monitoring директория не найдена"; \
	fi

logs-loki:
	@if [ -d "$(MONITORING_DIR)" ]; then \
		cd $(MONITORING_DIR) && docker compose logs -f loki; \
	else \
		echo "⚠️  Monitoring директория не найдена"; \
	fi

logs-elasticsearch:
	@if [ -d "$(MONITORING_DIR)" ]; then \
		cd $(MONITORING_DIR) && docker compose logs -f elasticsearch; \
	else \
		echo "⚠️  Monitoring директория не найдена"; \
	fi

logs-kibana:
	@if [ -d "$(MONITORING_DIR)" ]; then \
		cd $(MONITORING_DIR) && docker compose logs -f kibana; \
	else \
		echo "⚠️  Monitoring директория не найдена"; \
	fi

# Комбинированные команды

# Полный development environment с мониторингом
start-dev-full: setup-monitoring start-monitoring start-notification-infrastructure
	@echo "🚀 Полная dev среда запущена!"
	@echo ""
	@echo "=== Приложения ==="
	@echo "  🔗 API Gateway: http://localhost:8080"
	@echo "  👤 User Service: http://localhost:8081"
	@echo "  🔐 Auth Service: http://localhost:8082"
	@echo "  📨 Notification Service: http://localhost:8083"
	@echo ""
	@echo "=== Мониторинг ==="
	@echo "  📊 Prometheus: http://localhost:9090"
	@echo "  📈 Grafana: http://localhost:3000 (admin/admin)"
	@echo "  🔍 Loki: http://localhost:3100"
	@echo "  📋 Kibana: http://localhost:5601"
	@echo ""
	@echo "=== Базы данных ==="
	@echo "  🐧 PgAdmin: http://localhost:5050 (admin@admin.com/admin)"
	@echo "  🔴 Redis: localhost:6379"
	@echo "  🐛 Kafka UI: http://localhost:9091"

# Остановка всей dev среды
stop-dev-full: stop-monitoring stop-notification-infrastructure
	@echo "🛑 Полная dev среда остановлена"

# Очистка всей dev среды
clean-dev-full: clean-monitoring clean-notification-infrastructure
	@echo "🧹 Полная dev среда очищена"

# Запуск только с Prometheus stack (без ELK)
start-dev-light: setup-monitoring start-prometheus-stack start-notification-infrastructure
	@echo "🚀 Легкая dev среда запущена (без ELK stack)!"
	@echo ""
	@echo "=== Приложения ==="
	@echo "  🔗 API Gateway"
	@echo "  👤 User Service"
	@echo "  🔐 Auth Service"
	@echo "  📨 Notification Service"
	@echo ""
	@echo "=== Мониторинг ==="
	@echo "  📊 Prometheus: http://localhost:9090"
	@echo "  📈 Grafana: http://localhost:3000 (admin/admin)"
	@echo "  🔍 Loki: http://localhost:3100"