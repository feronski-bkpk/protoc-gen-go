# Туториал по protoc-gen-go

## Введение

`protoc-gen-go` — это генератор Go-кода для бинарных протоколов. Вы описываете протокол в простом текстовом DSL, а инструмент создаёт типобезопасный Go-код с полной сериализацией.

### Что вы получите

- Структуры данных с правильными типами
- Методы `MarshalBinary()` / `UnmarshalBinary()` / `Size()` / `Validate()`
- Константы смещений для отладки
- Никаких внешних зависимостей — только стандартная библиотека Go

## Установка

```bash
git clone https://github.com/feronski-bkpk/protoc-gen-go.git
cd protoc-gen-go
make build
make install
```

Проверьте установку:

```bash
protoc-gen-go --version
```

## Глава 1: Первый протокол

Создайте файл `sensor.dsl`:

```
protocol SensorData {
    id: 0x1000
    device_id: uint32
    temperature: float32
    humidity: uint8
}
```

Сгенерируйте код:

```bash
protoc-gen-go sensor.dsl
```

Будет создан файл `sensor.gen.go`. Посмотрим что внутри:

```go
type SensorData struct {
    Device_id   uint32
    Temperature float32
    Humidity    uint8
}

func (p *SensorData) Size() int { ... }
func (p *SensorData) MarshalBinary() ([]byte, error) { ... }
func (p *SensorData) UnmarshalBinary(data []byte) error { ... }
func (p *SensorData) Validate() error { ... }
```

### Использование в коде

```go
package main

import (
    "fmt"
    "yourproject/protocol"
)

func main() {
    // Создаём данные
    data := protocol.SensorData{
        Device_id:   0x12345678,
        Temperature: 23.5,
        Humidity:    65,
    }
    
    // Сериализуем в бинарный формат
    binary, _ := data.MarshalBinary()
    fmt.Printf("Размер: %d байт\n", len(binary))
    fmt.Printf("Данные: %x\n", binary)
    
    // Десериализуем обратно
    var decoded protocol.SensorData
    decoded.UnmarshalBinary(binary)
    fmt.Printf("Температура: %.1f°C\n", decoded.Temperature)
}
```

## Глава 2: Вложенные структуры

Протоколы могут содержать вложенные структуры:

```
protocol GPSData {
    id: 0x2000
    device_id: uint32
    location: struct {
        latitude: float64
        longitude: float64
        altitude: int32
    }
    timestamp: uint64
}
```

Генерируется две структуры:

```go
type Location struct {
    Latitude  float64
    Longitude float64
    Altitude  int32
}

type GPSData struct {
    Device_id uint32
    Location  Location
    Timestamp uint64
}
```

Каждая структура имеет свои методы `Size()`, `MarshalBinary()`, `UnmarshalBinary()`.

## Глава 3: Битовые поля

Для экономии места можно упаковать несколько флагов в один байт:

```
protocol Flags {
    id: 0x3000
    flags: bitstruct {
        ack: bit(7)           // одиночный бит
        error: bit(6)
        priority: bits[5:4]   // диапазон 2 бита (0-3)
        mode: bits[3:0]       // диапазон 4 бита (0-15)
    }
    value: uint32
}
```

Генерируются геттеры и сеттеры:

```go
// Одиночный бит
func (p *Flags) GetAck() bool { return (p.Flags & (1 << 7)) != 0 }
func (p *Flags) SetAck(v bool) { if v { p.Flags |= 1 << 7 } else { p.Flags &^= 1 << 7 } }

// Диапазон битов
func (p *Flags) GetPriority() uint8 { return (p.Flags >> 4) & 0x03 }
func (p *Flags) SetPriority(v uint8) {
    p.Flags = (p.Flags & 0xCF) | ((v & 0x03) << 4)
}
```

## Глава 4: Условные поля

Поля могут присутствовать только при выполнении условия:

```
protocol ConditionalData {
    id: 0x4000
    flags: uint8
    data_len: uint16
    data: bytes length_from: data_len if flags == 1
    extended: uint32 if flags == 2
}
```

В сгенерированном коде:

```go
func (p *ConditionalData) Size() int {
    size := 0
    size += 1 // flags
    size += 2 // data_len
    if p.Flags == 1 {
        size += len(p.Data)
    }
    if p.Flags == 2 {
        size += 4
    }
    return size
}
```

### Условия с путями

Можно ссылаться на конкретные биты:

