package main

import (
    "encoding/binary"
    "fmt"
)

type LittleEndianData struct {
    Value uint32
    Count uint16
}

func (p *LittleEndianData) Size() int { return 6 }

func (p *LittleEndianData) MarshalBinary() ([]byte, error) {
    buf := make([]byte, 6)
    binary.LittleEndian.PutUint32(buf[0:], p.Value)
    binary.LittleEndian.PutUint16(buf[4:], p.Count)
    return buf, nil
}

func (p *LittleEndianData) UnmarshalBinary(data []byte) error {
    p.Value = binary.LittleEndian.Uint32(data[0:])
    p.Count = binary.LittleEndian.Uint16(data[4:])
    return nil
}

func main() {
    fmt.Println("=== ДЕМО LITTLE ENDIAN ===\n")
    
    original := LittleEndianData{
        Value: 0x12345678,
        Count: 0xABCD,
    }
    
    data, _ := original.MarshalBinary()
    fmt.Printf("Value: 0x%08X\n", original.Value)
    fmt.Printf("Count: 0x%04X\n", original.Count)
    fmt.Printf("Бинарные данные (LE): %x\n", data)
    
    // BigEndian для сравнения
    be := make([]byte, 6)
    binary.BigEndian.PutUint32(be[0:], original.Value)
    binary.BigEndian.PutUint16(be[4:], original.Count)
    fmt.Printf("Бинарные данные (BE): %x\n", be)
    
    var decoded LittleEndianData
    decoded.UnmarshalBinary(data)
    fmt.Printf("\nВосстановлено: Value=0x%08X, Count=0x%04X\n", decoded.Value, decoded.Count)
    
    if original == decoded {
        fmt.Println("\n✅ LittleEndian работает корректно!")
    }
    
    fmt.Println("\n=== ДЕМО ЗАВЕРШЕНО ===")
}
