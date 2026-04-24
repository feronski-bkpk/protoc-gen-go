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
- сохранить схему протокола в компактный бинарный формат

**Не требует внешних зависимостей** — только стандартная библиотека Go.

## Принцип работы

```
DSL (.dsl) → Собственный парсер → AST → Анализатор → Генератор → Go-код (.gen.go)
                  ↕
           Бинарный формат (.bin)
```

1. Пользователь описывает протокол в файле `.dsl`
2. Собственный парсер (Recursive Descent) строит AST
3. Анализатор проверяет семантику и строит таблицу символов
4. Генератор создаёт Go-файл со структурами и методами сериализации
5. Схему можно сохранить в `.bin` и восстановить обратно

Порядок байт — **Big Endian** (сетевой порядок).

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
protoc-gen-go --save-bin <файл.dsl>   # Сохранить схему в .bin
protoc-gen-go --load-bin <файл.bin>   # Загрузить схему из .bin
protoc-gen-go -v                      # Версия
protoc-gen-go -h                      # Справка
```

Создаёт `<имя_файла>.gen.go` в той же директории.

### Makefile

| Команда | Назначение |
|---------|------------|
| `make build` | собрать бинарный файл |
| `make test` | запустить все тесты |
| `make test-parser` | тесты парсера |
| `make test-analyzer` | тесты анализатора |
| `make test-binary` | тесты бинарного формата |
| `make demo` | демонстрация базового протокола |
| `make demo-arrays` | демонстрация слайсов |
| `make demo-dns` | демонстрация DNS |
| `make demo-conditions` | демонстрация условий с путями |
| `make save-bin` | сохранить схему в бинарный формат |
| `make load-bin` | загрузить схему из бинарного формата |
| `make clean` | удалить артефакты |
| `make fmt` | форматировать код |
| `make lint` | запустить go vet |
| `make install` | установить в `$GOPATH/bin` |

## DSL — синтаксис и правила

### Общая структура

```
protocol <ИмяПротокола> {
    id: <HexID>
    <поле>
    ...
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
    ack: bit(7)           // одиночный бит → GetAck() bool, SetAck(bool)
    opcode: bits[6:3]     // диапазон 4 бита → GetOpcode() uint8, SetOpcode(uint8)
}
```

### Вложенные структуры

```
location: struct {
    latitude: float64
    longitude: float64
}
```

### Поля переменной длины (bytes)

```
name_len: uint16
name: bytes length_from: name_len
```

### Фиксированные массивы

```
readings: [10]float32
points: [5]struct {
    x: float32
    y: float32
}
```

### Слайсы (массивы переменной длины)

```
readings_len: uint16
readings: []float32 length: readings_len

samples: []struct {
    x: float32
    y: float32
} length: samples_len
```

### Условные поля

```
extended: uint32 if flags == 1
error_msg: bytes length_from: error_len if flags.ack == 1
```

Поддерживаются пути: `flags.ack == 1`

Операторы: `==`, `!=`, `<`, `>`, `<=`, `>=`.

### Комментарии

```
// однострочный комментарий
field: uint16   // после поля
```

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

Для анонимных структур создаётся тип с суффиксом `Elem` (например, `SamplesElem`).

### Методы

```go
func (p *SensorData) Size() int
func (p *SensorData) MarshalBinary() ([]byte, error)
func (p *SensorData) UnmarshalBinary(data []byte) error
func (p *SensorData) Validate() error

// Для битовых полей
func (p *SensorData) GetAck() bool
func (p *SensorData) SetAck(val bool)
func (p *SensorData) GetOpcode() uint8
func (p *SensorData) SetOpcode(val uint8)
```

### Константы смещений

```go
const SensorData_Device_id_Offset = 0
const SensorData_Device_id_Size   = 4
```

### Бинарный формат

Схему протокола можно сохранить в компактный бинарный формат:

```bash
protoc-gen-go --save-bin sensor.dsl    # создаст sensor.bin
protoc-gen-go --load-bin sensor.bin    # сгенерирует код из .bin
```

## Примеры

| Пример | Описание |
|--------|----------|
| `examples/simple/` | Базовые типы и структуры |
| `examples/bitfields/` | Битовые поля и DNS флаги |
| `examples/arrays/` | Массивы и слайсы |
| `examples/dns/` | DNS протокол |
| `examples/conditions/` | Условные поля с путями |
| `demo/run.go` | Автономное демо |

```bash
make demo              # базовый сенсор
make demo-arrays       # слайсы
make demo-dns          # DNS протокол
make demo-conditions   # условия с путями
```

## Структура проекта

```
.
├── cmd/protoc-gen-go/     # CLI
├── internal/
│   ├── ast/               # AST определения
│   ├── parser/            # Собственный парсер
│   │   ├── lexer.go       # Лексер (токенизация)
│   │   ├── token.go       # Типы токенов
│   │   └── parser.go      # Recursive Descent парсер
│   ├── analyzer/          # Семантический анализ
│   │   ├── analyzer.go    # Таблица символов, валидация
│   │   └── analyzer_test.go
│   ├── generator/         # Генератор Go-кода
│   │   └── generator.go   # Size/Marshal/Unmarshal/Validate
│   └── binary/            # Бинарный формат
│       ├── types.go       # Константы типов
│       ├── writer.go      # Сериализация AST → bin
│       └── reader.go      # Десериализация bin → AST
├── pkg/protocol/          # Runtime (опционально)
├── examples/              # Примеры DSL
├── demo/                  # Демонстрация
├── testdata/              # Тесты
├── docs/                  # Документация
│   ├── grammar.md         # BNF-грамматика
│   └── parser.md          # Архитектура парсера
├── Makefile
├── go.mod
├── LICENSE
└── README.md
```

## Документация

- [Грамматика DSL](docs/grammar.md) — полная BNF-нотация
- [Архитектура парсера](docs/parser.md) — описание компонентов

## Разработка

### Требования

- Go 1.21+

### Тесты

```bash
make test            # все тесты (парсер + анализатор + бинарный)
make test-parser     # только парсер (13 тестов)
make test-analyzer   # только анализатор (3 теста)
make test-binary     # бинарный формат
```

### Сборка

```bash
make build           # собрать бинарник
make install         # установить в $GOPATH/bin
make dev             # полная пересборка
```

## Лицензия

MIT. Полный текст см. в файле [LICENSE](LICENSE).