```
protocol Advanced {
    id: 0x5000
    flags: bitstruct {
        ack: bit(7)
        error: bit(6)
    }
    error_msg_len: uint16
    error_msg: bytes length_from: error_msg_len if flags.error == 1
}
```

### Вложенные условия

Поддерживаются `&&` и `||`:

```
data: bytes length_from: data_len if flags == 1 && count > 5
extended: uint32 if flags.error == 0 || flags.ack == 1
```

## Глава 5: Массивы и слайсы

### Фиксированные массивы

Размер известен на этапе компиляции:

```
protocol FixedArrays {
    id: 0x6000
    readings: [10]float32
    points: [5]struct {
        x: float32
        y: float32
    }
}
```

Генерируется:

```go
type FixedArrays struct {
    Readings [10]float32
    Points   [5]PointsElem
}
```

### Слайсы (переменная длина)

Длина хранится в отдельном поле:

```
protocol SliceData {
    id: 0x7000
    readings_len: uint16
    readings: []float32 length: readings_len
}
```

При десериализации создаётся слайс нужного размера:

```go
func (p *SliceData) UnmarshalBinary(data []byte) error {
    // ...
    p.Readings = make([]float32, p.Readings_len)
    for i := 0; i < int(p.Readings_len); i++ {
        p.Readings[i] = math.Float32frombits(binary.BigEndian.Uint32(data[offset:]))
        offset += 4
    }
    // ...
}
```

## Глава 6: Enum-типы

```
protocol Status {
    id: 0x8000
    state: enum {
        OK = 0
        ERROR = 1
        PENDING = 2
    }
    value: uint32
}
```

Генерируется:

```go
type StateEnum uint8

const (
    OK      StateEnum = 0
    ERROR   StateEnum = 1
    PENDING StateEnum = 2
)

type Status struct {
    State StateEnum
    Value uint32
}
```

Enum можно использовать в условиях:

```
extended: uint32 if state == OK
```

## Глава 7: Порядок байт

По умолчанию используется Big Endian. Для Little Endian добавьте `endian: little`:

```
protocol LEData {
    id: 0x9000
    endian: little
    value: uint32
}
```

Генерируется:

```go
func (p *LEData) MarshalBinary() ([]byte, error) {
    // ...
    binary.LittleEndian.PutUint32(buf[offset:], p.Value)
    // ...
}
```

## Глава 8: Алиасы типов

Для улучшения читаемости:

```
protocol AliasData {
    id: 0xA000
    alias ID: uint32
    alias Name: bytes
    user_id: ID
    username: Name length_from: name_len
    name_len: uint16
}
```

## Глава 9: Константы

Для избежания магических чисел:

```
protocol Config {
    id: 0xB000
    const MAX_SIZE = 16
    const TIMEOUT = 30
    buffer: [MAX_SIZE]uint8
    timeout: uint16
}
```

## Глава 10: Работа с бинарным форматом

Схему протокола можно сохранить в компактный `.bin` формат:

```bash
protoc-gen-go --save-bin sensor.dsl
# Создаст sensor.bin (303 байта)
```

И загрузить обратно:

```bash
protoc-gen-go --load-bin sensor.bin
# Выведет сгенерированный Go-код
```

## Глава 11: Форматирование DSL

```bash
protoc-gen-go fmt sensor.dsl
```

Выведет отформатированный DSL с правильными отступами.

## Глава 12: Makefile команды

| Команда | Назначение |
|---------|------------|
| `make build` | Собрать бинарный файл |
| `make test` | Запустить все тесты |
| `make bench` | Бенчмарки |
| `make demo` | Демонстрация |
| `make demo-all` | Все демонстрации |
| `make examples` | Сгенерировать все примеры |
| `make clean` | Очистить артефакты |

## Реальные примеры

В директории `examples/` находятся готовые протоколы:

| Протокол | Файл | Особенности |
|----------|------|-------------|
| MQTT CONNECT | `examples/mqtt/connect.dsl` | IoT, битовые поля, bytes |
| Modbus RTU | `examples/modbus/rtu.dsl` | Промышленный, вложенные условия |
| TCP Header | `examples/tcp/header.dsl` | Два bitstruct, условный слайс |
| Ethernet Frame | `examples/ethernet/frame.dsl` | Фиксированные массивы |
| HTTP Request | `examples/http/request.dsl` | Enum с вложенными условиями |
| DNS Header | `examples/dns/dns.dsl` | Сетевой протокол |