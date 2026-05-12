from construct import Struct, Float32b, Int32ub, Int16ub, Array, this

SensorData = Struct(
    "device_id" / Int32ub,
    "temperature" / Float32b,
    "humidity" / Float32b,
    "readings_len" / Int16ub,
    "readings" / Array(10, Float32b),
)
