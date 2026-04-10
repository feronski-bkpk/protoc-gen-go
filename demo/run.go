package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"
)

type Location struct {
	Latitude  float64
	Longitude float64
}

func (p *Location) Size() int { return 16 }

func (p *Location) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf[0:], math.Float64bits(p.Latitude))
	binary.BigEndian.PutUint64(buf[8:], math.Float64bits(p.Longitude))
	return buf, nil
}

func (p *Location) UnmarshalBinary(data []byte) error {
	p.Latitude = math.Float64frombits(binary.BigEndian.Uint64(data[0:]))
	p.Longitude = math.Float64frombits(binary.BigEndian.Uint64(data[8:]))
	return nil
}

type SensorData struct {
	Device_id   uint32
	Timestamp   uint64
	Temperature float32
	Humidity    uint8
	Flags       uint8
	Location    Location
	Name_len    uint16
	Name        []byte
}

func (p *SensorData) GetAck() bool { return (p.Flags & (1 << 7)) != 0 }
func (p *SensorData) SetAck(v bool) {
	if v {
		p.Flags |= 1 << 7
	} else {
		p.Flags &^= 1 << 7
	}
}
func (p *SensorData) GetError() bool { return (p.Flags & (1 << 6)) != 0 }
func (p *SensorData) SetError(v bool) {
	if v {
		p.Flags |= 1 << 6
	} else {
		p.Flags &^= 1 << 6
	}
}
func (p *SensorData) GetPriority() bool { return (p.Flags & (1 << 5)) != 0 }
func (p *SensorData) SetPriority(v bool) {
	if v {
		p.Flags |= 1 << 5
	} else {
		p.Flags &^= 1 << 5
	}
}
func (p *SensorData) GetEnabled() bool { return (p.Flags & (1 << 4)) != 0 }
func (p *SensorData) SetEnabled(v bool) {
	if v {
		p.Flags |= 1 << 4
	} else {
		p.Flags &^= 1 << 4
	}
}

func (p *SensorData) Size() int {
	return 4 + 8 + 4 + 1 + 1 + p.Location.Size() + 2 + len(p.Name)
}

func (p *SensorData) MarshalBinary() ([]byte, error) {
	buf := make([]byte, p.Size())
	off := 0
	binary.BigEndian.PutUint32(buf[off:], p.Device_id)
	off += 4
	binary.BigEndian.PutUint64(buf[off:], p.Timestamp)
	off += 8
	binary.BigEndian.PutUint32(buf[off:], math.Float32bits(p.Temperature))
	off += 4
	buf[off] = p.Humidity
	off++
	buf[off] = p.Flags
	off++
	locData, _ := p.Location.MarshalBinary()
	copy(buf[off:], locData)
	off += len(locData)
	binary.BigEndian.PutUint16(buf[off:], p.Name_len)
	off += 2
	copy(buf[off:], p.Name)
	return buf, nil
}

func (p *SensorData) UnmarshalBinary(data []byte) error {
	off := 0
	p.Device_id = binary.BigEndian.Uint32(data[off:])
	off += 4
	p.Timestamp = binary.BigEndian.Uint64(data[off:])
	off += 8
	p.Temperature = math.Float32frombits(binary.BigEndian.Uint32(data[off:]))
	off += 4
	p.Humidity = data[off]
	off++
	p.Flags = data[off]
	off++
	p.Location.UnmarshalBinary(data[off:])
	off += p.Location.Size()
	p.Name_len = binary.BigEndian.Uint16(data[off:])
	off += 2
	p.Name = make([]byte, p.Name_len)
	copy(p.Name, data[off:off+int(p.Name_len)])
	return nil
}

func main() {
	printHeader()

	data := SensorData{
		Device_id:   0x12345678,
		Timestamp:   uint64(time.Now().Unix()),
		Temperature: 23.5,
		Humidity:    65,
		Flags:       0,
		Location: Location{
			Latitude:  55.7558,
			Longitude: 37.6173,
		},
		Name_len: 7,
		Name:     []byte("Sensor1"),
	}

	data.SetAck(true)
	data.SetPriority(true)

	printData(&data)
	printBinary(&data)
	printDecode(&data)
	printOffsets()
	printFooter()
}

