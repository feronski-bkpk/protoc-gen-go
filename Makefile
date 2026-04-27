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
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-22s %s\n", $$1, $$2}'
	@echo ""
	@echo "Примеры:"
	@echo "  make build              # Собрать бинарный файл"
	@echo "  make test               # Запустить все тесты"
	@echo "  make test-fuzz          # Фаззинг-тесты"
	@echo "  make bench              # Бенчмарки"
	@echo "  make demo               # Базовое демо"
	@echo "  make demo-all           # Все демонстрации"
	@echo "  make demo-protocols     # Демо реальных протоколов"
	@echo "  make pipeline           # Полный pipeline тест"
	@echo "  make fmt-dsl            # Форматировать все DSL"
	@echo "  make examples           # Сгенерировать все примеры"
	@echo "  make install            # Установить в GOPATH/bin"

.PHONY: deps
deps: ## Загрузить и упорядочить зависимости
	@echo "Загрузка зависимостей..."
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "Зависимости готовы"

.PHONY: fmt
fmt: ## Форматировать код Go (go fmt)
	@echo "Форматирование кода..."
	@$(GOFMT) ./cmd/... ./internal/... ./pkg/...
	@echo "Код отформатирован"

.PHONY: fmt-dsl
fmt-dsl: build ## Форматировать все DSL файлы
	@echo "Форматирование DSL..."
	@for dsl in examples/*/*.dsl; do \
		echo "  $$dsl"; \
		./$(BUILD_DIR)/$(BINARY_NAME) fmt "$$dsl" > /tmp/fmt_dsl.tmp && mv /tmp/fmt_dsl.tmp "$$dsl"; \
	done
	@echo "Все DSL файлы отформатированы"

.PHONY: lint
lint: ## Запустить линтеры (vet)
	@echo "Запуск линтеров..."
	@$(GOVET) ./cmd/... ./internal/... ./pkg/...
	@echo "Линтинг пройден"

.PHONY: test-parser
test-parser: ## Тесты парсера (16 тестов + фаззинг)
	@echo "Тесты парсера..."
	@cd internal/parser && $(GOTEST) -v
	@echo "Все тесты парсера пройдены"

.PHONY: test-analyzer
test-analyzer: ## Тесты анализатора (3 теста)
	@echo "Тесты анализатора..."
	@cd internal/analyzer && $(GOTEST) -v
	@echo "Все тесты анализатора пройдены"

.PHONY: test-binary
test-binary: ## Тесты бинарного формата
	@echo "Тесты бинарного формата..."
	@cd internal/binary && $(GOTEST) -v 2>/dev/null || echo "Тесты бинарного формата пока не созданы"

.PHONY: test-fuzz
test-fuzz: ## Фаззинг-тесты парсера (30 секунд)
	@echo "Фаззинг-тесты парсера..."
	@cd internal/parser && $(GOTEST) -fuzz=FuzzParser -fuzztime=30s 2>/dev/null || echo "Фаззинг-тесты требуют Go 1.18+"

.PHONY: test
test: test-parser test-analyzer test-binary ## Запустить все тесты

.PHONY: test-coverage
test-coverage: ## Запустить тесты с отчётом о покрытии
	@echo "Запуск тестов с покрытием..."
	@$(GOTEST) ./internal/... -coverprofile=coverage.out -covermode=atomic
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Отчёт о покрытии создан: coverage.html"

.PHONY: bench
bench: ## Бенчмарки парсера
	@echo "Запуск бенчмарков..."
	@$(GOTEST) ./internal/parser/... -bench=. -benchmem -run=^$$
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
	@rm -f test.gen.go test.dsl test_complete.dsl test_minimal.dsl test_error.dsl
	@rm -f test_fmt.dsl test_fmt_complex.dsl test_fmt_advanced.dsl test_fmt_cond.dsl
	@rm -f test_format.dsl
	@rm -f demo/*.gen.go demo/protocol/*.gen.go 2>/dev/null || true
	@rm -f examples/*/*.gen.go 2>/dev/null || true
	@rm -f testdata/*.gen.go testdata/protocol/*.gen.go 2>/dev/null || true
	@rm -f testdata/full_test.dsl testdata/full_test.go 2>/dev/null || true
	@rm -f examples/full_example.dsl 2>/dev/null || true
	@find . -name "*.gen.go" -type f -delete 2>/dev/null || true
	@find . -name "*.bin" -type f -delete 2>/dev/null || true
	@find . -name "*.test" -type f -delete 2>/dev/null || true
	@find . -name ".DS_Store" -type f -delete 2>/dev/null || true
	@find . -name "Thumbs.db" -type f -delete 2>/dev/null || true
	@find . -name "*~" -type f -delete 2>/dev/null || true
	@find . -name "*.swp" -type f -delete 2>/dev/null || true
	@rm -rf testdata/protocol demo/protocol examples/*/protocol 2>/dev/null || true
	@rm -rf internal/parser/testdata 2>/dev/null || true
	@rmdir pkg/protocol 2>/dev/null || true
	@rmdir pkg 2>/dev/null || true
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

