package main

import (
	"encoding/binary"
	"fmt"
)

type ConditionalPath struct {
	Flags    uint8
	Data_len uint16
	Data     []byte
	Extended uint32
}

func (p *ConditionalPath) GetAck() bool   { return (p.Flags & (1 << 7)) != 0 }
func (p *ConditionalPath) GetError() bool { return (p.Flags & (1 << 6)) != 0 }
func (p *ConditionalPath) SetAck(v bool) {
	if v {
		p.Flags |= 1 << 7
	} else {
		p.Flags &^= 1 << 7
	}
}
func (p *ConditionalPath) SetError(v bool) {
	if v {
		p.Flags |= 1 << 6
	} else {
		p.Flags &^= 1 << 6
	}
}

func (p *ConditionalPath) Size() int {
	size := 1 + 2 + 4
	if p.GetAck() {
		size += len(p.Data)
	}
	return size
}

func (p *ConditionalPath) MarshalBinary() ([]byte, error) {
	buf := make([]byte, p.Size())
	offset := 0
	buf[offset] = p.Flags
	offset++
	binary.BigEndian.PutUint16(buf[offset:], p.Data_len)
	offset += 2
	if p.GetAck() {
		copy(buf[offset:], p.Data)
		offset += len(p.Data)
	}
	binary.BigEndian.PutUint32(buf[offset:], p.Extended)
	return buf, nil
}

func main() {
	fmt.Println("=== ДЕМО УСЛОВИЙ С ПУТЯМИ ===\n")

	// Тест 1: Ack=true → data присутствует
	p1 := ConditionalPath{
		Flags:    0,
		Data_len: 4,
		Data:     []byte("test"),
		Extended: 0xDEADBEEF,
	}
	p1.SetAck(true)
	p1.SetError(false)

	fmt.Println("Тест 1: Ack=true, Error=false")
	fmt.Printf("  Flags: 0x%02X (Ack=%v, Error=%v)\n", p1.Flags, p1.GetAck(), p1.GetError())
	fmt.Printf("  Data присутствует: %v\n", len(p1.Data) > 0)
	fmt.Printf("  Размер: %d байт\n", p1.Size())

	data1, _ := p1.MarshalBinary()
	fmt.Printf("  Бинарные данные: %x\n\n", data1)

	// Тест 2: Ack=false → data отсутствует
	p2 := ConditionalPath{
		Flags:    0,
		Data_len: 0,
		Extended: 0xCAFEBABE,
	}
	p2.SetAck(false)
	p2.SetError(false)

	fmt.Println("Тест 2: Ack=false, Error=false")
	fmt.Printf("  Flags: 0x%02X (Ack=%v, Error=%v)\n", p2.Flags, p2.GetAck(), p2.GetError())
	fmt.Printf("  Data присутствует: %v\n", len(p2.Data) > 0)
	fmt.Printf("  Размер: %d байт\n", p2.Size())

	data2, _ := p2.MarshalBinary()
	fmt.Printf("  Бинарные данные: %x\n\n", data2)

	// Сравнение размеров
	fmt.Printf("Размер с data: %d байт\n", len(data1))
	fmt.Printf("Размер без data: %d байт\n", len(data2))
	fmt.Printf("Экономия: %d байт\n\n", len(data1)-len(data2))

	fmt.Println("✅ Условные поля с путями работают корректно!")
}
