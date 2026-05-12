protocol Ipv4Header {
    id: 0x4004
    version_ihl: bitstruct {
        version: bits[7:4]
        ihl: bits[3:0]
    }
    dscp_ecn: bitstruct {
        dscp: bits[7:2]
        ecn: bits[1:0]
    }
    total_length: uint16
    identification: uint16
    flags_fragment: bitstruct {
        flags: bits[7:5]
        fragment_offset: bits[4:0]
    }
    fragment_offset_cont: uint8
    ttl: uint8
    proto: uint8
    header_checksum: uint16
    src_addr: uint32
    dst_addr: uint32
    options_len: uint16
    options: bytes length_from: options_len if version_ihl.ihl > 5
}
