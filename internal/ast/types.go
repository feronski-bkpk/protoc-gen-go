package ast

type ScalarType string

const (
	UINT8   ScalarType = "uint8"
	UINT16  ScalarType = "uint16"
	UINT32  ScalarType = "uint32"
	UINT64  ScalarType = "uint64"
	INT8    ScalarType = "int8"
	INT16   ScalarType = "int16"
	INT32   ScalarType = "int32"
	INT64   ScalarType = "int64"
	FLOAT32 ScalarType = "float32"
	FLOAT64 ScalarType = "float64"
)

func (s ScalarType) Size() int {
	switch s {
	case UINT8, INT8:
		return 1
	case UINT16, INT16:
		return 2
	case UINT32, INT32, FLOAT32:
		return 4
	case UINT64, INT64, FLOAT64:
		return 8
	default:
		return 0
	}
}

type Protocol struct {
	Name     string
	PacketID uint16
	Fields   []Field
	Types    map[string]*StructType
}

type Field interface {
	GetName() string
	GetType() string
	GetSize() int
}

type ScalarField struct {
	Name      string
	Type      ScalarType
	Offset    int
	Condition *Condition
	Min       *uint64
	Max       *uint64
	Required  bool
}

func (f *ScalarField) GetName() string { return f.Name }
func (f *ScalarField) GetType() string { return string(f.Type) }
func (f *ScalarField) GetSize() int    { return f.Type.Size() }

type StructField struct {
	Name      string
	TypeRef   string
	Struct    *StructType
	Offset    int
	Condition *Condition
}

func (f *StructField) GetName() string { return f.Name }
func (f *StructField) GetType() string { return f.TypeRef }
func (f *StructField) GetSize() int {
	if f.Struct != nil {
		return f.Struct.Size
	}
	return 0
}

type StructType struct {
	Name        string
	Fields      []Field
	Size        int
	IsBitPacked bool
}

type BytesField struct {
	Name       string
	LengthFrom string
	MaxLength  int
	Offset     int
	Condition  *Condition
	Required   bool
}

func (f *BytesField) GetName() string { return f.Name }
func (f *BytesField) GetType() string { return "[]byte" }
func (f *BytesField) GetSize() int    { return 0 }

type Condition struct {
	Field    string
	Operator string
	Value    uint64
}

type BitFieldSpec struct {
	Name    string
	Bit     int
	HighBit int
	LowBit  int
	IsRange bool
}

type BitStructField struct {
	Name      string
	Fields    []*BitFieldSpec
	Condition *Condition
}

func (f *BitStructField) GetName() string { return f.Name }
func (f *BitStructField) GetType() string { return "bitstruct" }
func (f *BitStructField) GetSize() int    { return 1 }

type ArrayField struct {
	Name        string
	ElementType Field
	FixedLength int
	LengthFrom  string
	Condition   *Condition
}

func (f *ArrayField) GetName() string { return f.Name }
func (f *ArrayField) GetType() string { return "[]" + f.ElementType.GetType() }
func (f *ArrayField) GetSize() int {
	if f.FixedLength > 0 {
		return f.ElementType.GetSize() * f.FixedLength
	}
	return 0
}
