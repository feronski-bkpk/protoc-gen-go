protocol TCPHeader {
    id: 0x3000
    src_port: uint16
    dst_port: uint16
    seq_num: uint32
    ack_num: uint32
    
    flags: bitstruct {
        data_offset: bits[7:4]
        reserved: bits[3:1]
        ns: bit(0)
    }
    
    flags2: bitstruct {
        cwr: bit(7)
        ece: bit(6)
        urg: bit(5)
        ack: bit(4)
        psh: bit(3)
        rst: bit(2)
        syn: bit(1)
        fin: bit(0)
    }
    
    window_size: uint16
    checksum: uint16
    urgent_ptr: uint16
    
    options_len: uint8
    options: []uint8 length: options_len if flags2.syn == 1
}
