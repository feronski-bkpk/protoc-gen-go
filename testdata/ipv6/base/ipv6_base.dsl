protocol Ipv6BaseHeader {
    id: 0x6001
    endian: big
    version_tc: bitstruct {
        version: bits[7:4]
        traffic_class_high: bits[3:0]
    }
    tc_flow: bitstruct {
        traffic_class_low: bits[7:4]
        flow_label_high: bits[3:0]
    }
    flow_label_low: uint16
    payload_length: uint16
    next_header: uint8
    hop_limit: uint8
    src_addr: [16]uint8
    dst_addr: [16]uint8
}