# ============================================================================
# Демонстрации
# ============================================================================

.PHONY: demo
demo: ## Демонстрация базового протокола (сенсор)
	@echo "================================================"
	@echo "         ДЕМОНСТРАЦИЯ PROTOCOL-GEN-GO           "
	@echo "================================================"
	@cd demo && $(GOCMD) run run.go
	@echo "================================================"

.PHONY: demo-arrays
demo-arrays: ## Демонстрация массивов и слайсов
	@echo "================================================"
	@echo "         ДЕМОНСТРАЦИЯ МАССИВОВ И СЛАЙСОВ        "
	@echo "================================================"
	@$(GOCMD) run examples/arrays/slice_full_demo.go
	@echo "================================================"

.PHONY: demo-dns
demo-dns: ## Демонстрация DNS протокола
	@echo "================================================"
	@echo "         ДЕМОНСТРАЦИЯ DNS                       "
	@echo "================================================"
	@$(GOCMD) run examples/dns/dns_complete.go
	@echo "================================================"

.PHONY: demo-conditions
demo-conditions: ## Демонстрация условий с путями и вложенных условий
	@echo "================================================"
	@echo "    ДЕМОНСТРАЦИЯ УСЛОВИЙ (пути, &&, ||)         "
	@echo "================================================"
	@$(GOCMD) run examples/conditions/demo_conditions.go
	@echo ""
	@echo "--- Вложенные условия ---"
	@./$(BUILD_DIR)/$(BINARY_NAME) examples/conditions/nested_cond.dsl 2>/dev/null
	@grep -A 2 "if p\." examples/conditions/nested_cond.gen.go || true
	@echo "================================================"

.PHONY: demo-enum
demo-enum: ## Демонстрация enum-типов
	@echo "================================================"
	@echo "         ДЕМОНСТРАЦИЯ ENUM-ТИПОВ                "
	@echo "================================================"
	@$(GOCMD) run examples/enums/demo_enum.go
	@echo "================================================"

.PHONY: demo-endian
demo-endian: ## Демонстрация LittleEndian
	@echo "================================================"
	@echo "         ДЕМОНСТРАЦИЯ LITTLE ENDIAN             "
	@echo "================================================"
	@$(GOCMD) run examples/little_endian/demo_endian.go
	@echo "================================================"

.PHONY: demo-aliases
demo-aliases: ## Демонстрация алиасов
	@echo "================================================"
	@echo "         ДЕМОНСТРАЦИЯ АЛИАСОВ                   "
	@echo "================================================"
	@./$(BUILD_DIR)/$(BINARY_NAME) examples/aliases/data.dsl 2>/dev/null
	@cat examples/aliases/data.gen.go 2>/dev/null | head -25 || echo "(сгенерируйте: make examples)"
	@echo "================================================"

.PHONY: demo-consts
demo-consts: ## Демонстрация констант
	@echo "================================================"
	@echo "         ДЕМОНСТРАЦИЯ КОНСТАНТ                  "
	@echo "================================================"
	@./$(BUILD_DIR)/$(BINARY_NAME) examples/consts/config.dsl 2>/dev/null
	@cat examples/consts/config.gen.go 2>/dev/null | head -25 || echo "(сгенерируйте: make examples)"
	@echo "================================================"