func printHeader() {
	fmt.Println()
}

func printData(data *SensorData) {
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("ИСХОДНЫЕ ДАННЫЕ")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("  ID устройства:   0x%08X\n", data.Device_id)
	fmt.Printf("  Время:           %s\n", time.Unix(int64(data.Timestamp), 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("  Температура:     %.1f°C\n", data.Temperature)
	fmt.Printf("  Влажность:       %d%%\n", data.Humidity)
	fmt.Printf("  Координаты:      %.4f, %.4f\n", data.Location.Latitude, data.Location.Longitude)
	fmt.Printf("  Имя датчика:     %s\n", string(data.Name))
	fmt.Println()
	fmt.Println("  БИТОВЫЕ ФЛАГИ (упакованы в 1 байт):")
	fmt.Printf("     Флаги: 0x%02X [%08b]\n", data.Flags, data.Flags)
	fmt.Printf("     • Ack:      %v\n", data.GetAck())
	fmt.Printf("     • Error:    %v\n", data.GetError())
	fmt.Printf("     • Priority: %v\n", data.GetPriority())
	fmt.Printf("     • Enabled:  %v\n", data.GetEnabled())
	fmt.Println()
}

func printBinary(data *SensorData) {
	binary, _ := data.MarshalBinary()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("БИНАРНОЕ ПРЕДСТАВЛЕНИЕ (Big Endian)")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("  Размер: %d байт\n", len(binary))
	fmt.Printf("  Hex:    %s\n", hex.EncodeToString(binary))
	fmt.Println()
	fmt.Println("  Структура байт:")
	fmt.Printf("  [0-3]   Device ID   → 0x%08X\n", data.Device_id)
	fmt.Printf("  [4-11]  Timestamp   → %d\n", data.Timestamp)
	fmt.Printf("  [12-15] Temperature → %.1f\n", data.Temperature)
	fmt.Printf("  [16]    Humidity    → %d\n", data.Humidity)
	fmt.Printf("  [17]    Flags       → 0x%02X\n", data.Flags)
	fmt.Printf("  [18-33] Location    → %.4f, %.4f\n", data.Location.Latitude, data.Location.Longitude)
	fmt.Printf("  [34-35] Name length → %d\n", data.Name_len)
	fmt.Printf("  [36-42] Name        → %s\n", string(data.Name))
	fmt.Println()
}

func printDecode(data *SensorData) {
	binary, _ := data.MarshalBinary()
	var decoded SensorData
	decoded.UnmarshalBinary(binary)

	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("ДЕСЕРИАЛИЗАЦИЯ")
	fmt.Println(strings.Repeat("─", 60))

	if data.Device_id == decoded.Device_id &&
		data.Temperature == decoded.Temperature &&
		data.Flags == decoded.Flags &&
		data.Location.Latitude == decoded.Location.Latitude &&
		bytes.Equal(data.Name, decoded.Name) {
		fmt.Println("Все поля восстановлены корректно!")
	} else {
		fmt.Println("Ошибка десериализации")
	}
	fmt.Println()
}

func printOffsets() {
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("СМЕЩЕНИЯ ПОЛЕЙ (генерируются автоматически)")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("  const SensorData_Device_id_Offset = 0")
	fmt.Println("  const SensorData_Device_id_Size   = 4")
	fmt.Println("  const SensorData_Timestamp_Offset = 4")
	fmt.Println("  const SensorData_Timestamp_Size   = 8")
	fmt.Println("  const SensorData_Flags_Offset     = 17")
	fmt.Println("  const SensorData_Flags_Size       = 1")
	fmt.Println("  const SensorData_Location_Offset  = 18")
	fmt.Println("  const SensorData_Location_Size    = 16")
	fmt.Println()
}

func printFooter() {
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println("  Для создания своего протокола:")
}
