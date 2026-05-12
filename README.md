# protoc-gen-go

Генератор Go-кода для бинарных протоколов из человеко-читаемого DSL.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Назначение

Инструмент позволяет:

- описать бинарный протокол в декларативном текстовом формате (DSL)
- автоматически получить строго типизированные Go-структуры
- сгенерировать методы `MarshalBinary` / `UnmarshalBinary` / `Size` / `Validate`
- получить константы смещений для отладки и прямого доступа к полям
- сохранить схему протокола в компактный бинарный формат (`.bin`)
- форматировать DSL файлы в едином стиле

**Не требует внешних зависимостей** — только стандартная библиотека Go.

## Принцип работы

```
DSL (.dsl) → Собственный парсер → AST → Анализатор → Генератор → Go-код (.gen.go)
                  ↕                        ↓
           Бинарный формат (.bin)      Форматтер (fmt)
```

1. Пользователь описывает протокол в файле `.dsl`
2. Собственный парсер (Recursive Descent) строит AST
3. Анализатор проверяет семантику и строит таблицу символов
4. Генератор создаёт Go-файл со структурами и методами сериализации
5. Схему можно сохранить в `.bin` и восстановить обратно
6. DSL можно отформатировать через `protoc-gen-go fmt`

Порядок байт по умолчанию — **Big Endian** (сетевой порядок). Поддерживается **LittleEndian** через `endian: little`.

## Производительность

Бенчмарки проводились на Intel Core i3-7020U @ 2.30GHz, Go 1.23. Сравнение с индустриальными аналогами:

### SensorData (простой протокол: uint32 + float32 × 3 + массив)

| Инструмент | Marshal (ns/op) | Unmarshal (ns/op) | Аллокаций |
|-----------|----------------|-------------------|-----------|
| **protoc-gen-go** | **74** | **26** | 0–1 |
| Google Protobuf | 256 | 381 | 1–2 |
| Python Construct | 26 500 | 27 700 | N/A |

### IPv4 Header (битовые поля, 20 байт)

| Инструмент | Marshal (ns/op) | Unmarshal (ns/op) | Аллокаций |
|-----------|----------------|-------------------|-----------|
| **protoc-gen-go** | **46** | **12** | 0–1 |
| Google Protobuf | 319 | 327 | 1 |

### IPv6 с цепочкой расширений (битовые поля, условные структуры, переменная длина)

| Инструмент | Marshal (ns/op) | Unmarshal (ns/op) | Аллокаций |
|-----------|----------------|-------------------|-----------|
| **protoc-gen-go** | **1 378** | **380** | 0–1 |
| Ручная реализация Go | 1 276 | 389 | 0–1 |
| Google Protobuf | 4 144 | 2 880 | 2–21 |
| Python Construct | 149 646 | 148 462 | N/A |

**protoc-gen-go быстрее Protobuf в 3–12 раз** и **быстрее Python Construct в 100–750 раз** в зависимости от размера и сложности протокола.

Подробные результаты, графики и методология — в [отчёте о сравнительном тестировании](benchmarks/experiment/report/final/BENCHMARK_REPORT.md) и [автоматически сгенерированных отчётах](benchmarks/experiment/report/).

### Что измерялось

- **ns/op** — наносекунд на одну операцию (меньше = быстрее)
- **Аллокаций** — количество выделений памяти в куче (меньше = меньше нагрузки на GC)
- **Marshal** — превращение структуры в байтовый массив (отправка/сохранение)
- **Unmarshal** — чтение структуры из байтового массива (приём/загрузка)

### Преимущества protoc-gen-go

- **Специализация на бинарных протоколах**: полный контроль над каждым битом и байтом
- **Максимальная производительность**: в 3–12× быстрее Protobuf
- **Нет внешних зависимостей**: только стандартная библиотека Go
- **Условные поля**: поддержка `&&`, `||`, путей к битам, сравнений с enum
- **Компактный DSL**: описание протокола в 2 раза короче `.proto`-файла

