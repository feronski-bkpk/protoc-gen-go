protocol ConditionalPath {
    id: 0xD000
    flags: bitstruct {
        ack: bit(7)
        error: bit(6)
    }
    data_len: uint16
    data: bytes length_from: data_len if flags.ack == 1
    extended: uint32 if flags.error == 0
}
