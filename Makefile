# Makefile для protoc-gen-go
# Генератор Go протоколов из DSL

# ============================================================================
# Конфигурация
# ============================================================================

BINARY_NAME := protoc-gen-go
BUILD_DIR := bin
CMD_DIR := ./cmd/protoc-gen-go
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Параметры Go
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOCLEAN := $(GOCMD) clean
GOINSTALL := $(GOCMD) install
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

# ============================================================================
# Основные цели
# ============================================================================

.PHONY: all
all: clean deps fmt lint test build

.PHONY: help
help:
	@echo "protoc-gen-go - Генератор протоколов из DSL в Go"
	@echo ""
	@echo "Использование:"
	@echo "  make [цель]"
	@echo ""
	@echo "Доступные цели:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'
	@echo ""
	@echo "Примеры:"
	@echo "  make build          # Собрать бинарный файл"
	@echo "  make test           # Запустить тесты"
	@echo "  make demo           # Запустить демонстрацию"
	@echo "  make install        # Установить в GOPATH/bin"

.PHONY: deps
deps: ## Загрузить и упорядочить зависимости
	@echo "Загрузка зависимостей..."
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "Зависимости готовы"

.PHONY: fmt
fmt: ## Форматировать код Go
	@echo "Форматирование кода..."
	@$(GOFMT) ./cmd/... ./internal/... ./pkg/...
	@echo "Код отформатирован"

.PHONY: lint
lint: ## Запустить линтеры (vet)
	@echo "Запуск линтеров..."
	@$(GOVET) ./cmd/... ./internal/... ./pkg/...
	@echo "Линтинг пройден"

.PHONY: test
test: ## Запустить модульные тесты
	@echo "Запуск тестов..."
	@cd internal/dsl && $(GOTEST) -v
	@echo "Все тесты пройдены"

.PHONY: test-integration
test-integration: build ## Запустить интеграционные тесты
	@echo "Интеграционные тесты..."
	@./$(BUILD_DIR)/$(BINARY_NAME) testdata/full_test.dsl
	@mkdir -p testdata/protocol
	@mv testdata/full_test.gen.go testdata/protocol/ 2>/dev/null || true
	@cd testdata && $(GOTEST) -v
	@echo "Интеграционные тесты пройдены"

.PHONY: test-all
test-all: test test-integration ## Запустить все тесты

.PHONY: test-coverage
test-coverage: ## Запустить тесты с отчётом о покрытии
	@echo "Запуск тестов с покрытием..."
	@$(GOTEST) ./internal/... -coverprofile=coverage.out -covermode=atomic
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Отчёт о покрытии создан: coverage.html"

.PHONY: bench
bench: ## Запустить бенчмарки
	@echo "Запуск бенчмарков..."
	@$(GOTEST) ./internal/... -bench=. -benchmem -run=^$$
	@echo "Бенчмарки завершены"

.PHONY: build
build: ## Собрать бинарный файл
	@echo "Сборка $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Бинарный файл собран: $(BUILD_DIR)/$(BINARY_NAME) ($(VERSION))"

.PHONY: build-release
build-release: ## Собрать релизные версии для всех платформ
	@echo "Сборка релизных версий..."
	@mkdir -p $(BUILD_DIR)/release
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	@GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	@echo "Релизные версии собраны в $(BUILD_DIR)/release/"

.PHONY: clean
clean: ## Очистить артефакты сборки и сгенерированные файлы
	@echo "Очистка..."
	@rm -rf $(BUILD_DIR)/
	@rm -f coverage.out coverage.html
	@rm -f test.gen.go test.dsl test_complete.dsl test_minimal.dsl
	@rm -f demo/*.gen.go demo/protocol/*.gen.go 2>/dev/null || true
	@rm -f examples/*/*.gen.go 2>/dev/null || true
	@rm -f testdata/*.gen.go testdata/protocol/*.gen.go 2>/dev/null || true
	@find . -name "*.gen.go" -type f -delete 2>/dev/null || true
	@find . -name "*.test" -type f -delete 2>/dev/null || true
	@find . -name ".DS_Store" -type f -delete 2>/dev/null || true
	@find . -name "Thumbs.db" -type f -delete 2>/dev/null || true
	@find . -name "*~" -type f -delete 2>/dev/null || true
	@find . -name "*.swp" -type f -delete 2>/dev/null || true
	@rm -rf testdata/protocol demo/protocol 2>/dev/null || true
	@$(GOCLEAN) -cache -testcache
	@echo "Очистка завершена"

.PHONY: distclean
distclean: clean ## Глубокая очистка, включая кэш модулей
	@echo "Глубокая очистка кэша модулей..."
	@$(GOCLEAN) -modcache
	@echo "Глубокая очистка завершена"

.PHONY: install
install: build ## Установить бинарный файл в GOPATH/bin
	@echo "Установка $(BINARY_NAME)..."
	@$(GOINSTALL) $(LDFLAGS) $(CMD_DIR)
	@echo "Установлено в GOPATH/bin"

.PHONY: demo
demo: ## Запустить демонстрацию возможностей
	@echo "================================================"
	@echo "         ДЕМОНСТРАЦИЯ PROTOCOL-GEN-GO           "
	@echo "================================================"
	@echo ""
	@echo "Запуск демо..."
	@cd demo && $(GOCMD) run run.go
	@echo ""
	@echo "================================================"

.PHONY: examples
examples: build ## Сгенерировать все примеры
	@echo "Генерация примеров..."
	@for dsl in examples/*/*.dsl; do \
		echo "  Генерация $$dsl..."; \
		./$(BUILD_DIR)/$(BINARY_NAME) "$$dsl" || exit 1; \
	done
	@echo "Все примеры сгенерированы"

.PHONY: dev
dev: clean deps build test ## Полная пересборка для разработки

.PHONY: quick
quick: build demo ## Быстрая сборка и запуск демо

.PHONY: check
check: fmt lint test ## Запустить все проверки

.PHONY: version
version: ## Показать версию
	@echo "$(VERSION)"

.PHONY: tag
tag: ## Создать и отправить git тег (использование: make tag VERSION=v1.0.0)
	@[ -n "$(V)" ] || { echo "Использование: make tag V=v1.0.0"; exit 1; }
	@git tag -a $(V) -m "Релиз $(V)"
	@git push origin $(V)
	@echo "Тег $(V) создан и отправлен"

.PHONY: release
release: check build-release ## Создать релиз (проверки и сборка под все платформы)
	@echo "Релиз $(VERSION) готов в $(BUILD_DIR)/release/"
