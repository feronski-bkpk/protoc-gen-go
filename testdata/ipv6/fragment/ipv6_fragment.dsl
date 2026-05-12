protocol Ipv6Fragment {
    id: 0x6003
    endian: big
    next_header: uint8
    reserved: uint8
    fragment_offset_flags: bitstruct {
        fragment_offset_high: bits[7:3]
        res: bit(2)
        m_flag: bit(1)
    }
    fragment_offset_low: uint8
    identification: uint32
}
