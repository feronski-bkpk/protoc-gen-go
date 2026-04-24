protocol TemperatureReading {
    id: 0x1234
    sensor_id: uint16
    temperature: int16
    humidity: uint8
    flags: bitstruct {
        ack: bit(7)
        error: bit(6)
        reserved: bits[5:0]
    }
    location: struct {
        latitude: float64
        longitude: float64
    }
    error_msg_len: uint16
    error_msg: bytes length_from: error_msg_len if flags == 1
    samples_count: uint16
    samples: []struct {
        timestamp: uint32
        value: float32
    } length: samples_count
}