### Ограничения

- Один целевой язык (Go)
- Фиксированная структура полей — нет обратной совместимости при изменении схемы (для эволюционирующих API лучше Protobuf)
- Одна аллокация при Marshal (буфер создаётся заново)

## Установка

```bash
git clone https://github.com/feronski-bkpk/protoc-gen-go.git
cd protoc-gen-go
make build
make install
```

## Быстрый старт

```bash
cat > sensor.dsl << 'EOF'
protocol SensorData {
    id: 0xABCD
    device_id: uint32
    temperature: float32
    flags: bitstruct {
        ack: bit(7)
        error: bit(6)
        priority: bits[5:4]
    }
    readings_len: uint16
    readings: []float32 length: readings_len
    name_len: uint16
    name: bytes length_from: name_len
}
EOF

protoc-gen-go sensor.dsl
```

## Использование

### CLI

```bash
protoc-gen-go <файл.dsl>              # Генерация Go кода
protoc-gen-go fmt <файл.dsl>          # Форматировать DSL файл
protoc-gen-go --save-bin <файл.dsl>   # Сохранить схему в .bin
protoc-gen-go --load-bin <файл.bin>   # Загрузить схему из .bin и показать код
protoc-gen-go -v                      # Версия
protoc-gen-go -h                      # Справка
```

### Makefile

| Команда | Назначение |
|---------|------------|
| `make build` | собрать бинарный файл |
| `make test` | запустить все тесты (19+) |
| `make test-parser` | тесты парсера (16 тестов) |
| `make test-analyzer` | тесты анализатора (3 теста) |
| `make test-fuzz` | фаззинг-тесты парсера |
| `make bench-quick` | быстрые бенчмарки DSL (только цифры) |
| `make bench-report` | бенчмарки DSL + ручная реализация + отчёт с графиками |
| `make experiment` | полный эксперимент: сравнение с Protobuf и Construct + отчёт |
| `make pipeline` | полный тест pipeline |
| `make demo` | базовая демонстрация |
| `make demo-all` | все демонстрации (8 шт) |
| `make demo-protocols` | демо реальных протоколов |
| `make fmt-dsl` | форматировать все DSL файлы |
| `make save-bin` | сохранить схему в .bin |
| `make load-bin` | загрузить схему из .bin |
| `make examples` | сгенерировать все примеры |
| `make clean` | удалить артефакты |
| `make distclean` | глубокая очистка (включая кэш модулей) |
| `make install` | установить в `$GOPATH/bin` |

## DSL — синтаксис и правила

### Общая структура

```
protocol <ИмяПротокола> {
    id: <HexID>
    [endian: big|little]
    [alias <Имя>: <тип>]
    [const <Имя> = <число>]
    <поле>
    ...
}
```

### Порядок байт (endian)

```
protocol Data {
    id: 0x1000
    endian: little
    value: uint32
}
```

### Константы

```
protocol Config {
    id: 0x9000
    const MAX_SIZE = 16
    buffer: [MAX_SIZE]uint8
}
```

### Алиасы типов

```
protocol Data {
    id: 0x7000
    alias ID: uint32
    user_id: ID
}
```

### Скалярные поля

| Тип | Размер | | Тип | Размер |
|-----|--------|-|-----|--------|
| `uint8` | 1 | | `int8` | 1 |
| `uint16` | 2 | | `int16` | 2 |
| `uint32` | 4 | | `int32` | 4 |
| `uint64` | 8 | | `int64` | 8 |
| `float32` | 4 | | `float64` | 8 |

### Битовые поля (bitstruct)

```
flags: bitstruct {
    ack: bit(7)           // одиночный бит → GetAck() bool
    opcode: bits[6:3]     // диапазон 4 бита → GetOpcode() uint8
}
```

### Enum-типы

```
state: enum {
    OK = 0
    ERROR = 1
}
// Можно использовать в условиях: if state == OK
```

### Вложенные структуры

```
location: struct {
    latitude: float64
    longitude: float64
}
```

