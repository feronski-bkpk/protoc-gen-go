protocol TcpHeader {
    id: 0x2002
    src_port: uint16
    dst_port: uint16
    seq_num: uint32
    ack_num: uint32
    data_offset_flags: bitstruct {
        data_offset: bits[7:4]
        reserved: bits[3:0]
    }
    flags: bitstruct {
        cwr: bit(7)
        ece: bit(6)
        urg: bit(5)
        ack: bit(4)
        psh: bit(3)
        rst: bit(2)
        syn: bit(1)
        fin: bit(0)
    }
    window: uint16
    checksum: uint16
    urgent_ptr: uint16
    options_len: uint16
    options: bytes length_from: options_len if data_offset_flags.data_offset > 5
}
