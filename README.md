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

Порядок байт по умолчанию — **Big Endian** (сетевой порядок). Поддерживается **LittleEndian** через `endian: little`.

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
    endian: big
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
protoc-gen-go --load-bin <файл.bin>   # Загрузить схему из .bin и показать код
protoc-gen-go -v                      # Версия
protoc-gen-go -h                      # Справка
```

### Makefile

| Команда | Назначение |
|---------|------------|
| `make build` | собрать бинарный файл |
| `make test` | запустить все тесты |
| `make test-parser` | тесты парсера |
| `make test-analyzer` | тесты анализатора |
| `make test-fuzz` | фаззинг-тесты парсера |
| `make bench` | бенчмарки |
| `make demo` | демонстрация базового протокола |
| `make demo-all` | запустить все демонстрации |
| `make save-bin` | сохранить схему в .bin |
| `make load-bin` | загрузить схему из .bin |
| `make examples` | сгенерировать все примеры |
| `make clean` | удалить артефакты |
| `make fmt` | форматировать код |
| `make lint` | запустить go vet |
| `make install` | установить в `$GOPATH/bin` |

## DSL — синтаксис и правила

### Общая структура

```
protocol <ИмяПротокола> {
    id: <HexID>
    [endian: big|little]
    [alias <Имя>: <тип>]
    <поле>
    ...
}
```

### Порядок байт (endian)

```
protocol Data {
    id: 0x1000
    endian: little    // big (по умолчанию) или little
    value: uint32
}
```

### Алиасы типов

```
protocol Data {
    id: 0x7000
    alias ID: uint32
    alias Name: bytes
    user_id: ID
    username: Name length_from: name_len
    name_len: uint16
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

### Enum-типы

```
state: enum {
    OK = 0
    ERROR = 1
    PENDING = 2
}
```

Генерируется как отдельный тип с константами.

### Вложенные структуры

```
location: struct {
    latitude: float64
    longitude: float64
}
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
```

### Условные поля

```
extended: uint32 if flags == 1
error_msg: bytes length_from: error_len if flags.ack == 1
```

Поддерживаются пути: `flags.ack == 1`. Операторы: `==`, `!=`, `<`, `>`, `<=`, `>=`.

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

### Методы

```go
func (p *SensorData) Size() int
func (p *SensorData) MarshalBinary() ([]byte, error)
func (p *SensorData) UnmarshalBinary(data []byte) error
func (p *SensorData) Validate() error

// Для битовых полей
func (p *SensorData) GetAck() bool
func (p *SensorData) SetAck(val bool)
```

### Константы смещений

```go
const SensorData_Device_id_Offset = 0
const SensorData_Device_id_Size   = 4
```

## Примеры

| Пример | Описание |
|--------|----------|
| `examples/simple/` | Базовые типы и структуры |
| `examples/bitfields/` | Битовые поля |
| `examples/arrays/` | Массивы и слайсы |
| `examples/dns/` | DNS протокол |
| `examples/conditions/` | Условные поля с путями |
| `examples/enums/` | Enum-типы |
| `examples/little_endian/` | LittleEndian |
| `examples/aliases/` | Алиасы типов |

```bash
make demo              # базовый сенсор
make demo-arrays       # слайсы
make demo-dns          # DNS
make demo-conditions   # условия с путями
make demo-enum         # enum-типы
make demo-endian       # LittleEndian
make demo-all          # все демо
```

## Структура проекта

```
.
├── cmd/protoc-gen-go/     # CLI
├── internal/
│   ├── ast/               # AST определения
│   ├── parser/            # Собственный парсер
│   ├── analyzer/          # Семантический анализ
│   ├── generator/         # Генератор Go-кода
│   └── binary/            # Бинарный формат
├── examples/              # Примеры DSL
│   ├── simple/
│   ├── bitfields/
│   ├── arrays/
│   ├── dns/
│   ├── conditions/
│   ├── enums/
│   ├── little_endian/
│   └── aliases/
├── demo/                  # Демонстрация
├── docs/                  # Документация
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
make test            # все тесты
make test-parser     # тесты парсера
make test-analyzer   # тесты анализатора
make test-fuzz       # фаззинг-тесты
make bench           # бенчмарки
```

### Сборка

```bash
make build           # собрать бинарник
make install         # установить в $GOPATH/bin
make dev             # полная пересборка
```

## Лицензия

MIT. Полный текст см. в файле [LICENSE](LICENSE).