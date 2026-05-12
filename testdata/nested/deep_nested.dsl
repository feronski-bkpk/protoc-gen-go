protocol DeepNested {
    id: 0xBEEF
    outer_id: uint16
    inner: struct {
        level1_id: uint32
        deeper: struct {
            level2_flag: uint8
            deepest: struct {
                value: uint64
            }
        }
    }
    trailer: uint16
}
