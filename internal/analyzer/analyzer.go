package analyzer

import (
	"fmt"
	"strings"

	"github.com/feronski-bkpk/protoc-gen-go/internal/ast"
)

// Analyzer выполняет семантический анализ протокола
type Analyzer struct {
	proto       *ast.Protocol
	symbolTable map[string]*SymbolInfo
	errors      []string
}

// SymbolInfo хранит информацию о символе (поле)
type SymbolInfo struct {
	Field  ast.Field // само поле
	Path   string    // полный путь (например "location.latitude")
	Offset int       // смещение в байтах
	Size   int       // размер в байтах
}

// NewAnalyzer создаёт новый анализатор
func NewAnalyzer(proto *ast.Protocol) *Analyzer {
	return &Analyzer{
		proto:       proto,
		symbolTable: make(map[string]*SymbolInfo),
	}
}

// Analyze выполняет полный анализ протокола
func (a *Analyzer) Analyze() error {
	// 1. Строим таблицу символов
	a.buildSymbolTable(a.proto.Fields, "", 0)

	// 2. Вычисляем смещения
	a.computeOffsets(a.proto.Fields, 0)

	// 3. Проверяем ссылки (length_from)
	a.validateReferences()

	// 4. Проверяем битовые поля
	a.validateBitFields()

	// 5. Проверяем циклические зависимости
	a.validateCycles()

	if len(a.errors) > 0 {
		return fmt.Errorf("ошибки анализа:\n%s", strings.Join(a.errors, "\n"))
	}

	return nil
}

// buildSymbolTable рекурсивно строит таблицу символов
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

// computeOffsets вычисляет смещения полей
func (a *Analyzer) computeOffsets(fields []ast.Field, baseOffset int) {
	offset := baseOffset

	for _, field := range fields {
		path := field.GetName()
		for key, info := range a.symbolTable {
			if strings.HasSuffix(key, path) {
				info.Offset = offset
				break
			}
		}

		switch f := field.(type) {
		case *ast.ScalarField:
			f.Offset = offset
			offset += f.GetSize()

		case *ast.StructField:
			f.Offset = offset
			a.computeOffsets(f.Struct.Fields, offset)
			offset += a.symbolTable[path].Size

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

// validateReferences проверяет корректность ссылок length_from
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
				if _, exists := a.symbolTable[f.Condition.Field]; !exists {
					a.errors = append(a.errors,
						fmt.Sprintf("поле '%s': условие ссылается на несуществующее поле '%s'",
							info.Path, f.Condition.Field))
				}
			}
		}
	}
}

// validateBitFields проверяет перекрытие битов в bitstruct
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

// validateCycles проверяет циклические зависимости в структурах
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

// GetSymbolTable возвращает таблицу символов
func (a *Analyzer) GetSymbolTable() map[string]*SymbolInfo {
	return a.symbolTable
}

// GetFieldOffset возвращает смещение поля по пути
func (a *Analyzer) GetFieldOffset(path string) (int, error) {
	info, exists := a.symbolTable[path]
	if !exists {
		return 0, fmt.Errorf("поле '%s' не найдено", path)
	}
	return info.Offset, nil
}
