protocol SensorFlags {
    id: 0x5000
    
    sensor_id: uint16
    value: float32
    
    flags: bitstruct {
        ack: bit(7)
        error: bit(6)
        ready: bit(5)
        enabled: bit(4)
    }
    
    temperature: int16
}
