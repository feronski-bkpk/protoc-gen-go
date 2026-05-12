protocol Ipv6HopByHop {
    id: 0x6002
    endian: big
    next_header: uint8
    hdr_ext_len: uint8
    options_len: uint16
    options: bytes length_from: options_len
}
