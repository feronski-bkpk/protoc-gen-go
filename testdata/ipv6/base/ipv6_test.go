package base

import (
	"encoding/hex"
	"testing"
)

const realIPv6Hex = "60000000000006400000000000000000000000000000000100000000000000000000000000000001"

func TestIPv6BaseRealPacket(t *testing.T) {
	data, err := hex.DecodeString(realIPv6Hex)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) != 40 {
		t.Fatalf("ожидалось 40 байт, получено %d", len(data))
	}

	var hdr Ipv6BaseHeader
	if err := hdr.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}

	version := hdr.GetVersion()
	if version != 6 {
		t.Errorf("version = %d, ожидалось 6", version)
	}

	trafficClassHigh := hdr.GetTraffic_class_high()
	trafficClassLow := hdr.GetTraffic_class_low()
	trafficClass := (trafficClassHigh << 4) | trafficClassLow
	if trafficClass != 0 {
		t.Errorf("traffic_class = 0x%02x, ожидалось 0x00", trafficClass)
	}

	flowLabelHigh := hdr.GetFlow_label_high()
	flowLabel := (uint32(flowLabelHigh) << 16) | uint32(hdr.Flow_label_low)
	if flowLabel != 0 {
		t.Errorf("flow_label = 0x%05x, ожидалось 0x00000", flowLabel)
	}

	if hdr.Payload_length != 0 {
		t.Errorf("payload_length = %d, ожидалось 0", hdr.Payload_length)
	}

	if hdr.Next_header != 6 {
		t.Errorf("next_header = %d, ожидалось 6 (TCP)", hdr.Next_header)
	}

	if hdr.Hop_limit != 64 {
		t.Errorf("hop_limit = %d, ожидалось 64", hdr.Hop_limit)
	}

	expectedSrc := [16]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	expectedDst := [16]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

	if hdr.Src_addr != expectedSrc {
		t.Errorf("src_addr =\n%v\nожидалось\n%v", hdr.Src_addr, expectedSrc)
	}
	if hdr.Dst_addr != expectedDst {
		t.Errorf("dst_addr =\n%v\nожидалось\n%v", hdr.Dst_addr, expectedDst)
	}

	marshaled, err := hdr.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	if len(marshaled) != len(data) {
		t.Fatalf("длина marshaled %d, ожидалось %d", len(marshaled), len(data))
	}

	for i := range data {
		if marshaled[i] != data[i] {
			t.Errorf("байт %d: marshaled 0x%02x != original 0x%02x", i, marshaled[i], data[i])
		}
	}

	var hdr2 Ipv6BaseHeader
	if err := hdr2.UnmarshalBinary(marshaled); err != nil {
		t.Fatalf("UnmarshalBinary roundtrip: %v", err)
	}

	if hdr != hdr2 {
		t.Errorf("roundtrip mismatch:\n  original: %+v\n  restored: %+v", hdr, hdr2)
	}
}

func TestIPv6Size(t *testing.T) {
	hdr := Ipv6BaseHeader{}
	if size := hdr.Size(); size != 40 {
		t.Errorf("Size() = %d, ожидалось 40", size)
	}
}
