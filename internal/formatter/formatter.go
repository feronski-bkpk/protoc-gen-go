package formatter

import (
	"fmt"
	"strings"

	"github.com/feronski-bkpk/protoc-gen-go/internal/ast"
	"github.com/feronski-bkpk/protoc-gen-go/internal/parser"
)

func Format(input string) (string, error) {
	proto, err := parser.ParseString(input)
	if err != nil {
		return "", err
	}

	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("protocol %s {\n", proto.Name))
	buf.WriteString(fmt.Sprintf("    id: 0x%04X\n", proto.PacketID))

	if proto.Endian != "" && proto.Endian != "big" {
		buf.WriteString(fmt.Sprintf("    endian: %s\n", proto.Endian))
	}

	if len(proto.Aliases) > 0 {
		for name, baseType := range proto.Aliases {
			buf.WriteString(fmt.Sprintf("    alias %s: %s\n", name, baseType))
		}
	}

	if len(proto.Consts) > 0 {
		for name, val := range proto.Consts {
			buf.WriteString(fmt.Sprintf("    const %s = %d\n", name, val))
		}
	}

	formatFields(&buf, proto.Fields, "    ")

	buf.WriteString("}\n")

	return buf.String(), nil
}

func formatFields(buf *strings.Builder, fields []ast.Field, indent string) {
	for _, field := range fields {
		switch f := field.(type) {
		case *ast.ScalarField:
			typeName := string(f.Type)
			if f.OriginalType != "" && f.OriginalType != typeName {
				typeName = f.OriginalType
			}
			buf.WriteString(fmt.Sprintf("%s%s: %s", indent, f.Name, typeName))
			writeCondition(buf, f.Condition)
			buf.WriteString("\n")

		case *ast.StructField:
			buf.WriteString(fmt.Sprintf("%s%s: struct {\n", indent, f.Name))
			formatFields(buf, f.Struct.Fields, indent+"    ")
			buf.WriteString(fmt.Sprintf("%s}", indent))
			writeCondition(buf, f.Condition)
			buf.WriteString("\n")

		case *ast.BitStructField:
			buf.WriteString(fmt.Sprintf("%s%s: bitstruct {\n", indent, f.Name))
			for _, bit := range f.Fields {
				if bit.IsRange {
					buf.WriteString(fmt.Sprintf("%s    %s: bits[%d:%d]\n", indent, bit.Name, bit.HighBit, bit.LowBit))
				} else {
					buf.WriteString(fmt.Sprintf("%s    %s: bit(%d)\n", indent, bit.Name, bit.Bit))
				}
			}
			buf.WriteString(fmt.Sprintf("%s}\n", indent))

		case *ast.BytesField:
			buf.WriteString(fmt.Sprintf("%s%s: bytes", indent, f.Name))
			if f.LengthFrom != "" {
				buf.WriteString(fmt.Sprintf(" length_from: %s", f.LengthFrom))
			}
			writeCondition(buf, f.Condition)
			buf.WriteString("\n")

		case *ast.EnumField:
			buf.WriteString(fmt.Sprintf("%s%s: enum {\n", indent, f.Name))
			if f.Enum != nil {
				for _, val := range f.Enum.Values {
					buf.WriteString(fmt.Sprintf("%s    %s = %d\n", indent, val.Name, val.Value))
				}
			}
			buf.WriteString(fmt.Sprintf("%s}\n", indent))

		case *ast.ArrayField:
			if f.FixedLength > 0 {
				buf.WriteString(fmt.Sprintf("%s%s: [%d]", indent, f.Name, f.FixedLength))
			} else {
				buf.WriteString(fmt.Sprintf("%s%s: []", indent, f.Name))
			}
			switch elem := f.ElementType.(type) {
			case *ast.ScalarField:
				typeName := string(elem.Type)
				if elem.OriginalType != "" && elem.OriginalType != typeName {
					typeName = elem.OriginalType
				}
				buf.WriteString(typeName)
			case *ast.StructField:
				buf.WriteString("struct {\n")
				formatFields(buf, elem.Struct.Fields, indent+"        ")
				buf.WriteString(fmt.Sprintf("%s    }", indent))
			}
			if f.LengthFrom != "" {
				buf.WriteString(fmt.Sprintf(" length: %s", f.LengthFrom))
			}
			writeCondition(buf, f.Condition)
			buf.WriteString("\n")
		}
	}
}

func writeCondition(buf *strings.Builder, cond *ast.Condition) {
	if cond == nil {
		return
	}
	buf.WriteString(fmt.Sprintf(" if %s %s ", cond.Field, cond.Operator))
	if cond.EnumValue != "" {
		buf.WriteString(cond.EnumValue)
	} else {
		buf.WriteString(fmt.Sprintf("%d", cond.Value))
	}
}
