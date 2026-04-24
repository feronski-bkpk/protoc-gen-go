package binary

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/feronski-bkpk/protoc-gen-go/internal/ast"
)

func ReadProtocol(data []byte) (*ast.Protocol, error) {
	buf := bytes.NewReader(data)

	magic := make([]byte, 4)
	io.ReadFull(buf, magic)
	if string(magic) != Magic {
		return nil, fmt.Errorf("неверная сигнатура: %s", string(magic))
	}

	var version, packetID, fieldCount, flags uint16
	binary.Read(buf, binary.BigEndian, &version)
	binary.Read(buf, binary.BigEndian, &packetID)
	binary.Read(buf, binary.BigEndian, &fieldCount)
	binary.Read(buf, binary.BigEndian, &flags)

	buf.Seek(4, io.SeekCurrent)

	_ = version
	_ = flags

	name := readString(buf)

	proto := &ast.Protocol{
		Name:     name,
		PacketID: packetID,
		Fields:   make([]ast.Field, 0, fieldCount),
		Types:    make(map[string]*ast.StructType),
	}

	for i := uint16(0); i < fieldCount; i++ {
		field, err := readField(buf)
		if err != nil {
			return nil, err
		}
		proto.Fields = append(proto.Fields, field)
	}

	return proto, nil
}

func readField(buf *bytes.Reader) (ast.Field, error) {
	fieldType, _ := buf.ReadByte()
	name := readString(buf)

	switch fieldType {
	case TypeScalar:
		typeName := readString(buf)
		var size uint16
		binary.Read(buf, binary.BigEndian, &size)
		_ = size

		field := &ast.ScalarField{
			Name: name,
			Type: ast.ScalarType(typeName),
		}
		field.Condition = readCondition(buf)
		return field, nil

	case TypeStruct:
		var subFieldCount uint16
		binary.Read(buf, binary.BigEndian, &subFieldCount)

		fields := make([]ast.Field, 0, subFieldCount)
		for j := uint16(0); j < subFieldCount; j++ {
			subField, _ := readField(buf)
			fields = append(fields, subField)
		}

		field := &ast.StructField{
			Name:   name,
			Struct: &ast.StructType{Fields: fields},
		}
		field.Condition = readCondition(buf)
		return field, nil

	case TypeBitStruct:
		var bitCount uint16
		binary.Read(buf, binary.BigEndian, &bitCount)

		bitFields := make([]*ast.BitFieldSpec, 0, bitCount)
		for j := uint16(0); j < bitCount; j++ {
			bitName := readString(buf)
			bitType, _ := buf.ReadByte()
			if bitType == 2 {
				hi, _ := buf.ReadByte()
				lo, _ := buf.ReadByte()
				bitFields = append(bitFields, &ast.BitFieldSpec{
					Name:    bitName,
					HighBit: int(hi),
					LowBit:  int(lo),
					IsRange: true,
				})
			} else {
				bit, _ := buf.ReadByte()
				bitFields = append(bitFields, &ast.BitFieldSpec{
					Name: bitName,
					Bit:  int(bit),
				})
			}
		}

		field := &ast.BitStructField{
			Name:   name,
			Fields: bitFields,
		}
		field.Condition = readCondition(buf)
		return field, nil

	case TypeBytes:
		lengthFrom := readString(buf)
		field := &ast.BytesField{
			Name:       name,
			LengthFrom: lengthFrom,
		}
		field.Condition = readCondition(buf)
		return field, nil

	case TypeArray, TypeSlice:
		var fixedLength uint16
		binary.Read(buf, binary.BigEndian, &fixedLength)
		lengthFrom := readString(buf)

		elemField, err := readField(buf)
		if err != nil {
			return nil, err
		}

		field := &ast.ArrayField{
			Name:        name,
			ElementType: elemField,
			FixedLength: int(fixedLength),
			LengthFrom:  lengthFrom,
		}
		field.Condition = readCondition(buf)
		return field, nil
	}

	return nil, fmt.Errorf("неизвестный тип поля: %d", fieldType)
}

func readCondition(buf *bytes.Reader) *ast.Condition {
	hasCond, _ := buf.ReadByte()
	if hasCond == 0 {
		return nil
	}

	field := readString(buf)
	op := readString(buf)
	var val uint64
	binary.Read(buf, binary.BigEndian, &val)

	return &ast.Condition{
		Field:    field,
		Operator: op,
		Value:    val,
	}
}

func readString(buf *bytes.Reader) string {
	var length uint16
	binary.Read(buf, binary.BigEndian, &length)
	data := make([]byte, length)
	io.ReadFull(buf, data)
	return string(data)
}
