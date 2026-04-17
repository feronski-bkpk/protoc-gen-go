protocol SliceArray {
    id: 0x3000
    device_id: uint32
    readings_len: uint16
    readings: []float32 length: readings_len
    samples_len: uint8
    samples: []struct {
        x: float32
        y: float32
        z: float32
    } length: samples_len
    flags: uint8
}
