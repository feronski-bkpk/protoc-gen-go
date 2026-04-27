# Архитектура парсера protoc-gen-go

## Общая схема

```
  Входной текст (.dsl)
          │
          ▼
    ┌────────────┐
    │   Лексер   │  Разбивает текст на токены (token.go, lexer.go)
    └─────┬──────┘
          │
          ▼
    ┌────────────┐
    │   Парсер   │  Recursive Descent, строит AST (parser.go)
    └─────┬──────┘
          │
          ▼
    ┌────────────┐
    │ Анализатор │  Семантический анализ, таблица символов (analyzer.go)
    └─────┬──────┘
          │
          ▼
    ┌────────────┐
    │ Генератор  │  Генерирует Go-код (generator.go)
    └─────┬──────┘
          │
          ▼
    ┌────────────┐
    │ Форматтер  │  Форматирует DSL (formatter.go)
    └────────────┘
```

Также поддерживается бинарный формат для сохранения/загрузки схемы (`binary/`).

## Лексер

Лексер — это конечный автомат, который проходит по символам входного текста и выделяет токены.

### Типы токенов

| Токен | Пример | Описание |
|-------|--------|----------|
| `TokenIdent` | `sensor_id`, `ID`, `big` | Идентификатор |
| `TokenNumber` | `10`, `256` | Целое число |
| `TokenHexNumber` | `0x1234` | Шестнадцатеричное число |
| `TokenString` | `"hello"` | Строка (для будущего использования) |
| `TokenProtocol` | `protocol` | Ключевое слово |
| `TokenStruct` | `struct` | Ключевое слово |
| `TokenBitStruct` | `bitstruct` | Ключевое слово |
| `TokenEnum` | `enum` | Ключевое слово |
| `TokenID` | `id` | Ключевое слово (deprecated) |
| `TokenIf` | `if` | Ключевое слово |
| `TokenLengthFrom` | `length_from` | Ключевое слово |
| `TokenLength` | `length` | Ключевое слово |
| `TokenColon` | `:` | Пунктуация |
| `TokenLBrace` | `{` | Пунктуация |
| `TokenRBrace` | `}` | Пунктуация |
| `TokenLBracket` | `[` | Пунктуация |
| `TokenRBracket` | `]` | Пунктуация |
| `TokenLParen` | `(` | Пунктуация |
| `TokenRParen` | `)` | Пунктуация |
| `TokenComma` | `,` | Пунктуация |
| `TokenDot` | `.` | Пунктуация (для путей) |
| `TokenEqAssign` | `=` | Присваивание (для enum/const) |
| `TokenEq` | `==` | Оператор сравнения |
| `TokenNotEq` | `!=` | Оператор сравнения |
| `TokenLT` | `<` | Оператор сравнения |
| `TokenGT` | `>` | Оператор сравнения |
| `TokenLTE` | `<=` | Оператор сравнения |
| `TokenGTE` | `>=` | Оператор сравнения |
| `TokenAnd` | `&&` | Логическое И |
| `TokenOr` | `\|\|` | Логическое ИЛИ |

### Алгоритм

```go
func (l *Lexer) Tokenize() ([]Token, error) {
    for l.pos < len(l.input) {
        ch := l.current()
        
        switch {
        case unicode.IsSpace(ch):
            l.advance()
        case ch == '/' && l.peek() == '/':
            l.skipComment()
        case ch == '0' && l.peek() == 'x':
            l.tokenizeHex()
        case unicode.IsDigit(ch):
            l.tokenizeNumber()
        case unicode.IsLetter(ch) || ch == '_':
            l.tokenizeIdent()  // также определяет ключевые слова
        case ch == '"':
            l.tokenizeString()
        default:
            l.tokenizePunctOrOperator()
        }
    }
    return l.tokens, nil
}
```

Ключевые слова определяются в `tokenizeIdent()` через `strings.ToLower(value)`.

## Парсер

Парсер — это **Recursive Descent Parser**. Каждое правило грамматики представлено отдельным методом.

### Иерархия методов

```
Parse()
  └── parseProtocol()
        ├── parseFields()
        │     └── parseField()
        │           ├── parseStructField()
        │           ├── parseBitStructField()
        │           ├── parseEnumField()
        │           ├── parseArrayField()
        │           │     └── parseSliceField()
        │           ├── parseBytesField()
        │           └── parseScalarField()
        └── parseCondition()
              └── parseSingleCondition()
```

### Принцип работы

