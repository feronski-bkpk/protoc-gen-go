package main

import "fmt"

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

func (p *Status) Size() int { return 5 }

func (p *Status) MarshalBinary() ([]byte, error) {
    buf := make([]byte, 5)
    buf[0] = byte(p.State)
    buf[1] = byte(p.Value >> 24)
    buf[2] = byte(p.Value >> 16)
    buf[3] = byte(p.Value >> 8)
    buf[4] = byte(p.Value)
    return buf, nil
}

func (p *Status) UnmarshalBinary(data []byte) error {
    p.State = StateEnum(data[0])
    p.Value = uint32(data[1])<<24 | uint32(data[2])<<16 | uint32(data[3])<<8 | uint32(data[4])
    return nil
}

func main() {
    fmt.Println("=== ДЕМО ENUM-ТИПОВ ===\n")
    
    // Создаём статус
    s := Status{
        State: OK,
        Value: 42,
    }
    
    fmt.Printf("State: %v (%d)\n", s.State, s.State)
    fmt.Printf("Value: %d\n", s.Value)
    
    // Меняем состояние
    s.State = ERROR
    fmt.Printf("\nПосле изменения:\n")
    fmt.Printf("State: %v (%d)\n", s.State, s.State)
    
    // Сериализация
    data, _ := s.MarshalBinary()
    fmt.Printf("\nБинарные данные: %x\n", data)
    
    // Десериализация
    var s2 Status
    s2.UnmarshalBinary(data)
    fmt.Printf("Восстановлено: State=%v, Value=%d\n", s2.State, s2.Value)
    
    // Все возможные значения
    fmt.Println("\nВсе значения enum:")
    states := []StateEnum{OK, ERROR, PENDING}
    names := []string{"OK", "ERROR", "PENDING"}
    for i, state := range states {
        fmt.Printf("  %s = %d\n", names[i], state)
    }
    
    fmt.Println("\n=== ДЕМО ЗАВЕРШЕНО ===")
}
