package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
)

type PointsElem struct {
	X float32
	Y float32
}

type FixedArray struct {
	Device_id uint32
	Readings  [10]float32
	Points    [5]PointsElem
	Flags     uint8
}

func (p *FixedArray) GetAck() bool { return (p.Flags & (1 << 7)) != 0 }
func (p *FixedArray) SetAck(v bool) {
	if v {
		p.Flags |= 1 << 7
	} else {
		p.Flags &^= 1 << 7
	}
}

func (p *FixedArray) Size() int {
	return 4 + 10*4 + 5*8 + 1
}

func (p *FixedArray) MarshalBinary() ([]byte, error) {
	buf := make([]byte, p.Size())
	offset := 0
	binary.BigEndian.PutUint32(buf[offset:], p.Device_id)
	offset += 4
	for i := 0; i < 10; i++ {
		binary.BigEndian.PutUint32(buf[offset:], math.Float32bits(p.Readings[i]))
		offset += 4
	}
	for i := 0; i < 5; i++ {
		binary.BigEndian.PutUint32(buf[offset:], math.Float32bits(p.Points[i].X))
		offset += 4
		binary.BigEndian.PutUint32(buf[offset:], math.Float32bits(p.Points[i].Y))
		offset += 4
	}
	buf[offset] = p.Flags
	return buf, nil
}

func (p *FixedArray) UnmarshalBinary(data []byte) error {
	offset := 0
	p.Device_id = binary.BigEndian.Uint32(data[offset:])
	offset += 4
	for i := 0; i < 10; i++ {
		p.Readings[i] = math.Float32frombits(binary.BigEndian.Uint32(data[offset:]))
		offset += 4
	}
	for i := 0; i < 5; i++ {
		p.Points[i].X = math.Float32frombits(binary.BigEndian.Uint32(data[offset:]))
		offset += 4
		p.Points[i].Y = math.Float32frombits(binary.BigEndian.Uint32(data[offset:]))
		offset += 4
	}
	p.Flags = data[offset]
	return nil
}

func main() {
	fmt.Println("=== ДЕМОНСТРАЦИЯ ФИКСИРОВАННЫХ МАССИВОВ ===\n")

	original := FixedArray{
		Device_id: 0x12345678,
		Readings:  [10]float32{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0},
		Points: [5]PointsElem{
			{X: 1, Y: 2},
			{X: 3, Y: 4},
			{X: 5, Y: 6},
			{X: 7, Y: 8},
			{X: 9, Y: 10},
		},
		Flags: 0,
	}
	original.SetAck(true)

	fmt.Printf("Device ID: 0x%08X\n", original.Device_id)
	fmt.Printf("Readings: %v\n", original.Readings[:3])
	fmt.Printf("Points: %v\n", original.Points[:2])
	fmt.Printf("Flags: 0x%02X (Ack=%v)\n", original.Flags, original.GetAck())

	binary, _ := original.MarshalBinary()
	fmt.Printf("\nБинарные данные (%d байт): %s...\n", len(binary), hex.EncodeToString(binary)[:32])

	var decoded FixedArray
	decoded.UnmarshalBinary(binary)

	if original.Device_id == decoded.Device_id && original.Readings == decoded.Readings && original.Points == decoded.Points {
		fmt.Println("\nВсе поля совпадают! Фиксированные массивы работают корректно!")
	}

	fmt.Println("\n=== ДЕМО ЗАВЕРШЕНО ===")
}
