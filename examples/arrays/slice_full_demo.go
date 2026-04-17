package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
)

type SamplesElem struct {
	X float32
	Y float32
	Z float32
}

func (p *SamplesElem) Size() int { return 12 }

func (p *SamplesElem) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint32(buf[0:], math.Float32bits(p.X))
	binary.BigEndian.PutUint32(buf[4:], math.Float32bits(p.Y))
	binary.BigEndian.PutUint32(buf[8:], math.Float32bits(p.Z))
	return buf, nil
}

func (p *SamplesElem) UnmarshalBinary(data []byte) error {
	p.X = math.Float32frombits(binary.BigEndian.Uint32(data[0:]))
	p.Y = math.Float32frombits(binary.BigEndian.Uint32(data[4:]))
	p.Z = math.Float32frombits(binary.BigEndian.Uint32(data[8:]))
	return nil
}

type SliceArray struct {
	Device_id    uint32
	Readings_len uint16
	Readings     []float32
	Samples_len  uint8
	Samples      []SamplesElem
	Flags        uint8
}

func (p *SliceArray) GetAck() bool { return (p.Flags & (1 << 7)) != 0 }
func (p *SliceArray) SetAck(v bool) {
	if v {
		p.Flags |= 1 << 7
	} else {
		p.Flags &^= 1 << 7
	}
}
func (p *SliceArray) GetError() bool { return (p.Flags & (1 << 6)) != 0 }
func (p *SliceArray) SetError(v bool) {
	if v {
		p.Flags |= 1 << 6
	} else {
		p.Flags &^= 1 << 6
	}
}

func (p *SliceArray) Size() int {
	size := 0
	size += 4 // Device_id
	size += 2 // Readings_len
	for range p.Readings {
		size += 4
	}
	size += 1 // Samples_len
	for i := range p.Samples {
		size += p.Samples[i].Size()
	}
	size += 1 // Flags
	return size
}

func (p *SliceArray) MarshalBinary() ([]byte, error) {
	buf := make([]byte, p.Size())
	offset := 0
	binary.BigEndian.PutUint32(buf[offset:], p.Device_id)
	offset += 4
	binary.BigEndian.PutUint16(buf[offset:], p.Readings_len)
	offset += 2
	for i := range p.Readings {
		binary.BigEndian.PutUint32(buf[offset:], math.Float32bits(p.Readings[i]))
		offset += 4
	}
	buf[offset] = p.Samples_len
	offset += 1
	for i := range p.Samples {
		data, _ := p.Samples[i].MarshalBinary()
		copy(buf[offset:], data)
		offset += len(data)
	}
	buf[offset] = p.Flags
	return buf, nil
}

func (p *SliceArray) UnmarshalBinary(data []byte) error {
	offset := 0
	p.Device_id = binary.BigEndian.Uint32(data[offset:])
	offset += 4
	p.Readings_len = binary.BigEndian.Uint16(data[offset:])
	offset += 2
	p.Readings = make([]float32, p.Readings_len)
	for i := 0; i < int(p.Readings_len); i++ {
		p.Readings[i] = math.Float32frombits(binary.BigEndian.Uint32(data[offset:]))
		offset += 4
	}
	p.Samples_len = data[offset]
	offset += 1
	p.Samples = make([]SamplesElem, p.Samples_len)
	for i := 0; i < int(p.Samples_len); i++ {
		if err := p.Samples[i].UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		offset += p.Samples[i].Size()
	}
	p.Flags = data[offset]
	return nil
}

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           ПОЛНАЯ ДЕМОНСТРАЦИЯ СЛАЙСОВ                        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	original := SliceArray{
		Device_id:    0x12345678,
		Readings_len: 3,
		Readings:     []float32{23.5, 24.1, 22.8},
		Samples_len:  2,
		Samples: []SamplesElem{
			{X: 1.0, Y: 2.0, Z: 3.0},
			{X: 4.0, Y: 5.0, Z: 6.0},
		},
		Flags: 0,
	}
	original.SetAck(true)

	fmt.Println("📊 ИСХОДНЫЕ ДАННЫЕ:")
	fmt.Printf("  Device ID:     0x%08X\n", original.Device_id)
	fmt.Printf("  Readings_len:  %d\n", original.Readings_len)
	fmt.Printf("  Readings:      %v\n", original.Readings)
	fmt.Printf("  Samples_len:   %d\n", original.Samples_len)
	fmt.Printf("  Samples:       %d элементов\n", len(original.Samples))
	for i, s := range original.Samples {
		fmt.Printf("    [%d] X=%.1f Y=%.1f Z=%.1f\n", i, s.X, s.Y, s.Z)
	}
	fmt.Printf("  Flags:         0x%02X (Ack=%v, Error=%v)\n", original.Flags, original.GetAck(), original.GetError())
	fmt.Println()

	binary, _ := original.MarshalBinary()
	fmt.Println("БИНАРНЫЕ ДАННЫЕ:")
	fmt.Printf("  Размер: %d байт\n", len(binary))
	fmt.Printf("  Hex:    %s\n", hex.EncodeToString(binary))
	fmt.Println()

	var decoded SliceArray
	decoded.UnmarshalBinary(binary)

	fmt.Println("ДЕСЕРИАЛИЗОВАННЫЕ ДАННЫЕ:")
	fmt.Printf("  Device ID:     0x%08X\n", decoded.Device_id)
	fmt.Printf("  Readings_len:  %d\n", decoded.Readings_len)
	fmt.Printf("  Readings:      %v\n", decoded.Readings)
	fmt.Printf("  Samples_len:   %d\n", decoded.Samples_len)
	fmt.Printf("  Samples:       %d элементов\n", len(decoded.Samples))
	for i, s := range decoded.Samples {
		fmt.Printf("    [%d] X=%.1f Y=%.1f Z=%.1f\n", i, s.X, s.Y, s.Z)
	}
	fmt.Printf("  Flags:         0x%02X (Ack=%v, Error=%v)\n", decoded.Flags, decoded.GetAck(), decoded.GetError())
	fmt.Println()

	fmt.Println("ПРОВЕРКА:")
	allMatch := original.Device_id == decoded.Device_id &&
		original.Readings_len == decoded.Readings_len &&
		len(original.Readings) == len(decoded.Readings) &&
		original.Samples_len == decoded.Samples_len &&
		len(original.Samples) == len(decoded.Samples) &&
		original.Flags == decoded.Flags

	for i := range original.Readings {
		if original.Readings[i] != decoded.Readings[i] {
			allMatch = false
			break
		}
	}

	for i := range original.Samples {
		if original.Samples[i] != decoded.Samples[i] {
			allMatch = false
			break
		}
	}

	if allMatch {
		fmt.Println("Все поля совпадают! Слайсы работают корректно!")
	} else {
		fmt.Println("Обнаружены расхождения!")
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    ДЕМОНСТРАЦИЯ ЗАВЕРШЕНА                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
}
