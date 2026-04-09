protocol SensorData {
    id: 0xABCD
    
    device_id: uint32
    timestamp: uint64
    temperature: float32
    humidity: uint8
    pressure: uint32
    
    flags: uint8
    
    location: struct {
        latitude: float64
        longitude: float64
        altitude: int32
    }
    
    name_len: uint16
    name: bytes length_from: name_len
    
    error_msg_len: uint16
    error_msg: bytes length_from: error_msg_len if flags == 1
    
    calibration_data: uint32 if flags == 2
}
