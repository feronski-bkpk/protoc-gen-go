protocol EthernetFrame {
    id: 0x4000
    dst_mac: [6]uint8
    src_mac: [6]uint8
    ethertype: uint16
    payload_len: uint16
    payload: bytes length_from: payload_len
    crc: uint32
}