1. **ParseProtocol** — верхний уровень, парсит `protocol Name { ... }`. Обрабатывает `id`, `endian`, `alias`, `const`, затем поля.
2. **ParseFields** — цикл по полям до `}`
3. **ParseField** — определяет тип поля по первому токену:
   - `struct` → parseStructField
   - `bitstruct` → parseBitStructField
   - `enum` → parseEnumField
   - `[` → parseArrayField
   - скалярный тип (`uint8`...) → ScalarField
   - `bytes` → parseBytesField
4. **ParseCondition** — парсит условие `if expr`. Поддерживает:
   - Пути: `flags.ack == 1`
   - Enum: `state == OK`
   - Вложенные: `a == 1 && b > 5`

### Контекстная обработка ошибок

Парсер отслеживает контекст через `setContext()`:

```go
p.setContext("поля '" + nameTok.Value + "'")
```

Сообщения об ошибках содержат контекст:
```
ожидался тип поля при парсинге поля 'field' (строка 3, колонка 23: 'nonexistent')
```

### Подстановка алиасов и констант

Алиасы подставляются в `parseField()`:
```go
if p.aliases != nil {
    if baseType, ok := p.aliases[typeName]; ok {
        typeName = baseType
    }
}
```

Константы подставляются в `parseArrayField()` для размера массива:
```go
if constVal, ok := p.consts[constName]; ok {
    size = constVal
}
```

## Анализатор

После парсинга AST проходит семантический анализ:

1. **buildSymbolTable** — строит таблицу символов с полными путями. Для битовых полей добавляет пути `flags.ack`, `flags.error`.
2. **computeOffsets** — вычисляет смещения полей с учётом вложенности.
3. **validateReferences** — проверяет что `length_from` и условия ссылаются на существующие поля.
4. **validateBitFields** — проверяет перекрытие битов в `bitstruct`.
5. **validateCycles** — проверяет циклические зависимости в структурах.

### Таблица символов

```go
type SymbolInfo struct {
    Field  ast.Field
    Path   string    // "location.latitude"
    Offset int       // 6
    Size   int       // 8
}
```

## Генератор

Генерирует Go-код со следующими компонентами:

1. **Заголовок** — `// Code generated by protoc-gen-go. DO NOT EDIT.`
2. **Импорты** — `encoding/binary`, `math`, `fmt` (при необходимости)
3. **Enum-типы** — `type StateEnum uint8` с константами
4. **Структуры** — основная и вложенные (с суффиксом `Elem`)
5. **Методы**:
   - `Size() int`
   - `MarshalBinary() ([]byte, error)`
   - `UnmarshalBinary([]byte) error`
   - `Validate() error`
   - Геттеры/сеттеры для битовых полей
6. **Константы смещений** — `Protocol_Field_Offset`, `Protocol_Field_Size`

### Особенности кодогенерации

- **Endian**: проверяется `g.proto.Endian` для выбора `BigEndian`/`LittleEndian`
- **Условия**: рекурсивно генерируются через `conditionToGo()` с поддержкой `&&`/`||`
- **Enum в условиях**: `cond.EnumValue` выводится как идентификатор
- **Слайсы**: `for range` для скаляров (без переменной), `for i := range` для структур

## Форматтер

Форматирует DSL через обратную генерацию из AST:

1. Парсит DSL в AST
2. Обходит AST и выводит с правильными отступами
3. Сохраняет оригинальные имена алиасов через `OriginalType`
4. Выводит `enum` значения с правильными отступами
5. Поддерживает вложенные условия через `writeCondition()`

## Бинарный формат

Схему протокола можно сохранить в компактный `.bin` формат и восстановить обратно.

### Структура .bin файла

```
┌─────────────────────────────────┐
│ Magic: "PROT" (4 байта)         │
│ Version: 1 (2 байта)            │
│ PacketID (2 байта)              │
│ FieldCount (2 байта)            │
│ Flags (2 байта)                 │
│ Reserved (4 байта)              │
├─────────────────────────────────┤
│ Name (string)                   │
├─────────────────────────────────┤
│ Fields...                       │
│   Type (1 байт)                 │
│   NameLen (2 байта)             │
│   Name (N байт)                 │
│   Size/SubFields/...            │
│   Condition (опционально)       │
└─────────────────────────────────┘
```

## Производительность

Бенчмарки на Intel Core i3-7020U @ 2.30GHz:

| Операция | Время | Память |
|----------|-------|--------|
| Лексер (сложный DSL) | 20.7 µs | 15.8 KB |
| Парсер (сложный DSL) | 16.7 µs | 9.9 KB |
| Полный pipeline | 38.8 µs | 25.7 KB |
| Лексер (простой DSL) | 2.1 µs | 1.5 KB |
| Парсер (простой DSL) | 3.5 µs | 2.3 KB |