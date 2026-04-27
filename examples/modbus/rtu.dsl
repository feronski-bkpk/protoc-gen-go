protocol ModbusRTU {
    id: 0x2000
    address: uint8
    function_code: uint8
    data_len: uint16
    data: bytes length_from: data_len if function_code == 3 || function_code == 16
    crc: uint16
}
