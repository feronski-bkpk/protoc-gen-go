protocol DNSHeader {
    id: 0x0000
    tx_id: uint16
    flags: bitstruct {
        qr: bit(7)
        aa: bit(2)
        tc: bit(1)
        rd: bit(0)
    }
    flags2: bitstruct {
        ra: bit(7)
        rcode0: bit(3)
        rcode1: bit(2)
        rcode2: bit(1)
        rcode3: bit(0)
    }
    qdcount: uint16
    ancount: uint16
    nscount: uint16
    arcount: uint16
}
