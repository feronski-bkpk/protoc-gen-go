#!/bin/bash

BIN="./bin/protoc-gen-go"
GREEN="\033[0;32m"
BLUE="\033[0;34m"
YELLOW="\033[1;33m"
NC="\033[0m"

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     ДЕМОНСТРАЦИЯ РЕАЛЬНЫХ ПРОТОКОЛОВ                       ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# MQTT
echo -e "${YELLOW}━━━ MQTT CONNECT (IoT протокол) ━━━${NC}"
echo ""
echo "DSL описание:"
cat examples/mqtt/connect.dsl
echo ""
echo "Генерация..."
$BIN examples/mqtt/connect.dsl 2>/dev/null
echo ""
echo "Сгенерированная структура:"
grep -A 15 "type MQTTConnect struct" examples/mqtt/connect.gen.go
echo ""
echo "Размер сгенерированного файла: $(wc -c < examples/mqtt/connect.gen.go) байт"
echo ""

# Modbus
echo -e "${YELLOW}━━━ Modbus RTU (промышленный протокол) ━━━${NC}"
echo ""
echo "DSL описание:"
cat examples/modbus/rtu.dsl
echo ""
echo "Генерация..."
$BIN examples/modbus/rtu.dsl 2>/dev/null
echo ""
echo "Сгенерированная структура:"
grep -A 8 "type ModbusRTU struct" examples/modbus/rtu.gen.go
echo ""
echo "Условное поле (вложенное условие ||):"
grep -B 2 "if p.Function_code == 3 || p.Function_code == 16" examples/modbus/rtu.gen.go | head -5
echo ""

# TCP
echo -e "${YELLOW}━━━ TCP Header (сетевой протокол) ━━━${NC}"
echo ""
echo "DSL описание:"
head -20 examples/tcp/header.dsl
echo "..."
echo ""
echo "Генерация..."
$BIN examples/tcp/header.dsl 2>/dev/null
echo ""
echo "Битовые поля (два bitstruct):"
grep -A 6 "GetData_offset\|GetAck\|GetSyn" examples/tcp/header.gen.go | head -15
echo ""

# Ethernet
echo -e "${YELLOW}━━━ Ethernet Frame (L2 протокол) ━━━${NC}"
echo ""
echo "DSL описание:"
cat examples/ethernet/frame.dsl
echo ""
echo "Генерация..."
$BIN examples/ethernet/frame.dsl 2>/dev/null
echo ""
echo "Фиксированный массив MAC адресов:"
grep -A 3 "Dst_mac \[6\]uint8" examples/ethernet/frame.gen.go
echo ""

# HTTP
echo -e "${YELLOW}━━━ HTTP Request (прикладной протокол) ━━━${NC}"
echo ""
echo "DSL описание:"
cat examples/http/request.dsl
echo ""
echo "Генерация..."
$BIN examples/http/request.dsl 2>/dev/null
echo ""
echo "Enum и условное поле:"
grep -A 8 "type MethodEnum\|if p.Method == POST || p.Method == PUT" examples/http/request.gen.go | head -12
echo ""

echo -e "${GREEN}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║              ДЕМОНСТРАЦИЯ ЗАВЕРШЕНА                         ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"
