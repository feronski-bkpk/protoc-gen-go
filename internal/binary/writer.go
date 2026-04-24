package binary

import (
	"bytes"
	"encoding/binary"

	"github.com/feronski-bkpk/protoc-gen-go/internal/ast"
)

// WriteProtocol сериализует протокол в бинарный формат
func WriteProtocol(proto *ast.Protocol) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString(Magic)                                          // 4 байта
	binary.Write(&buf, binary.BigEndian, uint16(Version))           // 2 байта
	binary.Write(&buf, binary.BigEndian, proto.PacketID)            // 2 байта
	binary.Write(&buf, binary.BigEndian, uint16(len(proto.Fields))) // 2 байта
	binary.Write(&buf, binary.BigEndian, uint16(0))                 // Flags
	buf.Write(make([]byte, 4))                                      // Reserved

	writeString(&buf, proto.Name)

	for _, field := range proto.Fields {
		if err := writeField(&buf, field); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func writeField(buf *bytes.Buffer, field ast.Field) error {
	switch f := field.(type) {
	case *ast.ScalarField:
		buf.WriteByte(TypeScalar)
		writeString(buf, f.Name)
		writeString(buf, string(f.Type))
		binary.Write(buf, binary.BigEndian, uint16(f.GetSize()))

	case *ast.StructField:
		buf.WriteByte(TypeStruct)
		writeString(buf, f.Name)
		binary.Write(buf, binary.BigEndian, uint16(len(f.Struct.Fields)))
		for _, sf := range f.Struct.Fields {
			writeField(buf, sf)
		}

	case *ast.BitStructField:
		buf.WriteByte(TypeBitStruct)
		writeString(buf, f.Name)
		binary.Write(buf, binary.BigEndian, uint16(len(f.Fields)))
		for _, bf := range f.Fields {
			writeString(buf, bf.Name)
			if bf.IsRange {
				buf.WriteByte(2) // bits
				buf.WriteByte(byte(bf.HighBit))
				buf.WriteByte(byte(bf.LowBit))
			} else {
				buf.WriteByte(1) // bit
				buf.WriteByte(byte(bf.Bit))
			}
		}

	case *ast.BytesField:
		buf.WriteByte(TypeBytes)
		writeString(buf, f.Name)
		writeString(buf, f.LengthFrom)

	case *ast.ArrayField:
		if f.FixedLength > 0 {
			buf.WriteByte(TypeArray)
		} else {
			buf.WriteByte(TypeSlice)
		}
		writeString(buf, f.Name)
		binary.Write(buf, binary.BigEndian, uint16(f.FixedLength))
		writeString(buf, f.LengthFrom)
		writeField(buf, f.ElementType)
	}

	if err := writeCondition(buf, field); err != nil {
		return err
	}

	return nil
}

func writeCondition(buf *bytes.Buffer, field ast.Field) error {
	var cond *ast.Condition
	switch f := field.(type) {
	case *ast.ScalarField:
		cond = f.Condition
	case *ast.StructField:
		cond = f.Condition
	case *ast.BytesField:
		cond = f.Condition
	case *ast.ArrayField:
		cond = f.Condition
	}

	if cond != nil {
		buf.WriteByte(1)
		writeString(buf, cond.Field)
		writeString(buf, cond.Operator)
		binary.Write(buf, binary.BigEndian, cond.Value)
	} else {
		buf.WriteByte(0)
	}

	return nil
}

func writeString(buf *bytes.Buffer, s string) {
	binary.Write(buf, binary.BigEndian, uint16(len(s)))
	buf.WriteString(s)
}
