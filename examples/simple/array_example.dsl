protocol SensorData {
    id: 0x2000
    sensor_id: uint16
    samples_count: uint16
    samples: []struct {
        timestamp: uint32
        value: float32
    } length_from: samples_count
}
