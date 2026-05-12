protocol SensorData {
    id: 0x1001
    device_id: uint32
    temperature: float32
    humidity: float32
    readings_len: uint16
    readings: [10]float32
}
