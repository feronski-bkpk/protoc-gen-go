package testdata

import (
	"bytes"
	"testing"

	"github.com/feronski-bkpk/protoc-gen-go/testdata/protocol"
)

func TestFullProtocol_BasicTypes(t *testing.T) {
	original := protocol.FullTest{
		Uint8_val:   255,
		Uint16_val:  65535,
		Float64_val: 3.14159,
	}

	data, err := original.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary failed: %v", err)
	}

	var decoded protocol.FullTest
	if err := decoded.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary failed: %v", err)
	}

	if original.Uint8_val != decoded.Uint8_val {
		t.Errorf("Uint8_val: expected %d, got %d", original.Uint8_val, decoded.Uint8_val)
	}

	t.Logf("Basic types test passed. Data size: %d bytes", len(data))
}

func TestFullProtocol_NestedStruct(t *testing.T) {
	original := protocol.FullTest{
		Nested: protocol.Nested{
			X:        1.234,
			Y:        5.678,
			Name_len: 5,
			Name:     []byte("hello"),
		},
	}

	data, err := original.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary failed: %v", err)
	}

	var decoded protocol.FullTest
	if err := decoded.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary failed: %v", err)
	}

	if !bytes.Equal(original.Nested.Name, decoded.Nested.Name) {
		t.Errorf("Nested.Name: expected %s, got %s", original.Nested.Name, decoded.Nested.Name)
	}

	t.Logf("Nested struct test passed")
}

func TestFullProtocol_ConditionalFields(t *testing.T) {
	original := protocol.FullTest{
		Flags:    1,
		Data_len: 4,
		Data:     []byte("test"),
	}

	data, err := original.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary failed: %v", err)
	}

	var decoded protocol.FullTest
	if err := decoded.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary failed: %v", err)
	}

	if !bytes.Equal(original.Data, decoded.Data) {
		t.Errorf("Data: expected %s, got %s", original.Data, decoded.Data)
	}

	t.Logf("Conditional fields test passed")
}
