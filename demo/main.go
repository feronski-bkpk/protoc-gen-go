package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/feronski-bkpk/protoc-gen-go/demo/protocol"
)

func main() {
	fmt.Println("=== Демонстрация protoc-gen-go ===\n")

	data := protocol.SensorData{
		Device_id:   0x12345678,
		Timestamp:   uint64(time.Now().Unix()),
		Temperature: 23.5,
		Humidity:    65,
		Pressure:    101325,
		Flags:       1,
		Location: protocol.Location{
			Latitude:  55.7558,
			Longitude: 37.6173,
			Altitude:  150,
		},
		Name_len:      7,
		Name:          []byte("Sensor1"),
		Error_msg_len: 12,
		Error_msg:     []byte("Calibration!"),
	}

	fmt.Println("Исходные данные:")
	fmt.Printf("  Device ID:    0x%08X\n", data.Device_id)
	fmt.Printf("  Timestamp:    %d\n", data.Timestamp)
	fmt.Printf("  Temperature:  %.2f C\n", data.Temperature)
	fmt.Printf("  Humidity:     %d %%\n", data.Humidity)
	fmt.Printf("  Pressure:     %d Pa\n", data.Pressure)
	fmt.Printf("  Flags:        0x%02X\n", data.Flags)
	fmt.Printf("  Location:     %.4f, %.4f (alt: %d)\n", data.Location.Latitude, data.Location.Longitude, data.Location.Altitude)
	fmt.Printf("  Name:         %s\n", string(data.Name))
	fmt.Printf("  Error:        %s\n", string(data.Error_msg))
	fmt.Println()

	fmt.Println("Сериализация в бинарный формат...")
	binary, err := data.MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("  Размер:       %d байт\n", len(binary))
	if len(binary) > 32 {
		fmt.Printf("  Hex (первые 32 байта): %s...\n", hex.EncodeToString(binary[:32]))
	} else {
		fmt.Printf("  Hex: %s\n", hex.EncodeToString(binary))
	}
	fmt.Println()

	fmt.Println("Десериализация...")
	var decoded protocol.SensorData
	if err := decoded.UnmarshalBinary(binary); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Проверка:")
	allMatch := true
	if data.Device_id != decoded.Device_id {
		fmt.Printf("  ОШИБКА: Device_id не совпадает\n")
		allMatch = false
	}
	if data.Timestamp != decoded.Timestamp {
		fmt.Printf("  ОШИБКА: Timestamp не совпадает\n")
		allMatch = false
	}
	if data.Temperature != decoded.Temperature {
		fmt.Printf("  ОШИБКА: Temperature не совпадает\n")
		allMatch = false
	}
	if data.Location.Latitude != decoded.Location.Latitude {
		fmt.Printf("  ОШИБКА: Location не совпадает\n")
		allMatch = false
	}
	if !bytes.Equal(data.Name, decoded.Name) {
		fmt.Printf("  ОШИБКА: Name не совпадает\n")
		allMatch = false
	}
	if !bytes.Equal(data.Error_msg, decoded.Error_msg) {
		fmt.Printf("  ОШИБКА: Error_msg не совпадает\n")
		allMatch = false
	}

	if allMatch {
		fmt.Println("  Все поля совпадают!")
	}
	fmt.Println()

	fmt.Println("Смещения полей в бинарном формате:")
	fmt.Printf("  device_id:      offset=%d, size=%d\n", protocol.SensorData_Device_id_Offset, protocol.SensorData_Device_id_Size)
	fmt.Printf("  timestamp:      offset=%d, size=%d\n", protocol.SensorData_Timestamp_Offset, protocol.SensorData_Timestamp_Size)
	fmt.Printf("  temperature:    offset=%d, size=%d\n", protocol.SensorData_Temperature_Offset, protocol.SensorData_Temperature_Size)
	fmt.Printf("  location:       offset=%d, size=%d\n", protocol.SensorData_Location_Offset, protocol.SensorData_Location_Size)

	fmt.Println("\n=== Демонстрация завершена ===")
}
