# protoc-gen-go

Генератор Go-кода для бинарных протоколов, описанных в человеко-читаемом DSL.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Назначение

Инструмент позволяет:

- описать бинарный протокол в декларативном текстовом формате (DSL);
- автоматически получить строго типизированные Go-структуры;
- сгенерировать методы `MarshalBinary` / `UnmarshalBinary` / `Size`;
- получить константы смещений для отладки и прямого доступа к полям.

Проект ориентирован на случаи, где важен **фиксированный бинарный формат**:
встраиваемые системы, сетевые протоколы, драйверы, компактное межсервисное взаимодействие.

## Принцип работы

1. Пользователь описывает протокол в файле `.dsl`.
2. Парсер (на базе `participle`) строит AST.
3. Генератор создаёт Go-файл со структурами и методами сериализации.
4. Сгенерированный код не требует внешних зависимостей (только стандартная библиотека).

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
# Создайте DSL файл
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

# Сгенерируйте код
protoc-gen-go sensor.dsl

# Используйте в проекте
```

## Использование

### CLI

```bash
protoc-gen-go <файл.dsl>
```

После выполнения будет создан файл `<имя_файла>.gen.go` в той же директории.

### Makefile

| Команда | Назначение |
|---------|------------|
| `make build` | собрать бинарный файл |
| `make test` | модульные тесты |
| `make test-integration` | интеграционные тесты |
| `make test-all` | все тесты |
| `make demo` | демонстрация базового протокола |
| `make demo-arrays` | демонстрация слайсов |
| `make demo-dns` | демонстрация DNS протокола |
| `make clean` | удалить артефакты |
| `make fmt` | форматировать код |
| `make lint` | запустить go vet |
| `make install` | установить в `$GOPATH/bin` |
| `make help` | показать справку |

## DSL — синтаксис и правила

### Общая структура

```
protocol <ИмяПротокола> {
    id: <HexID>
    <поле>
    ...
}
```

- `<ИмяПротокола>` — допустимый Go-идентификатор.
- `<HexID>` — идентификатор протокола в шестнадцатеричном виде (например, `0x1234`).

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
    reserved: bits[2:0]   // диапазон 3 бита → GetReserved() uint8, SetReserved(uint8)
}
```

### Вложенные структуры

```
location: struct {
    latitude: float64
    longitude: float64
    altitude: int32
}
```

### Поля переменной длины (bytes)

```
name_len: uint16
name: bytes length_from: name_len
```

### Слайсы (массивы переменной длины)

```
readings_len: uint16
readings: []float32 length: readings_len

samples_len: uint8
samples: []struct {
    x: float32
    y: float32
    z: float32
} length: samples_len
```

Для слайсов обязательно указывать поле длины через `length: <поле>`.

### Условные поля

```
error_msg: bytes length_from: error_len if flags == 1
extended: uint32 if flags == 2
```

Поддерживаемые операторы: `==`, `!=`, `<`, `>`, `<=`, `>=`.

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
    Readings_len uint16
    Readings     []float32
    Name_len     uint16
    Name         []byte
}
```

Для анонимных структур в слайсах создаётся отдельный тип с суффиксом `Elem` (например, `SamplesElem`).

### Методы

```go
func (p *SensorData) Size() int
func (p *SensorData) MarshalBinary() ([]byte, error)
func (p *SensorData) UnmarshalBinary(data []byte) error

// Для bit(7)
func (p *SensorData) GetAck() bool
func (p *SensorData) SetAck(val bool)

// Для bits[5:4]
func (p *SensorData) GetPriority() uint8
func (p *SensorData) SetPriority(val uint8)
```

### Константы смещений

```go
const SensorData_Device_id_Offset = 0
const SensorData_Device_id_Size   = 4
const SensorData_Flags_Offset     = 8
const SensorData_Flags_Size       = 1
// Для динамических полей смещение не вычисляется
```

## Примеры

| Пример | Описание |
|--------|----------|
| `examples/simple/` | Базовые типы и структуры |
| `examples/bitfields/` | Битовые поля и DNS флаги |
| `examples/arrays/` | Слайсы с полем длины |
| `examples/dns/` | DNS протокол |
| `demo/run.go` | Автономное демо сенсора |

Запуск демо:
```bash
make demo          # базовый сенсор
make demo-arrays   # слайсы
make demo-dns      # DNS протокол
```

## Статус проекта

### Реализовано
- Все скалярные типы
- Вложенные структуры
- Битовые поля (`bitstruct` с `bit()` и `bits[high:low]`)
- Поля `bytes` с `length_from`
- Слайсы `[]type` и `[]struct` с полем `length`
- Условные поля (`if`)
- BigEndian кодирование
- Константы смещений
- Модульные тесты (11 тестов)
- Интеграционные тесты
- Демонстрация

## Структура проекта

```
.
├── cmd/protoc-gen-go/     # CLI
├── internal/
│   ├── ast/               # AST определения
│   ├── dsl/               # парсер DSL
│   └── generator/         # генератор Go-кода
├── pkg/protocol/          # runtime
├── examples/              # примеры DSL
│   ├── simple/            # базовые примеры
│   ├── bitfields/         # битовые поля
│   ├── arrays/            # слайсы
│   └── dns/               # DNS протокол
├── demo/                  # демонстрация
├── testdata/              # интеграционные тесты
├── Makefile
├── go.mod
├── LICENSE
└── README.md
```

## Разработка

### Требования
- Go 1.21+

### Запуск тестов
```bash
make test              # модульные тесты
make test-integration  # интеграционные тесты
make test-all          # все тесты
```

### Сборка
```bash
make build         # собрать бинарник
make install       # установить в $GOPATH/bin
```

## Лицензия

MIT. Полный текст см. в файле [LICENSE](LICENSE).