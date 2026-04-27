protocol MQTTConnect {
    id: 0x1000
    
    // Fixed Header
    flags: bitstruct {
        message_type: bits[7:4]
        dup: bit(3)
        qos: bits[2:1]
        retain: bit(0)
    }
    remaining_len: uint8
    
    // Variable Header
    protocol_name_len: uint16
    protocol_name: bytes length_from: protocol_name_len
    protocol_version: uint8
    
    connect_flags: bitstruct {
        username: bit(7)
        password: bit(6)
        will_retain: bit(5)
        will_qos: bits[4:3]
        will_flag: bit(2)
        clean_session: bit(1)
        reserved: bit(0)
    }
    
    keep_alive: uint16
    
    // Payload
    client_id_len: uint16
    client_id: bytes length_from: client_id_len
}
