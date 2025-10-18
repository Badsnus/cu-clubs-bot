# ================================================================================================
# 🎯 CU Clubs Bot - Development Makefile
# ================================================================================================

# Цвета для красивого вывода
RESET := $(shell tput sgr0 2>/dev/null || echo "")
RED := $(shell tput setaf 1 2>/dev/null || echo "")
GREEN := $(shell tput setaf 2 2>/dev/null || echo "")
YELLOW := $(shell tput setaf 3 2>/dev/null || echo "")
BLUE := $(shell tput setaf 4 2>/dev/null || echo "")
MAGENTA := $(shell tput setaf 5 2>/dev/null || echo "")
CYAN := $(shell tput setaf 6 2>/dev/null || echo "")
WHITE := $(shell tput setaf 7 2>/dev/null || echo "")
BOLD := $(shell tput bold 2>/dev/null || echo "")

# Конфигурация
DOCKER_COMPOSE_DEV := docker-compose-dev.yml
DOCKER_COMPOSE_PROD := docker-compose.yml
BOT_CONTAINER := bot
DB_CONTAINER := database
REDIS_CONTAINER := redis

# Настройки для продового деплоя
PROD_OWNER ?= badsnus
PROD_REPO  ?= cu-clubs-bot
PROD_REF   ?= main2
PROD_PATH  ?= docker-compose.yml
PROD_COMPOSE_FILE ?= docker-compose.yml
PROD_COMPOSE_URL ?= https://raw.githubusercontent.com/$(PROD_OWNER)/$(PROD_REPO)/$(PROD_REF)/$(PROD_PATH)

# По умолчанию показываем help
.DEFAULT_GOAL := help

# ================================================================================================
# 📋 СПРАВКА И ИНФОРМАЦИЯ
# ================================================================================================

.PHONY: help version

help: ## 📋 Показать справку (это сообщение)
	@printf "$(GREEN)$(BOLD)🎯 CU Clubs Bot - Makefile$(RESET)\n"
	@printf "$(CYAN)Доступные команды:$(RESET)\n"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'

version: ## ℹ️ Показать версию проекта
	@printf "$(GREEN)$(BOLD)ℹ️ Версия проекта:$(RESET)\n"
	@printf "$(BLUE)Ветка:$(RESET) $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'неизвестно')\n"
	@printf "$(BLUE)Commit:$(RESET) $(shell git rev-parse --short HEAD 2>/dev/null || echo 'неизвестно')\n"
	@printf "$(BLUE)Дата:$(RESET) $(shell git log -1 --format=%cd --date=format:"%Y-%m-%d %H:%M:%S" 2>/dev/null || echo 'неизвестно')\n"

# ================================================================================================
# 🔄 ОСНОВНЫЕ КОМАНДЫ РАЗРАБОТКИ
# ================================================================================================

.PHONY: dev build start stop restart status logs shell

# Основные команды для разработки
dev: build up ## 🚀 Запустить проект в режиме разработки (синоним для build + up)

build: ## 🔨 Собрать образ
	@printf "$(BLUE)$(BOLD)🔨 Собираю образ...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) build

up: ## ▶️ Запустить контейнеры в foreground режиме
	@printf "$(GREEN)$(BOLD)▶️ Запускаю контейнеры...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) up

start: ## ▶️ Запустить контейнеры в background режиме
	@printf "$(GREEN)$(BOLD)▶️ Запускаю контейнеры в фоновом режиме...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) up -d

stop: ## ⏹️ Остановить контейнеры
	@printf "$(RED)$(BOLD)⏹️ Останавливаю контейнеры...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) down

restart: ## 🔄 Перезапустить все контейнеры
	@printf "$(YELLOW)$(BOLD)🔄 Перезапускаю все контейнеры...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) down
	@docker compose -f $(DOCKER_COMPOSE_DEV) up -d

restart-bot: ## 🔄 Перезапустить только бот
	@printf "$(YELLOW)$(BOLD)🔄 Перезапускаю только бот...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) restart $(BOT_CONTAINER)

status: ## ℹ️ Показать статус контейнеров
	@printf "$(CYAN)$(BOLD)ℹ️ Статус контейнеров:$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) ps

logs: ## 📃 Показать логи всех контейнеров
	@printf "$(CYAN)$(BOLD)📃 Логи:$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) logs -f

# ================================================================================================
# 🖥️ РАБОТА С КОНТЕЙНЕРАМИ
# ================================================================================================

