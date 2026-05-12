protocol Ipv6Fragment {
    id: 0x6002
    endian: big
    next_header: uint8
    reserved: uint8
    offset_flags: bitstruct {
        fragment_offset_high: bits[7:3]
        res: bit(2)
        m_flag: bit(1)
    }
    offset_low: uint8
    identification: uint32
}