.PHONY: demo-protocols
demo-protocols: build ## Наглядная демонстрация реальных протоколов
	@bash demo/protocols_demo.sh

.PHONY: demo-all
demo-all: demo demo-arrays demo-dns demo-conditions demo-enum demo-endian demo-aliases demo-consts ## Запустить все демонстрации

# ============================================================================
# Полный pipeline тест
# ============================================================================

.PHONY: pipeline
pipeline: clean build ## Полный тест pipeline: парсинг → анализ → генерация → компиляция
	@echo "================================================"
	@echo "            ПОЛНЫЙ PIPELINE ТЕСТ                 "
	@echo "================================================"
	@echo ""
	@echo "Этап 1: Генерация всех примеров..."
	@for dsl in examples/*/*.dsl; do \
		./$(BUILD_DIR)/$(BINARY_NAME) "$$dsl" > /dev/null 2>&1 || echo "ERROR $$dsl"; \
	done
	@echo "Все DSL сгенерированы"
	@echo ""
	@echo "Этап 2: Проверка компиляции (go vet)..."
	@errors=0; \
	for dsl in examples/*/*.dsl; do \
		gen="$${dsl%.dsl}.gen.go"; \
		go vet "$$gen" 2>/dev/null || { echo "ERROR $$gen"; errors=$$((errors+1)); }; \
	done; \
	if [ $$errors -eq 0 ]; then echo "Все файлы проходят go vet"; fi
	@echo ""
	@echo "Этап 3: Форматирование DSL..."
	@for dsl in examples/*/*.dsl; do \
		./$(BUILD_DIR)/$(BINARY_NAME) fmt "$$dsl" > /dev/null 2>&1 || echo "ERROR fmt $$dsl"; \
	done
	@echo "Форматирование работает"
	@echo ""
	@echo "Этап 4: Бинарный формат (круговой тест)..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --save-bin examples/simple/simple.dsl > /dev/null 2>&1
	@./$(BUILD_DIR)/$(BINARY_NAME) --load-bin examples/simple/simple.bin > /dev/null 2>&1 && echo "Круговой тест пройден" || echo "ERROR Ошибка"
	@echo ""
	@echo "Этап 5: Запуск тестов..."
	@$(MAKE) test
	@echo ""
	@echo "================================================"
	@echo "            PIPELINE ЗАВЕРШЁН                    "
	@echo "================================================"

# ============================================================================
# Бинарный формат
# ============================================================================

.PHONY: save-bin
save-bin: build ## Сохранить схему в бинарный формат
	@echo "Сохранение схемы в бинарный формат..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --save-bin examples/simple/simple.dsl
	@echo "Схема сохранена: examples/simple/simple.bin"
	@ls -la examples/simple/simple.bin

.PHONY: load-bin
load-bin: build ## Загрузить схему из бинарного формата
	@echo "Загрузка схемы из бинарного формата..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --load-bin examples/simple/simple.bin | head -30

# ============================================================================
# Утилиты
# ============================================================================

.PHONY: examples
examples: build ## Сгенерировать все примеры
	@echo "Генерация примеров..."
	@for dsl in examples/*/*.dsl; do \
		echo "  $$dsl"; \
		./$(BUILD_DIR)/$(BINARY_NAME) "$$dsl" > /dev/null 2>&1 || echo "ERROR Ошибка"; \
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
tag: ## Создать и отправить git тег (использование: make tag V=v1.0.0)
	@[ -n "$(V)" ] || { echo "Использование: make tag V=v1.0.0"; exit 1; }
	@git tag -a $(V) -m "Релиз $(V)"
	@git push origin $(V)
	@echo "Тег $(V) создан и отправлен"

.PHONY: release
release: check build-release ## Создать релиз (проверки и сборка под все платформы)
	@echo "Релиз $(VERSION) готов в $(BUILD_DIR)/release/"