.PHONY: logs-bot logs-db logs-redis shell-bot shell-db shell-redis rebuild clean clean-all

logs-bot: ## 📃 Показать логи бота
	@printf "$(CYAN)$(BOLD)📃 Логи бота:$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) logs -f $(BOT_CONTAINER)

logs-db: ## 📃 Показать логи базы данных
	@printf "$(CYAN)$(BOLD)📃 Логи базы данных:$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) logs -f $(DB_CONTAINER)

logs-redis: ## 📃 Показать логи Redis
	@printf "$(CYAN)$(BOLD)📃 Логи Redis:$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) logs -f $(REDIS_CONTAINER)

shell-bot: ## 💻 Запустить оболочку в контейнере бота
	@printf "$(BLUE)$(BOLD)💻 Запускаю оболочку в контейнере бота...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) exec $(BOT_CONTAINER) /bin/sh

shell-db: ## 💻 Запустить оболочку в контейнере базы данных
	@printf "$(BLUE)$(BOLD)💻 Запускаю оболочку в контейнере базы данных...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) exec $(DB_CONTAINER) /bin/sh

shell-redis: ## 💻 Запустить оболочку в контейнере Redis
	@printf "$(BLUE)$(BOLD)💻 Запускаю оболочку в контейнере Redis...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) exec $(REDIS_CONTAINER) redis-cli -a "$${REDIS_PASSWORD}"

rebuild: ## 🔨 Пересобрать проект с нуля
	@printf "$(BLUE)$(BOLD)🔨 Пересобираю проект...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) down
	@docker compose -f $(DOCKER_COMPOSE_DEV) build --no-cache

clean: ## 🧹 Очистить контейнеры
	@printf "$(RED)$(BOLD)🧹 Очищаю контейнеры...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) down -v

clean-all: ## 🧹 Очистить все контейнеры и образы
	@printf "$(RED)$(BOLD)⚠️ ВНИМАНИЕ: Вы собираетесь удалить ВСЕ контейнеры, образы и тома Docker!$(RESET)\n"
	@printf "$(YELLOW)Эта операция необратима и приведет к потере всех данных в контейнерах.$(RESET)\n"
	@read -p "Вы уверены, что хотите продолжить? (y/N): " confirm && [ "$$confirm" = "y" ] || { printf "$(GREEN)Операция отменена.$(RESET)\n"; exit 1; }
	@printf "$(RED)$(BOLD)🧹 Очищаю все контейнеры и образы...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_DEV) down -v
	@docker system prune -af --volumes

# ================================================================================================
# 🧪 GO РАЗРАБОТКА
# ================================================================================================

.PHONY: go-test go-test-coverage go-lint go-lint-fix go-fmt go-vet go-mod go-deps go-build go-run go-check

go-test: ## 🧪 Запустить Go тесты
	@printf "$(BLUE)$(BOLD)🧪 Запускаю тесты...$(RESET)\n"
	@cd bot && go test ./... -v

go-test-coverage: ## 📊 Запустить тесты и показать покрытие
	@printf "$(BLUE)$(BOLD)📊 Запускаю тесты с покрытием...$(RESET)\n"
	@cd bot && go test ./... -coverprofile=coverage.out
	@cd bot && go tool cover -func=coverage.out
	@cd bot && go tool cover -html=coverage.out -o coverage.html
	@printf "$(GREEN)✅ Отчет о покрытии создан: bot/coverage.html$(RESET)\n"

go-lint: ## 🔍 Запустить линтер (golangci-lint)
	@printf "$(BLUE)$(BOLD)🔍 Запускаю линтер...$(RESET)\n"
	@cd bot && golangci-lint run --timeout=5m

go-lint-fix: ## 🔧 Запустить линтер с автофиксом проблем
	@printf "$(BLUE)$(BOLD)🔧 Запускаю линтер с автофиксом...$(RESET)\n"
	@cd bot && golangci-lint run --timeout=5m --fix

go-fmt: ## 🎨 Форматировать Go код
	@printf "$(BLUE)$(BOLD)🎨 Форматирую код...$(RESET)\n"
	@test -f "$$(go env GOPATH)/bin/goimports" || { printf "$(RED)❌ goimports не найден. Запустите: make install-tools$(RESET)\n"; exit 1; }
	@test -f "$$(go env GOPATH)/bin/gofumpt" || { printf "$(RED)❌ gofumpt не найден. Запустите: make install-tools$(RESET)\n"; exit 1; }
	@cd bot && $$(go env GOPATH)/bin/gofumpt -l -w .
	@cd bot && $$(go env GOPATH)/bin/goimports -w -local github.com/Badsnus/cu-clubs-bot .
	@printf "$(GREEN)✅ Код отформатирован$(RESET)\n"

