protocol TemperatureReading {
    id: 0x1234
    sensor_id: uint16
    temperature: int16
    humidity: uint8
    flags: uint8
    location: struct {
        latitude: float64
        longitude: float64
    }
    error_msg_len: uint16
    error_msg: bytes length_from: error_msg_len if flags == 1
    extended_data: uint32 if flags == 2
}
