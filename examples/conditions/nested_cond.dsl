protocol NestedCond {
    id: 0xA000
    flags: bitstruct {
        ack: bit(7)
        error: bit(6)
    }
    count: uint16
    data_len: uint16
    data: bytes length_from: data_len if flags.ack == 1 && count > 5
    extended: uint32 if flags.error == 0 || flags.ack == 1
}