go-vet: ## 🔍 Запустить go vet
	@printf "$(BLUE)$(BOLD)🔍 Запускаю go vet...$(RESET)\n"
	@cd bot && go vet ./...

go-mod: ## 📦 Обновить зависимости Go
	@printf "$(BLUE)$(BOLD)📦 Обновляю зависимости...$(RESET)\n"
	@cd bot && go mod tidy

go-deps: ## 📋 Показать список зависимостей
	@printf "$(BLUE)$(BOLD)📋 Список зависимостей:$(RESET)\n"
	@cd bot && go list -m all

go-build: ## 🔨 Собрать Go приложение
	@printf "$(BLUE)$(BOLD)🔨 Собираю Go приложение...$(RESET)\n"
	@cd bot && go build -v ./...
	@printf "$(GREEN)✅ Сборка завершена$(RESET)\n"

go-run: ## ▶️ Запустить Go приложение
	@printf "$(GREEN)$(BOLD)▶️ Запускаю Go приложение...$(RESET)\n"
	@cd bot && go run ./cmd/app/main.go

go-check: go-fmt go-vet go-lint go-test ## 🔍 Запустить все проверки кода

# ================================================================================================
# 🛠️ ИНСТРУМЕНТЫ
# ================================================================================================

.PHONY: install-tools git-hooks git-pre-commit

install-tools: ## 📥 Установить все необходимые инструменты
	@printf "$(YELLOW)$(BOLD)📥 Устанавливаю Go инструменты...$(RESET)\n"
	@which golangci-lint >/dev/null 2>&1 || { \
		printf "$(YELLOW)Устанавливаю golangci-lint...$(RESET)\n"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.5.0; \
	}
	@[ -f $$(go env GOPATH)/bin/goimports ] || { \
		printf "$(YELLOW)Устанавливаю goimports...$(RESET)\n"; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	}
	@[ -f $$(go env GOPATH)/bin/gofumpt ] || { \
		printf "$(YELLOW)Устанавливаю gofumpt...$(RESET)\n"; \
		go install mvdan.cc/gofumpt@latest; \
	}

	@printf "$(GREEN)✅ Все инструменты установлены$(RESET)\n"
	@printf "$(GREEN)golangci-lint: $$(golangci-lint version 2>/dev/null || echo 'не найден')$(RESET)\n"
	@printf "$(GREEN)goimports: $$($$(go env GOPATH)/bin/goimports -h 2>&1 | grep -q "usage:" && echo 'установлен' || echo 'не найден')$(RESET)\n"
	@printf "$(GREEN)gofumpt: $$($$(go env GOPATH)/bin/gofumpt -h >/dev/null 2>&1 && echo 'установлен' || echo 'не найден')$(RESET)\n"

git-hooks: ## 🔗 Установить Git хуки
	@printf "$(YELLOW)$(BOLD)🔗 Устанавливаю Git хуки...$(RESET)\n"
	@mkdir -p .git/hooks
	@echo '#!/bin/sh' > .git/hooks/pre-commit
	@echo 'echo "🔍 Запускаю pre-commit проверки..."' >> .git/hooks/pre-commit
	@echo 'echo "• Форматирование и проверка кода"' >> .git/hooks/pre-commit
	@echo 'make go-fmt go-vet go-lint' >> .git/hooks/pre-commit
	@echo 'if [ $$? -ne 0 ]; then' >> .git/hooks/pre-commit
	@echo '    echo "❌ Проверки кода не пройдены"' >> .git/hooks/pre-commit
	@echo '    exit 1' >> .git/hooks/pre-commit
	@echo 'fi' >> .git/hooks/pre-commit
	@echo 'echo "✅ Все проверки пройдены успешно!"' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@printf "$(GREEN)✅ Git хуки установлены$(RESET)\n"

# ================================================================================================
# 📊 МОНИТОРИНГ И СТАТИСТИКА
# ================================================================================================

.PHONY: metrics docker-stats

