from construct import Struct, Int8ub, Int16ub, Int32ub, Byte

Ipv4Header = Struct(
    "version_ihl" / Int8ub,
    "dscp_ecn" / Int8ub,
    "total_length" / Int16ub,
    "identification" / Int16ub,
    "flags_fragment" / Int8ub,
    "fragment_offset" / Int8ub,
    "ttl" / Int8ub,
    "proto" / Int8ub,
    "header_checksum" / Int16ub,
    "src_addr" / Int32ub,
    "dst_addr" / Int32ub,
)
