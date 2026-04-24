protocol Status {
    id: 0x5000
    state: enum {
        OK = 0
        ERROR = 1
        PENDING = 2
    }
    value: uint32
}
