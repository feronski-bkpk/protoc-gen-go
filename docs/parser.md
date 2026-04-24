# Архитектура парсера

## Общая схема

```
  Входной текст (.dsl)
          │
          ▼
    ┌────────────┐
    │   Лексер   │  Разбивает текст на токены
    └─────┬──────┘
          │
          ▼
    ┌────────────┐
    │   Парсер   │  Recursive Descent, строит AST
    └─────┬──────┘
          │
          ▼
    ┌────────────┐
    │ Анализатор │  Семантический анализ, таблица символов
    └─────┬──────┘
          │
          ▼
    ┌────────────┐
    │ Генератор  │  Генерирует Go-код
    └────────────┘
```

## Лексер

Лексер — это конечный автомат, который проходит по символам входного текста и выделяет токены.

### Типы токенов

| Токен | Пример | Описание |
|-------|--------|----------|
| IDENT | `sensor_id` | Идентификатор |
| NUMBER | `10` | Целое число |
| HEX | `0x1234` | Шестнадцатеричное число |
| KEYWORD | `protocol`, `struct` | Ключевые слова |
| PUNCT | `{`, `}`, `:`, `[`, `]` | Знаки пунктуации |
| OPERATOR | `==`, `!=`, `<`, `>` | Операторы сравнения |

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
        │           ├── parseArrayField()
        │           │     └── parseSliceField()
        │           ├── parseBytesField()
        │           └── parseScalarField()
        └── parseCondition()
```

### Принцип работы

1. **ParseProtocol** — верхний уровень, парсит `protocol Name { ... }`
2. **ParseFields** — цикл по полям до `}`
3. **ParseField** — определяет тип поля по первому токену:
   - `struct` → parseStructField
   - `bitstruct` → parseBitStructField
   - `[` → parseArrayField
   - тип (`uint8`, ...) → parseScalarField
   - `bytes` → parseBytesField

### Обработка ошибок

Парсер использует метод `expect()` который паникует с детальным сообщением:
```
ожидался IDENT, получен NUMBER (строка 5, колонка 10: '123')
```

## Анализатор

После парсинга AST проходит семантический анализ:

1. **buildSymbolTable** — строит таблицу символов с полными путями
2. **computeOffsets** — вычисляет смещения полей
3. **validateReferences** — проверяет ссылки `length_from`
4. **validateBitFields** — проверяет перекрытие битов
5. **validateCycles** — проверяет циклические зависимости

### Таблица символов

```go
type SymbolInfo struct {
    Field  ast.Field
    Path   string   // "location.latitude"
    Offset int      // 6
    Size   int      // 8
}
```