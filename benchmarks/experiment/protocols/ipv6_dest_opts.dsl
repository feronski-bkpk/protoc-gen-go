protocol Ipv6DestOpts {
    id: 0x6003
    endian: big
    next_header: uint8
    hdr_ext_len: uint8
    options: [6]uint8
}
