# Грамматика DSL protoc-gen-go

## BNF нотация

```
<protocol>       ::= "protocol" <identifier> "{"
                     <id-field>
                     [ <endian-spec> ]
                     { <alias-def> }
                     { <const-def> }
                     <field-list>
                     "}"

<id-field>       ::= "id" ":" <hex-number>

<endian-spec>    ::= "endian" ":" ("big" | "little")

<alias-def>      ::= "alias" <identifier> ":" <type-name>

<const-def>      ::= "const" <identifier> "=" <number>

<field-list>     ::= { <field> }

<field>          ::= <identifier> ":" <field-type> [ <length-spec> ] [ <condition> ]

<field-type>     ::= <scalar-type>
                   | <struct-type>
                   | <bitstruct-type>
                   | <enum-type>
                   | <bytes-type>
                   | <array-type>
                   | <slice-type>

<scalar-type>    ::= "uint8"  | "uint16" | "uint32" | "uint64"
                   | "int8"   | "int16"  | "int32"  | "int64"
                   | "float32" | "float64"

<struct-type>    ::= "struct" "{" <field-list> "}"

<bitstruct-type> ::= "bitstruct" "{" <bit-field-list> "}"

<bit-field-list> ::= { <bit-field> }

<bit-field>      ::= <identifier> ":" <bit-spec>

<bit-spec>       ::= "bit" "(" <number> ")"
                   | "bits" "[" <number> ":" <number> "]"

<enum-type>      ::= "enum" "{" <enum-value-list> "}"

<enum-value-list>::= { <identifier> "=" <number> }

<bytes-type>     ::= "bytes"

<array-type>     ::= "[" ( <number> | <identifier> ) "]" ( <scalar-type> | <struct-type> )

<slice-type>     ::= "[]" ( <scalar-type> | <struct-type> )

<length-spec>    ::= "length" ":" <identifier>
                   | "length_from" ":" <identifier>

<condition>      ::= "if" <condition-expr>

<condition-expr> ::= <simple-cond> { ( "&&" | "||" ) <simple-cond> }

<simple-cond>    ::= <path> <operator> <value>

<path>           ::= <identifier> { "." <identifier> }

<operator>       ::= "==" | "!=" | "<" | ">" | "<=" | ">="

<value>          ::= <number> | <hex-number> | <identifier>

<type-name>      ::= "uint8" | "uint16" | "uint32" | "uint64"
                   | "int8" | "int16" | "int32" | "int64"
                   | "float32" | "float64" | "bytes"

<identifier>     ::= [a-zA-Z_][a-zA-Z0-9_]*

<number>         ::= [0-9]+

<hex-number>     ::= "0x" [0-9a-fA-F]+

<comment>        ::= "//" { <any-char> } <newline>
```

## Синтаксические диаграммы

### Протокол

```
┌───────────────────────────────────────────────────────────────────────────────────┐
│ protocol → IDENT "{" → id: HEX → [endian] → [aliases] → [consts] → { поле } → "}" │
└───────────────────────────────────────────────────────────────────────────────────┘
```

### Типы полей

```
скаляр:      IDENT ":" (uint8|uint16|uint32|uint64|int8|int16|int32|int64|float32|float64)
структура:   IDENT ":" "struct" "{" { поле } "}"
биты:        IDENT ":" "bitstruct" "{" { IDENT ":" bit(N) | bits[H:L] } "}"
enum:        IDENT ":" "enum" "{" { IDENT = N } "}"
bytes:       IDENT ":" "bytes" [ length: IDENT | length_from: IDENT ]
массив:      IDENT ":" "[" N "]" (скаляр | структура)
слайс:       IDENT ":" "[]" (скаляр | структура) [ length: IDENT ]
```

### Условия

```
Простое:    if x == 1
Путь:       if flags.ack == 1
Enum:       if state == OK
Вложенное:  if a == 1 && b > 5
            if a == 1 || b == 2
```

## Примеры

### Простейший протокол

```
protocol Simple {
    id: 0x1234
    value: uint16
}
```

### Протокол с битовыми полями

```
protocol Flags {
    id: 0x1000
    flags: bitstruct {
        ack: bit(7)
        error: bit(6)
        priority: bits[5:4]
        reserved: bits[3:0]
    }
}
```

### Протокол с enum

```
protocol Status {
    id: 0x8000
    state: enum {
        OK = 0
        ERROR = 1
    }
    value: uint32 if state == OK
}
```

### Протокол с массивами и константами

```
protocol Arrays {
    id: 0x2000
    const MAX = 10
    readings: [MAX]float32
    points: [5]struct {
        x: float32
        y: float32
    }
    data_len: uint16
    data: []uint8 length: data_len
}
```

### Протокол с вложенными условиями

```
protocol Conditional {
    id: 0x3000
    flags: bitstruct {
        ack: bit(7)
        error: bit(6)
    }
    count: uint16
    data: bytes length_from: data_len if flags.ack == 1 && count > 5
    data_len: uint16
}
```

### Протокол с LittleEndian и алиасами

```
protocol LEData {
    id: 0x9000
    endian: little
    alias ID: uint32
    user_id: ID
    value: uint32
}
```

## Документация

- [Туториал](tutorial.md) — пошаговое руководство
- [Архитектура парсера](parser.md) — описание компонентов