protocol Ipv6Routing {
    id: 0x6004
    endian: big
    next_header: uint8
    hdr_ext_len: uint8
    routing_type: uint8
    segments_left: uint8
    reserved: [4]uint8
}