metrics: ## 📊 Показать метрики приложения
	@printf "$(CYAN)$(BOLD)📊 Метрики приложения:$(RESET)\n"
	@printf "$(YELLOW)LOC Go:$(RESET) $$(find bot -name "*.go" | xargs wc -l | tail -1 | awk '{print $$1}') строк\n"
	@printf "$(YELLOW)Кол-во Go файлов:$(RESET) $$(find bot -name "*.go" | wc -l)\n"
	@printf "$(YELLOW)Кол-во тестов:$(RESET) $$(grep -r "func Test" bot --include="*_test.go" | wc -l)\n"
	@printf "$(YELLOW)Кол-во пакетов:$(RESET) $$(cd bot && go list ./... | wc -l)\n"
	@printf "$(YELLOW)Кол-во зависимостей:$(RESET) $$(cd bot && go list -m all | wc -l)\n"
	@printf "$(YELLOW)Размер исходного кода:$(RESET) $$(du -sh bot | awk '{print $$1}')\n"
	@printf "$(YELLOW)Размер docker образа:$(RESET) $$(docker images | grep cu-clubs-bot | awk '{print $$7}')\n"

docker-stats: ## 📊 Показать статистику Docker
	@printf "$(CYAN)$(BOLD)📊 Статистика Docker:$(RESET)\n"
	@docker stats --no-stream

# ================================================================================================
# 🚀 ПРОДОВЫЙ ЗАПУСК
# ================================================================================================

.PHONY: prod-start prod-stop prod-restart prod-status prod-logs prod-pull prod-deploy

prod-pull: ## 📥 Подтянуть последние образы с Docker registry
	@printf "$(BLUE)$(BOLD)📥 Подтягиваю последние образы с Docker registry...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_PROD) pull

prod-start: prod-check ## 🚀 Запустить проект в продовом режиме
	@printf "$(GREEN)$(BOLD)🚀 Запускаю проект в продовом режиме...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_PROD) up -d

prod-stop: ## ⏹️ Остановить продовый режим
	@printf "$(RED)$(BOLD)⏹️ Останавливаю продовые контейнеры...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_PROD) down

prod-restart: prod-pull ## 🔄 Перезапустить все продовые контейнеры с обновлением образов
	@printf "$(YELLOW)$(BOLD)🔄 Перезапускаю все продовые контейнеры...$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_PROD) down
	@docker compose -f $(DOCKER_COMPOSE_PROD) up -d

prod-status: ## ℹ️ Показать статус продовых контейнеров
	@printf "$(CYAN)$(BOLD)ℹ️ Статус продовых контейнеров:$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_PROD) ps

prod-logs: ## 📃 Показать логи всех продовых контейнеров
	@printf "$(CYAN)$(BOLD)📃 Логи продовых контейнеров:$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_PROD) logs -f

prod-logs-bot: ## 📃 Показать логи бота в продовом режиме
	@printf "$(CYAN)$(BOLD)📃 Логи продового бота:$(RESET)\n"
	@docker compose -f $(DOCKER_COMPOSE_PROD) logs -f $(BOT_CONTAINER)

prod-deploy: prod-check prod-pull prod-restart ## 🚀 Полный деплой в продовый режим
	@printf "$(GREEN)$(BOLD)🚀 Деплой в прод выполнен успешно!$(RESET)\n"

prod-check: ## 🔍 Проверить готовность к продовому запуску
	@printf "$(CYAN)$(BOLD)🔍 Проверяю готовность к продовому запуску...$(RESET)\n"
	@printf "$(YELLOW)Проверка переменных окружения...$(RESET)\n"
	@[ -f .env ] || { printf "$(RED)❌ Файл .env не найден$(RESET)\n"; exit 1; }
	@printf "$(YELLOW)Проверка наличия docker-compose.yml...$(RESET)\n"
	@[ -f $(DOCKER_COMPOSE_PROD) ] || { printf "$(RED)❌ Файл $(DOCKER_COMPOSE_PROD) не найден$(RESET)\n"; exit 1; }
	@printf "$(YELLOW)Проверка наличия конфигурации...$(RESET)\n"
	@[ -f config/config.yaml ] || { printf "$(RED)❌ Файл config/config.yaml не найден$(RESET)\n"; exit 1; }
	@printf "$(GREEN)✅ Все проверки пройдены успешно! Можно запускать в продовом режиме.$(RESET)\n"

# ================================================================================================
# 🚀 БЫСТРЫЕ КОМАНДЫ И КОМБИНАЦИИ
# ================================================================================================

.PHONY: quick-start dev-reset

quick-start: build start ## 🚀 Быстрый старт (собрать и запустить в фоне)

dev-reset: clean build start ## 🔄 Полный сброс и перезапуск проекта