Поддерживается произвольная глубина вложенности (проверено до 3 уровней: `struct { struct { struct { ... } } }`).

### Фиксированные массивы

```
readings: [10]float32
points: [5]struct { x: float32; y: float32 }
```

### Слайсы (переменная длина)

```
readings_len: uint16
readings: []float32 length: readings_len
```

### Условные поля

```
// Простое условие
extended: uint32 if flags == 1

// Путь к биту
error_msg: bytes length_from: error_len if flags.ack == 1

// Enum значение
data: bytes length_from: len if state == OK

// Вложенные условия (&&, ||)
data: bytes length_from: len if flags.ack == 1 && count > 5
extended: uint32 if flags.error == 0 || flags.ack == 1
```

Условные поля не занимают места в бинарном представлении, если условие ложно. Это позволяет описывать форматы с переменной структурой (TCP options, IPv6 extension headers, DNS resource records).

## Что генерируется

### Структура

```go
type SensorData struct {
    Device_id    uint32
    Temperature  float32
    Flags        uint8 // bitstruct
    Readings     [10]float32
    Name_len     uint16
    Name         []byte
}
```

### Методы

```go
func (p *SensorData) Size() int
func (p *SensorData) MarshalBinary() ([]byte, error)
func (p *SensorData) UnmarshalBinary(data []byte) error
func (p *SensorData) Validate() error

// Геттеры/сеттеры для битовых полей
func (p *SensorData) GetAck() bool
func (p *SensorData) SetAck(val bool)
```

### Константы смещений

```go
const SensorData_Device_id_Offset = 0
const SensorData_Device_id_Size   = 4
```

## Примеры

### Учебные примеры

| Пример | Описание |
|--------|----------|
| `examples/simple/` | Базовые типы и структуры |
| `examples/bitfields/` | Битовые поля |
| `examples/arrays/` | Массивы и слайсы |
| `examples/conditions/` | Условные поля с путями, &&, \|\| |
| `examples/enums/` | Enum-типы |
| `examples/little_endian/` | LittleEndian |
| `examples/aliases/` | Алиасы типов |
| `examples/consts/` | Константы |

### Реальные протоколы

| Протокол | Файл | Особенности |
|----------|------|-------------|
| MQTT CONNECT | `examples/mqtt/connect.dsl` | IoT, битовые поля, bytes |
| Modbus RTU | `examples/modbus/rtu.dsl` | Промышленный, вложенные условия \|\| |
| TCP Header | `examples/tcp/header.dsl` | Два bitstruct, условный слайс |
| IPv4 Header | `benchmarks/protocols/ipv4/` | bitstruct, условные опции |
| IPv6 Header | `testdata/ipv6/base/` | bitstruct, цепочка расширений |
| IPv6 Extensions | `testdata/ipv6/` | Hop-by-Hop, Fragment, цепочка |
| Ethernet Frame | `examples/ethernet/frame.dsl` | Фиксированные массивы |
| HTTP Request | `examples/http/request.dsl` | Enum + вложенные условия |
| DNS Header | `examples/dns/dns.dsl` | Сетевой протокол |

```bash
make demo              # базовый сенсор
make demo-protocols    # все реальные протоколы
make demo-all          # все демо
make pipeline          # полный тест
```

## Сравнительное тестирование

Проект включает автоматизированный эксперимент для сравнения производительности с аналогами:

```bash
# Только protoc-gen-go (без внешних зависимостей)
make bench-report

# Полное сравнение: DSL vs Hand vs Protobuf vs Construct
# Требует: Python 3, protoc, protoc-gen-go (Google), construct, matplotlib
make experiment
```

Результаты сохраняются в `benchmarks/experiment/report/`:
- `BENCHMARK_REPORT_*.md` — автоматически сгенерированный отчёт с таблицами
- `BENCHMARK_REPORT_LATEST.md` — ссылка на последний отчёт
- `*.png` — графики (9 штук)

