package analyzer

import (
	"fmt"
	"strings"

	"github.com/feronski-bkpk/protoc-gen-go/internal/ast"
)

type Analyzer struct {
	proto       *ast.Protocol
	symbolTable map[string]*SymbolInfo
	errors      []string
}

type SymbolInfo struct {
	Field  ast.Field
	Path   string
	Offset int
	Size   int
}

func NewAnalyzer(proto *ast.Protocol) *Analyzer {
	return &Analyzer{
		proto:       proto,
		symbolTable: make(map[string]*SymbolInfo),
	}
}

func (a *Analyzer) Analyze() error {
	a.buildSymbolTable(a.proto.Fields, "", 0)
	a.computeOffsets(a.proto.Fields, "", 0)
	a.validateReferences()
	a.validateBitFields()
	a.validateCycles()

	if len(a.errors) > 0 {
		return fmt.Errorf("ошибки анализа:\n%s", strings.Join(a.errors, "\n"))
	}

	return nil
}

func (a *Analyzer) buildSymbolTable(fields []ast.Field, parentPath string, baseOffset int) int {
	offset := baseOffset

	for _, field := range fields {
		path := field.GetName()
		if parentPath != "" {
			path = parentPath + "." + path
		}

		info := &SymbolInfo{
			Field:  field,
			Path:   path,
			Offset: offset,
			Size:   field.GetSize(),
		}
		a.symbolTable[path] = info

		switch f := field.(type) {
		case *ast.ScalarField:
			offset += f.GetSize()

		case *ast.StructField:
			innerSize := a.buildSymbolTable(f.Struct.Fields, path, offset)
			info.Size = innerSize
			offset += innerSize

		case *ast.BitStructField:
			info.Size = 1
			for _, bit := range f.Fields {
				bitPath := path + "." + bit.Name
				a.symbolTable[bitPath] = &SymbolInfo{
					Field:  f,
					Path:   bitPath,
					Offset: offset,
					Size:   1,
				}
			}
			offset += 1

		case *ast.ArrayField:
			if f.FixedLength > 0 {
				elemSize := f.ElementType.GetSize()
				info.Size = elemSize * f.FixedLength
				offset += info.Size
			} else {
				info.Size = 0
			}

		case *ast.BytesField:
			info.Size = 0
		}
	}

	return offset - baseOffset
}

func (a *Analyzer) computeOffsets(fields []ast.Field, parentPath string, baseOffset int) {
	offset := baseOffset

	for _, field := range fields {
		path := field.GetName()
		if parentPath != "" {
			path = parentPath + "." + path
		}

		if info, ok := a.symbolTable[path]; ok {
			info.Offset = offset
		}

		switch f := field.(type) {
		case *ast.ScalarField:
			f.Offset = offset
			offset += f.GetSize()

		case *ast.StructField:
			f.Offset = offset
			if f.Struct != nil && f.Struct.Fields != nil {
				innerSize := a.symbolTable[path].Size
				if innerSize == 0 {
					innerSize = a.computeStructSize(f.Struct.Fields)
					a.symbolTable[path].Size = innerSize
				}
				a.computeOffsets(f.Struct.Fields, path, offset)
				offset += innerSize
			} else {
				offset += f.GetSize()
			}

		case *ast.BitStructField:
			offset += 1

		case *ast.ArrayField:
			if f.FixedLength > 0 {
				offset += f.ElementType.GetSize() * f.FixedLength
			}

		case *ast.BytesField:
		}
	}
}

func (a *Analyzer) computeStructSize(fields []ast.Field) int {
	size := 0
	for _, field := range fields {
		switch f := field.(type) {
		case *ast.ScalarField:
			size += f.GetSize()
		case *ast.StructField:
			if f.Struct != nil {
				size += a.computeStructSize(f.Struct.Fields)
			}
		case *ast.BitStructField:
			size += 1
		case *ast.ArrayField:
			if f.FixedLength > 0 {
				size += f.ElementType.GetSize() * f.FixedLength
			}
		}
	}
	return size
}

func (a *Analyzer) validateReferences() {
	for _, info := range a.symbolTable {
		switch f := info.Field.(type) {
		case *ast.BytesField:
			if f.LengthFrom != "" {
				if _, exists := a.symbolTable[f.LengthFrom]; !exists {
					a.errors = append(a.errors,
						fmt.Sprintf("поле '%s': length_from ссылается на несуществующее поле '%s'",
							info.Path, f.LengthFrom))
				}
			}

		case *ast.ArrayField:
			if f.LengthFrom != "" {
				if _, exists := a.symbolTable[f.LengthFrom]; !exists {
					a.errors = append(a.errors,
						fmt.Sprintf("поле '%s': length ссылается на несуществующее поле '%s'",
							info.Path, f.LengthFrom))
				}
			}

		case *ast.ScalarField:
			if f.Condition != nil {
				condField := f.Condition.Field
				if _, exists := a.symbolTable[condField]; !exists {
					parts := strings.Split(condField, ".")
					if len(parts) > 0 {
						if _, exists := a.symbolTable[parts[0]]; !exists {
							a.errors = append(a.errors,
								fmt.Sprintf("поле '%s': условие ссылается на несуществующее поле '%s'",
									info.Path, condField))
						}
					} else {
						a.errors = append(a.errors,
							fmt.Sprintf("поле '%s': условие ссылается на несуществующее поле '%s'",
								info.Path, condField))
					}
				}
				if f.Condition.EnumValue != "" {
				}
			}
		}
	}
}

func (a *Analyzer) validateBitFields() {
	for _, info := range a.symbolTable {
		bitField, ok := info.Field.(*ast.BitStructField)
		if !ok {
			continue
		}

		usedBits := make(map[int]bool)
		for _, bit := range bitField.Fields {
			if bit.IsRange {
				for i := bit.LowBit; i <= bit.HighBit; i++ {
					if usedBits[i] {
						a.errors = append(a.errors,
							fmt.Sprintf("поле '%s': бит %d уже используется в bitstruct '%s'",
								info.Path, i, bitField.Name))
					}
					usedBits[i] = true
				}
			} else {
				if usedBits[bit.Bit] {
					a.errors = append(a.errors,
						fmt.Sprintf("поле '%s': бит %d уже используется в bitstruct '%s'",
							info.Path, bit.Bit, bitField.Name))
				}
				usedBits[bit.Bit] = true
			}
		}
	}
}

func (a *Analyzer) validateCycles() {
	visited := make(map[string]bool)
	for _, info := range a.symbolTable {
		if _, ok := info.Field.(*ast.StructField); ok {
			if err := a.checkCycle(info.Path, visited); err != nil {
				a.errors = append(a.errors, err.Error())
			}
		}
	}
}

func (a *Analyzer) checkCycle(path string, visited map[string]bool) error {
	if visited[path] {
		return fmt.Errorf("обнаружена циклическая зависимость: %s", path)
	}

	info, exists := a.symbolTable[path]
	if !exists {
		return nil
	}

	structField, ok := info.Field.(*ast.StructField)
	if !ok {
		return nil
	}

	visited[path] = true
	defer delete(visited, path)

	for _, child := range structField.Struct.Fields {
		childPath := path + "." + child.GetName()
		if err := a.checkCycle(childPath, visited); err != nil {
			return err
		}
	}

	return nil
}

func (a *Analyzer) GetSymbolTable() map[string]*SymbolInfo {
	return a.symbolTable
}

func (a *Analyzer) GetFieldOffset(path string) (int, error) {
	info, exists := a.symbolTable[path]
	if !exists {
		return 0, fmt.Errorf("поле '%s' не найдено", path)
	}
	return info.Offset, nil
}
