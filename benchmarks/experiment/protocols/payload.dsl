protocol Payload {
    id: 0x7000
    endian: big
    data_len: uint16
    data: bytes length_from: data_len
}
