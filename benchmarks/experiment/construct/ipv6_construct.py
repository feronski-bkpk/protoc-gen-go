from construct import Struct, Byte, Int16ub, Int32ub, Array, Bytes, this

# IPv6 Base Header (40 байт)
Ipv6BaseHeader = Struct(
    "version_tc" / Byte,
    "tc_flow" / Byte,
    "flow_label_low" / Int16ub,
    "payload_length" / Int16ub,
    "next_header" / Byte,
    "hop_limit" / Byte,
    "src_addr" / Array(16, Byte),
    "dst_addr" / Array(16, Byte),
)

# Расширения (по 8 байт каждое)
Ipv6HopByHop = Struct(
    "next_header" / Byte,
    "hdr_ext_len" / Byte,
    "options" / Array(6, Byte),
)

Ipv6Fragment = Struct(
    "next_header" / Byte,
    "reserved" / Byte,
    "offset_flags" / Byte,
    "offset_low" / Byte,
    "identification" / Int32ub,
)

Ipv6DestOpts = Struct(
    "next_header" / Byte,
    "hdr_ext_len" / Byte,
    "options" / Array(6, Byte),
)

Ipv6Routing = Struct(
    "next_header" / Byte,
    "hdr_ext_len" / Byte,
    "routing_type" / Byte,
    "segments_left" / Byte,
    "reserved" / Array(4, Byte),
)

# Payload
Payload = Struct(
    "data_len" / Int16ub,
    "data" / Bytes(this.data_len),
)

# ============================================================
# Сборка пакета
# ============================================================

def build_ext_header(ext_type, next_hdr):
    """Создаёт байты расширения в зависимости от типа"""
    if ext_type == 0:  # Hop-by-Hop
        obj = {"next_header": next_hdr, "hdr_ext_len": 0, "options": [0]*6}
        return Ipv6HopByHop.build(obj)
    elif ext_type == 44:  # Fragment
        obj = {"next_header": next_hdr, "reserved": 0, "offset_flags": 0, "offset_low": 0, "identification": 1}
        return Ipv6Fragment.build(obj)
    elif ext_type == 60:  # Destination Options
        obj = {"next_header": next_hdr, "hdr_ext_len": 0, "options": [0]*6}
        return Ipv6DestOpts.build(obj)
    elif ext_type == 43:  # Routing
        obj = {"next_header": next_hdr, "hdr_ext_len": 0, "routing_type": 0, "segments_left": 0, "reserved": [0]*4}
        return Ipv6Routing.build(obj)
    return b''


def build_packet(num_ext, payload_size):
    ext_proto = [0, 44, 60, 43]

    # Базовый заголовок
    base = {
        "version_tc": 0x60,  # version=6
        "tc_flow": 0,
        "flow_label_low": 0,
        "hop_limit": 64,
        "src_addr": [0]*15 + [1],  # ::1
        "dst_addr": [0]*15 + [1],  # ::1
        "next_header": ext_proto[0] if num_ext > 0 else 59,
        "payload_length": 0,  # временно
    }

    # Расширения
    ext_headers = []
    header_size = 40

    for i in range(num_ext):
        next_hdr = ext_proto[i+1] if i < num_ext-1 else 59
        hdr_bytes = build_ext_header(ext_proto[i], next_hdr)
        header_size += len(hdr_bytes)
        ext_headers.append(hdr_bytes)

    # Payload
    payload_data = {
        "data_len": payload_size,
        "data": b'\x00' * payload_size,
    }
    payload_bytes = Payload.build(payload_data)

    # Обновляем payload_length
    base["payload_length"] = header_size - 40 + len(payload_bytes)
    base_bytes = Ipv6BaseHeader.build(base)

    # Сборка
    packet = base_bytes
    for ext in ext_headers:
        packet += ext
    packet += payload_bytes

    return packet


def parse_packet(data, num_ext):
    ext_proto = [0, 44, 60, 43]

    base = Ipv6BaseHeader.parse(data[:40])
    offset = 40

    for i in range(num_ext):
        if ext_proto[i] == 0:
            Ipv6HopByHop.parse(data[offset:offset+8])
        elif ext_proto[i] == 44:
            Ipv6Fragment.parse(data[offset:offset+8])
        elif ext_proto[i] == 60:
            Ipv6DestOpts.parse(data[offset:offset+8])
        elif ext_proto[i] == 43:
            Ipv6Routing.parse(data[offset:offset+8])
        offset += 8

    Payload.parse(data[offset:])


# ============================================================
# Бенчмарки
# ============================================================
if __name__ == "__main__":
    import timeit

    def bench_build(name, num_ext, payload_size, number=50000):
        t = timeit.timeit(lambda: build_packet(num_ext, payload_size), number=number)
        ns = t / number * 1e9
        print(f"ConstructBuild_{name}: {ns:.1f} ns/op")

    def bench_parse(name, num_ext, payload_size, number=50000):
        data = build_packet(num_ext, payload_size)
        t = timeit.timeit(lambda: parse_packet(data, num_ext), number=number)
        ns = t / number * 1e9
        print(f"ConstructParse_{name}: {ns:.1f} ns/op")

    print("=== Эксперимент А: Варьирование размера ===")
    bench_build("64B_0ext", 0, 64, 20000)
    bench_parse("64B_0ext", 0, 64, 20000)
    bench_build("256B_2ext", 2, 256, 20000)
    bench_parse("256B_2ext", 2, 256, 20000)
    bench_build("1024B_4ext", 4, 1024, 10000)
    bench_parse("1024B_4ext", 4, 1024, 10000)
    bench_build("4096B_4ext", 4, 4096, 5000)
    bench_parse("4096B_4ext", 4, 4096, 5000)

    print("\n=== Эксперимент Б: Варьирование сложности ===")
    for i in range(5):
        bench_build(f"1024B_Ext{i}", i, 1024, 10000)
        bench_parse(f"1024B_Ext{i}", i, 1024, 10000)
