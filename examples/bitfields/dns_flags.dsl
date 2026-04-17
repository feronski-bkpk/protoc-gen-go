protocol DNSFlags {
    id: 0x0000
    
    tx_id: uint16
    
    flags: bitstruct {
        qr: bit(7)
        opcode: bits[6:3]
        aa: bit(2)
        tc: bit(1)
        rd: bit(0)
    }
    
    flags2: bitstruct {
        ra: bit(7)
        z: bits[6:4]
        rcode: bits[3:0]
    }
    
    qdcount: uint16
    ancount: uint16
    nscount: uint16
    arcount: uint16
}