Эталонные результаты находятся в `benchmarks/experiment/report/final/`.

Подробнее см. [BENCHMARK_REPORT.md](benchmarks/experiment/report/final/BENCHMARK_REPORT.md).

## Структура проекта

```
.
├── cmd/protoc-gen-go/     # CLI
├── internal/
│   ├── ast/               # AST определения
│   ├── parser/            # Собственный парсер
│   │   ├── lexer.go       # Лексер
│   │   ├── token.go       # Типы токенов
│   │   └── parser.go      # Recursive Descent парсер
│   ├── analyzer/          # Семантический анализ
│   ├── generator/         # Генератор Go-кода
│   ├── binary/            # Бинарный формат
│   └── formatter/         # Форматтер DSL
├── benchmarks/            # Бенчмарки и сравнение с аналогами
│   ├── experiment/        # Экспериментальный фреймворк
│   │   ├── protocols/     # DSL-описания тестовых протоколов
│   │   ├── handwritten/   # Ручная реализация Go
│   │   ├── protobuf/      # Protobuf-версия
│   │   ├── construct/     # Python Construct-версия
│   │   └── report/        # Генератор отчётов и графиков
│   │       ├── final/     # Эталонные результаты
│   │       └── generate_report.py
│   ├── protocols/         # Тестовые DSL-протоколы
│   │   ├── sensor/        # Простой сенсор
│   │   ├── tcp/           # TCP-заголовок
│   │   ├── dns/           # DNS-сообщение
│   │   └── ipv4/          # IPv4-заголовок
│   ├── golang/            # Бенчмарки Go
│   └── compare/           # Сравнение с Protobuf и Construct
├── examples/              # Примеры DSL
│   ├── simple/            # Базовые
│   ├── bitfields/         # Битовые поля
│   ├── arrays/            # Массивы
│   ├── conditions/        # Условия
│   ├── enums/             # Enum
│   ├── little_endian/     # LittleEndian
│   ├── aliases/           # Алиасы
│   ├── consts/            # Константы
│   ├── dns/               # DNS
│   ├── mqtt/              # MQTT
│   ├── modbus/            # Modbus
│   ├── tcp/               # TCP
│   ├── ethernet/          # Ethernet
│   └── http/              # HTTP
├── testdata/              # Данные для тестов
│   ├── ipv6/              # IPv6 заголовки и расширения
│   │   ├── base/          # Базовый IPv6
│   │   ├── hopbyhop/      # Hop-by-Hop Options
│   │   └── fragment/      # Fragment Header
│   └── nested/            # Вложенные структуры
├── demo/                  # Демонстрации
├── docs/                  # Документация
│   ├── grammar.md         # BNF-грамматика
│   ├── parser.md          # Архитектура парсера
│   └── tutorial.md        # Туториал
├── .github/workflows/     # CI/CD
├── Makefile
├── go.mod
├── LICENSE
└── README.md
```

## Документация

- [Туториал](docs/tutorial.md) — пошаговое руководство
- [Грамматика DSL](docs/grammar.md) — полная BNF-нотация
- [Архитектура парсера](docs/parser.md) — описание компонентов
- [Отчёт о сравнительном тестировании](benchmarks/experiment/report/final/BENCHMARK_REPORT.md) — бенчмарки, графики, анализ

## Разработка

### Требования

- Go 1.21+

### Тесты

```bash
make test            # все тесты
make test-parser     # тесты парсера
make test-analyzer   # тесты анализатора
make test-fuzz       # фаззинг-тесты
```

### Бенчмарки

```bash
make bench-quick     # быстрые бенчмарки DSL
make bench-report    # бенчмарки DSL + Hand + отчёт
make experiment      # полное сравнение с аналогами
```

### Сборка

```bash
make build           # собрать бинарник
make install         # установить в $GOPATH/bin
make dev             # полная пересборка
make check           # проверка (fmt + lint + test)
```

## Лицензия

MIT. Полный текст см. в файле [LICENSE](LICENSE).