protocol AllFeatures {
    id: 0xFFFF
    endian: big
    
    alias ID: uint32
    alias Name: bytes
    
    const MAX_SIZE = 16
    const DEFAULT_TIMEOUT = 30
    
    device_id: ID
    timestamp: uint64
    temperature: float32
    
    flags: bitstruct {
        ack: bit(7)
        error: bit(6)
        priority: bits[5:4]
        reserved: bits[3:0]
    }
    
    state: enum {
        OK = 0
        ERROR = 1
        PENDING = 2
    }
    
    location: struct {
        latitude: float64
        longitude: float64
    }
    
    readings: [MAX_SIZE]float32
    
    name_len: uint16
    name: Name length_from: name_len
    
    samples_len: uint16
    samples: []struct {
        x: float32
        y: float32
        z: float32
    } length: samples_len
    
    error_msg_len: uint16
    error_msg: bytes length_from: error_msg_len if flags.error == 1 && state == ERROR
    
    extended: uint32 if flags.ack == 1 || state == OK
    
    timeout: uint16
}
