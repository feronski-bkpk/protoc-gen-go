# Грамматика DSL protoc-gen-go

## BNF нотация

```
<protocol>       ::= "protocol" <identifier> "{" <id-field> <field-list> "}"

<id-field>       ::= "id" ":" <hex-number>

<field-list>     ::= { <field> }

<field>          ::= <identifier> ":" <field-type> [ <length-spec> ] [ <condition> ]

<field-type>     ::= <scalar-type>
                   | <struct-type>
                   | <bitstruct-type>
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

<bytes-type>     ::= "bytes"

<array-type>     ::= "[" <number> "]" ( <scalar-type> | <struct-type> )

<slice-type>     ::= "[]" ( <scalar-type> | <struct-type> )

<length-spec>    ::= "length" ":" <identifier>
                   | "length_from" ":" <identifier>

<condition>      ::= "if" <identifier> <operator> <value>

<operator>       ::= "==" | "!=" | "<" | ">" | "<=" | ">="

<value>          ::= <number> | <hex-number>

<identifier>     ::= [a-zA-Z_][a-zA-Z0-9_]*

<number>         ::= [0-9]+

<hex-number>     ::= "0x" [0-9a-fA-F]+

<comment>        ::= "//" { <any-char> } <newline>
```

## Синтаксические диаграммы

### Протокол

```
┌─────────────────────────────────────────────────┐
│ protocol → IDENT "{" → id: HEX → { поле } → "}" │
└─────────────────────────────────────────────────┘
```

### Типы полей

```
скаляр:   IDENT ":" (uint8|uint16|uint32|uint64|int8|int16|int32|int64|float32|float64)
структура: IDENT ":" "struct" "{" { поле } "}"
биты:     IDENT ":" "bitstruct" "{" { IDENT ":" bit(N) | bits[H:L] } "}"
bytes:    IDENT ":" "bytes" [ length: IDENT ]
массив:   IDENT ":" "[" N "]" (скаляр | структура)
слайс:    IDENT ":" "[]" (скаляр | структура) [ length: IDENT ]
```

### Условия

```
if IDENT == NUMBER
if IDENT != NUMBER
if IDENT < NUMBER
if IDENT > NUMBER
if IDENT <= NUMBER
if IDENT >= NUMBER
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

### Протокол с массивами

```
protocol Arrays {
    id: 0x2000
    readings: [10]float32
    points: [5]struct {
        x: float32
        y: float32
    }
    data_len: uint16
    data: []uint8 length: data_len
}
```

### Протокол с условными полями

```
protocol Conditional {
    id: 0x3000
    flags: uint8
    extended: uint32 if flags == 1
    error_msg: bytes length_from: error_len if flags == 2
    error_len: uint16
}
```