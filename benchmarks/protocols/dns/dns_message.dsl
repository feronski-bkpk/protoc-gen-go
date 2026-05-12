protocol DnsMessage {
    id: 0x3003
    transaction_id: uint16
    flags: bitstruct {
        qr: bit(7)
        opcode: bits[6:3]
        aa: bit(2)
        tc: bit(1)
        rd: bit(0)
    }
    flags2: bitstruct {
        ra: bit(7)
        z: bits[6:4]
        rcode: bits[3:0]
    }
    qdcount: uint16
    ancount: uint16
    nscount: uint16
    arcount: uint16

    // Body зависит от счетчиков (упрощенно)
    question_len: uint16
    question: bytes length_from: question_len if qdcount > 0

    answer_len: uint16
    answer: bytes length_from: answer_len if ancount > 0
}
