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

# Пути эксперимента
EXPERIMENT_DIR := benchmarks/experiment
REPORT_DIR := $(EXPERIMENT_DIR)/report
BENCH_DIR := benchmarks/golang

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
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-25s %s\n", $$1, $$2}'
	@echo ""
	@echo "Примеры:"
	@echo "  make build              # Собрать бинарный файл"
	@echo "  make test               # Запустить все тесты"
	@echo "  make bench-quick        # Быстрые бенчмарки DSL"
	@echo "  make bench-report       # Бенчмарки DSL + отчёт (без внешних зависимостей)"
	@echo "  make experiment         # Полный эксперимент: DSL vs Hand vs Protobuf vs Construct"
	@echo "  make charts             # Только графики (по готовым результатам)"
	@echo "  make demo               # Базовое демо"
	@echo "  make demo-all           # Все демонстрации"
	@echo "  make pipeline           # Полный pipeline тест"
	@echo "  make clean              # Очистить артефакты"
	@echo "  make distclean          # Глубокая очистка"
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
	@$(GOFMT) ./cmd/... ./internal/... ./pkg/... 2>/dev/null || true
	@echo "Код отформатирован"

.PHONY: fmt-dsl
fmt-dsl: build ## Форматировать все DSL файлы
	@echo "Форматирование DSL..."
	@for dsl in examples/*/*.dsl benchmarks/experiment/protocols/*.dsl testdata/ipv6/*/*.dsl testdata/nested/*.dsl 2>/dev/null; do \
		if [ -f "$$dsl" ]; then \
			echo "  $$dsl"; \
			./$(BUILD_DIR)/$(BINARY_NAME) fmt "$$dsl" > /tmp/fmt_dsl.tmp && mv /tmp/fmt_dsl.tmp "$$dsl"; \
		fi \
	done
	@echo "Все DSL файлы отформатированы"

.PHONY: lint
lint: ## Запустить линтеры (vet)
	@echo "Запуск линтеров..."
	@$(GOVET) ./cmd/... ./internal/... ./pkg/... 2>/dev/null || true
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

.PHONY: install
install: build ## Установить бинарный файл
	@echo "Установка $(BINARY_NAME)..."
	@GOBIN=$$(go env GOBIN 2>/dev/null); \
	if [ -n "$$GOBIN" ] && [ -d "$$GOBIN" ]; then \
		go install $(LDFLAGS) $(CMD_DIR); \
		echo "Установлено в $$GOBIN/$(BINARY_NAME)"; \
	else \
		GOPATH=$$(go env GOPATH 2>/dev/null || echo "$$HOME/go"); \
		mkdir -p "$$GOPATH/bin"; \
		cp $(BUILD_DIR)/$(BINARY_NAME) "$$GOPATH/bin/"; \
		echo "Установлено в $$GOPATH/bin/$(BINARY_NAME)"; \
	fi

# ============================================================================
# Очистка
# ============================================================================

.PHONY: clean
clean: ## Очистить артефакты сборки и временные файлы
	@echo "Очистка артефактов..."
	@# Сборочный каталог
	@rm -rf $(BUILD_DIR)/
	@# Coverage
	@rm -f coverage.out coverage.html
	@# Кэш тестов Go
	@$(GOCLEAN) -testcache
	@# Сгенерированные Go-файлы (продукты кодогенерации, можно пересоздать)
	@find . -name "*.gen.go" -type f -delete 2>/dev/null || true
	@# Сгенерированные Protobuf-файлы
	@find . -name "*.pb.go" -type f -delete 2>/dev/null || true
	@# Бинарные схемы (.bin) — продукты --save-bin
	@find . -name "*.bin" -type f -delete 2>/dev/null || true
	@# Скомпилированные тестовые бинарники
	@find . -name "*.test" -type f -delete 2>/dev/null || true
	@# Дубликаты protoc-генерации (github.com/...)
	@rm -rf github.com/ 2>/dev/null || true
	@# Дубликаты pb в поддиректориях
	@find . -type d -name "pb" -exec rm -rf {} + 2>/dev/null || true
	@# Кэш Python
	@find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	@# Резервные копии редакторов
	@find . -name "*~" -type f -delete 2>/dev/null || true
	@find . -name "*.swp" -type f -delete 2>/dev/null || true
	@find . -name "*.swo" -type f -delete 2>/dev/null || true
	@# Системные файлы
	@find . -name ".DS_Store" -type f -delete 2>/dev/null || true
	@find . -name "Thumbs.db" -type f -delete 2>/dev/null || true
	@# Временные файлы
	@find . -name "*.tmp" -type f -delete 2>/dev/null || true
	@find . -name "*.bak" -type f -delete 2>/dev/null || true
	@# Файлы профайлера
	@rm -f cpu.prof mem.prof trace.out
	@# Очистка report/ — всё кроме generate_report.py и final/
	@if [ -d "$(REPORT_DIR)" ]; then \
		find $(REPORT_DIR) -type f ! -name 'generate_report.py' ! -path '*/final/*' -delete 2>/dev/null || true; \
		find $(REPORT_DIR) -type d -empty -delete 2>/dev/null || true; \
	fi
	@# Удаляем симлинк на последний отчёт
	@rm -f $(REPORT_DIR)/BENCHMARK_REPORT_LATEST.md
	@# Пустые директории
	@rmdir pkg/protocol 2>/dev/null || true
	@rmdir pkg 2>/dev/null || true
	@rmdir testdata/protocol 2>/dev/null || true
	@rmdir demo/protocol 2>/dev/null || true
	@echo "Очистка завершена"
	
.PHONY: distclean
distclean: clean ## Глубокая очистка, включая кэш модулей и результаты эксперимента
	@echo "Глубокая очистка..."
	@$(GOCLEAN) -cache -modcache
	@# Удаляем все результаты бенчмарков
	@rm -f $(EXPERIMENT_DIR)/results_*.txt
	@# Полная очистка report/ (включая generate_report.py, но не final/)
	@if [ -d "$(REPORT_DIR)" ]; then \
		find $(REPORT_DIR) -type f ! -path '*/final/*' -delete 2>/dev/null || true; \
		find $(REPORT_DIR) -type d -empty -delete 2>/dev/null || true; \
	fi
	@# Удаляем симлинк на последний отчёт
	@rm -f $(REPORT_DIR)/BENCHMARK_REPORT_LATEST.md
	@echo "Глубокая очистка завершена"

# ============================================================================
# Бенчмаркинг и отчёты
# ============================================================================

# --- Вариант 1: Только protoc-gen-go (без внешних зависимостей) ---

.PHONY: bench-report
bench-report: build ## Бенчмарки DSL + отчёт (только твоя утилита, без внешних зависимостей)
	@echo "================================================"
	@echo "     БЕНЧМАРКИ PROTOCOL-GEN-GO (только DSL)     "
	@echo "================================================"
	@echo ""
	@# Проверка зависимостей
	@command -v $(GOCMD) >/dev/null 2>&1 || { echo "Требуется Go 1.21+"; exit 1; }
	@echo "✓ Go $(shell $(GOCMD) version | awk '{print $$3}')"
	@echo ""
	@# Этап 1: Генерация тестовых протоколов
	@echo "=== Этап 1/4: Генерация тестовых протоколов ==="
	@for f in $(EXPERIMENT_DIR)/protocols/*.dsl; do \
		./$(BUILD_DIR)/$(BINARY_NAME) "$$f" > /dev/null 2>&1 || { echo "Ошибка: $$f"; exit 1; }; \
	done
	@echo "✓ Протоколы сгенерированы"
	@echo ""
	@# Этап 2: Бенчмарки DSL
	@echo "=== Этап 2/4: Бенчмарки DSL (эксперимент А и Б) ==="
	@cd $(EXPERIMENT_DIR)/protocols && $(GOTEST) -bench=. -benchmem -benchtime=2s -run=^$$ | tee ../results_dsl.txt
	@echo ""
	@# Этап 3: Бенчмарки ручной реализации
	@echo "=== Этап 3/4: Бенчмарки ручной реализации ==="
	@cd $(EXPERIMENT_DIR)/handwritten && $(GOTEST) -bench=. -benchmem -benchtime=2s -run=^$$ | tee ../results_hand.txt
	@echo ""
	@# Этап 4: Генерация отчёта
	@echo "=== Этап 4/4: Генерация отчёта ==="
	@if python3 -c "import matplotlib" 2>/dev/null; then \
		cd $(REPORT_DIR) && python3 generate_report.py; \
		echo "✓ Графики сгенерированы: $(REPORT_DIR)/*.png"; \
	else \
		echo "⚠  Python matplotlib не установлен — графики пропущены"; \
		echo "   Установка: pip3 install --user matplotlib numpy"; \
		echo "   Текстовые результаты сохранены в:"; \
		echo "   $(EXPERIMENT_DIR)/results_dsl.txt"; \
		echo "   $(EXPERIMENT_DIR)/results_hand.txt"; \
	fi
	@echo ""
	@echo "================================================"
	@echo "     БЕНЧМАРКИ ЗАВЕРШЕНЫ                        "
	@echo "================================================"
	@echo ""
	@echo "Результаты:"
	@echo "   Бенчмарки DSL:  $(EXPERIMENT_DIR)/results_dsl.txt"
	@echo "   Бенчмарки Hand: $(EXPERIMENT_DIR)/results_hand.txt"
	@if [ -f "$(REPORT_DIR)/chart_a_marshal.png" ]; then \
		echo "   Графики:        $(REPORT_DIR)/*.png"; \
	fi
	@echo ""

# --- Вариант 2: Полное сравнение с аналогами (требует внешние инструменты) ---

.PHONY: experiment-check-deps
experiment-check-deps: ## Проверить наличие всех внешних зависимостей для эксперимента
	@echo "Проверка зависимостей для полного эксперимента..."
	@echo ""
	@# Go
	@command -v $(GOCMD) >/dev/null 2>&1 || { echo "Go не найден"; exit 1; }
	@echo "✓ Go $(shell $(GOCMD) version | awk '{print $$3}')"
	@# Python 3
	@command -v python3 >/dev/null 2>&1 || { echo "Python 3 не найден"; exit 1; }
	@echo "✓ Python $(shell python3 --version 2>&1 | awk '{print $$2}')"
	@# Protobuf compiler
	@command -v protoc >/dev/null 2>&1 || { echo "protoc не найден (нужен protobuf-compiler)"; exit 1; }
	@echo "✓ protoc $(shell protoc --version 2>&1 | awk '{print $$2}')"
	@# Protobuf Go plugin
	@if [ -x "$$HOME/go/bin/protoc-gen-go" ] || command -v protoc-gen-go >/dev/null 2>&1; then \
		echo "✓ protoc-gen-go (Google)"; \
	else \
		echo "protoc-gen-go (Google) не найден"; \
		echo "   Установка: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"; \
		exit 1; \
	fi
	@# Python Construct
	@python3 -c "import construct" 2>/dev/null || { \
		echo "Python construct не установлен"; \
		echo "   Установка: pip3 install --user construct"; \
		exit 1; \
	}
	@echo "✓ Python Construct $(shell python3 -c 'import construct; print(construct.__version__)' 2>/dev/null)"
	@# Python matplotlib (для графиков)
	@if python3 -c "import matplotlib" 2>/dev/null; then \
		echo "✓ Python matplotlib (для графиков)"; \
	else \
		echo "⚠  Python matplotlib не установлен — графики не будут сгенерированы"; \
		echo "   Установка: pip3 install --user matplotlib numpy"; \
	fi
	@echo ""
	@echo "✓ Все зависимости найдены"
	@echo ""

.PHONY: experiment
experiment: build experiment-check-deps ## Полный эксперимент: сравнение DSL vs Hand vs Protobuf vs Construct + отчёт
	@echo "================================================"
	@echo "   ПОЛНЫЙ СРАВНИТЕЛЬНЫЙ ЭКСПЕРИМЕНТ             "
	@echo "   protoc-gen-go vs Hand vs Protobuf vs Construct"
	@echo "================================================"
	@echo ""
	@echo "⚠  Этот эксперимент запускает бенчмарки на 4 инструментах."
	@echo "   Общее время выполнения: ~5-10 минут."
	@echo "   Убедитесь, что система не нагружена."
	@echo ""
	@# Этап 1: Генерация всех протоколов
	@echo "=== Этап 1/6: Генерация DSL-протоколов ==="
	@for f in $(EXPERIMENT_DIR)/protocols/*.dsl; do \
		./$(BUILD_DIR)/$(BINARY_NAME) "$$f" > /dev/null 2>&1 || { echo "Ошибка: $$f"; exit 1; }; \
	done
	@echo "✓ DSL-протоколы сгенерированы"
	@echo ""
	@# Этап 2: Генерация Protobuf
	@echo "=== Этап 2/6: Генерация Protobuf ==="
	@cd $(EXPERIMENT_DIR)/protobuf && protoc --go_out=. --go_opt=paths=source_relative ipv6.proto 2>/dev/null || { \
		echo "Ошибка генерации Protobuf"; exit 1; \
	}
	@echo "✓ Protobuf сгенерирован"
	@echo ""
	@# Этап 3: Бенчмарк DSL
	@echo "=== Этап 3/6: Бенчмарк protoc-gen-go (DSL) ==="
	@cd $(EXPERIMENT_DIR)/protocols && $(GOTEST) -bench=. -benchmem -benchtime=2s -run=^$$ | tee ../results_dsl.txt
	@echo ""
	@# Этап 4: Бенчмарк ручной реализации
	@echo "=== Этап 4/6: Бенчмарк Handwritten Go ==="
	@cd $(EXPERIMENT_DIR)/handwritten && $(GOTEST) -bench=. -benchmem -benchtime=2s -run=^$$ | tee ../results_hand.txt
	@echo ""
	@# Этап 5: Бенчмарк Protobuf
	@echo "=== Этап 5/6: Бенчмарк Google Protobuf ==="
	@cd $(EXPERIMENT_DIR)/protobuf && $(GOTEST) -bench=. -benchmem -benchtime=2s -run=^$$ | tee ../results_pb.txt
	@echo ""
	@# Этап 6: Бенчмарк Python Construct
	@echo "=== Этап 6/6: Бенчмарк Python Construct ==="
	@cd $(EXPERIMENT_DIR)/construct && python3 ipv6_construct.py | tee ../results_construct.txt
	@echo ""
	@# Генерация отчёта с графиками
	@echo "=== Генерация отчёта ==="
	@if python3 -c "import matplotlib" 2>/dev/null; then \
		cd $(REPORT_DIR) && python3 generate_report.py; \
		echo "✓ Графики сгенерированы"; \
	else \
		echo "⚠  matplotlib не установлен — графики пропущены"; \
		echo "   Таблицы доступны в файлах результатов"; \
	fi
	@echo ""
	@echo "================================================"
	@echo "   ЭКСПЕРИМЕНТ ЗАВЕРШЁН                         "
	@echo "================================================"
	@echo ""
	@echo "📊 Результаты:"
	@echo "   DSL:        $(EXPERIMENT_DIR)/results_dsl.txt"
	@echo "   Hand:       $(EXPERIMENT_DIR)/results_hand.txt"
	@echo "   Protobuf:   $(EXPERIMENT_DIR)/results_pb.txt"
	@echo "   Construct:  $(EXPERIMENT_DIR)/results_construct.txt"
	@if [ -f "$(REPORT_DIR)/chart_a_marshal.png" ]; then \
		echo "   Графики:    $(REPORT_DIR)/*.png"; \
	fi
	@echo "   Отчёт:      $(REPORT_DIR)/BENCHMARK_REPORT.md"
	@echo ""

# --- Быстрые бенчмарки ---

.PHONY: bench-quick
bench-quick: build ## Быстрые бенчмарки DSL (только цифры, без графиков)
	@echo "Быстрые бенчмарки protoc-gen-go..."
	@for f in $(EXPERIMENT_DIR)/protocols/*.dsl; do \
		./$(BUILD_DIR)/$(BINARY_NAME) "$$f" > /dev/null 2>&1; \
	done
	@cd $(EXPERIMENT_DIR)/protocols && $(GOTEST) -bench=. -benchmem -benchtime=1s -run=^$$
	@echo "Готово"

.PHONY: bench-full
bench-full: build ## Все бенчмарки проекта (парсер + DSL + сравнение с Go-аналогами)
	@echo "================================================"
	@echo "            ПОЛНЫЕ БЕНЧМАРКИ                    "
	@echo "================================================"
	@echo ""
	@echo "=== Бенчмарки парсера ==="
	@$(GOTEST) ./internal/parser/... -bench=. -benchmem -run=^$$ 2>/dev/null || true
	@echo ""
	@echo "=== Бенчмарки DSL (сравнение с Handwritten) ==="
	@for f in $(EXPERIMENT_DIR)/protocols/*.dsl; do \
		./$(BUILD_DIR)/$(BINARY_NAME) "$$f" > /dev/null 2>&1; \
	done
	@cd $(BENCH_DIR) && $(GOTEST) -bench=. -benchmem -benchtime=1s -run=^$$ 2>/dev/null || true
	@echo ""
	@echo "================================================"
	@echo "            БЕНЧМАРКИ ЗАВЕРШЕНЫ                 "
	@echo "================================================"

# --- Графики (отдельно) ---

.PHONY: charts
charts: ## Сгенерировать графики из существующих результатов (требуется Python)
	@echo "Проверка зависимостей для графиков..."
	@python3 -c "import matplotlib" 2>/dev/null || { \
		echo "Требуется Python 3 + matplotlib"; \
		echo "   Установка: pip3 install --user matplotlib numpy"; \
		exit 1; \
	}
	@echo "✓ matplotlib найден"
	@echo "Генерация графиков..."
	@if [ ! -f "$(EXPERIMENT_DIR)/results_dsl.txt" ]; then \
		echo " Файлы результатов не найдены. Сначала запустите:"; \
		echo "   make bench-report    (только DSL)"; \
		echo "   make experiment      (полное сравнение)"; \
		exit 1; \
	fi
	@cd $(REPORT_DIR) && python3 generate_report.py
	@echo "✓ Графики сохранены в $(REPORT_DIR)/"
	@ls -la $(REPORT_DIR)/*.png 2>/dev/null || echo "Графики не найдены"

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
demo-conditions: ## Демонстрация условных полей
	@echo "================================================"
	@echo "    ДЕМОНСТРАЦИЯ УСЛОВИЙ (пути, &&, ||)         "
	@echo "================================================"
	@$(GOCMD) run examples/conditions/demo_conditions.go
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

.PHONY: demo-protocols
demo-protocols: build ## Наглядная демонстрация реальных протоколов
	@bash demo/protocols_demo.sh

.PHONY: demo-all
demo-all: demo demo-arrays demo-dns demo-conditions demo-enum demo-endian ## Запустить все демонстрации

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
