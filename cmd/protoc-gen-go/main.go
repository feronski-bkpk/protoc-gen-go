package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/feronski-bkpk/protoc-gen-go/internal/analyzer"
	"github.com/feronski-bkpk/protoc-gen-go/internal/ast"
	"github.com/feronski-bkpk/protoc-gen-go/internal/binary"
	"github.com/feronski-bkpk/protoc-gen-go/internal/generator"
	"github.com/feronski-bkpk/protoc-gen-go/internal/parser"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	if os.Args[1] == "-v" || os.Args[1] == "--version" {
		fmt.Printf("protoc-gen-go version %s (built %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		printUsage()
		os.Exit(0)
	}

	if os.Args[1] == "--save-bin" {
		if len(os.Args) < 3 {
			log.Fatal("Использование: protoc-gen-go --save-bin <файл.dsl>")
		}
		filename := os.Args[2]
		proto, err := parser.ParseFile(filename)
		if err != nil {
			log.Fatalf("Ошибка парсинга: %v", err)
		}
		data, err := binary.WriteProtocol(proto)
		if err != nil {
			log.Fatalf("Ошибка сериализации: %v", err)
		}
		outFile := strings.TrimSuffix(filename, ".dsl") + ".bin"
		if err := os.WriteFile(outFile, data, 0644); err != nil {
			log.Fatalf("Ошибка записи: %v", err)
		}
		fmt.Printf("Схема сохранена: %s (%d байт)\n", outFile, len(data))
		os.Exit(0)
	}

	if os.Args[1] == "--load-bin" {
		if len(os.Args) < 3 {
			log.Fatal("Использование: protoc-gen-go --load-bin <файл.bin>")
		}
		data, err := os.ReadFile(os.Args[2])
		if err != nil {
			log.Fatalf("Ошибка чтения: %v", err)
		}
		proto, err := binary.ReadProtocol(data)
		if err != nil {
			log.Fatalf("Ошибка десериализации: %v", err)
		}
		fmt.Printf("Загружен протокол: %s (ID: 0x%04X)\n", proto.Name, proto.PacketID)
		fmt.Printf("Всего полей: %d\n\n", len(proto.Fields))

		gen := generator.NewGenerator(proto)
		code, err := gen.Generate()
		if err != nil {
			log.Fatalf("Ошибка генерации: %v", err)
		}
		fmt.Println(code)
		os.Exit(0)
	}

	filename := os.Args[1]

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Fatalf("Файл не найден: %s", filename)
	}

	fmt.Printf("Парсинг %s...\n", filename)
	proto, err := parser.ParseFile(filename)
	if err != nil {
		log.Fatalf("Ошибка парсинга %s: %v", filename, err)
	}

	fmt.Printf("Успешно разобран протокол: %s (ID: 0x%04X)\n", proto.Name, proto.PacketID)
	fmt.Printf("Всего полей: %d\n", len(proto.Fields))

	fmt.Printf("\nАнализ протокола...\n")
	a := analyzer.NewAnalyzer(proto)
	if err := a.Analyze(); err != nil {
		log.Fatalf("Ошибка анализа: %v", err)
	}
	fmt.Printf("Анализ завершён успешно\n")

	fmt.Println("\nТаблица символов:")
	symTable := a.GetSymbolTable()
	var paths []string
	for path := range symTable {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		info := symTable[path]
		fmt.Printf("  %-35s offset=%4d size=%4d\n", path, info.Offset, info.Size)
	}

	fmt.Printf("\nГенерация Go кода...\n")
	gen := generator.NewGenerator(proto)
	code, err := gen.Generate()
	if err != nil {
		log.Fatalf("Ошибка генерации кода: %v", err)
	}

	outputFile := strings.TrimSuffix(filename, ".dsl") + ".gen.go"
	if err := os.WriteFile(outputFile, []byte(code), 0644); err != nil {
		log.Fatalf("Ошибка записи файла: %v", err)
	}

	fmt.Printf("Код сгенерирован: %s\n", outputFile)

	fmt.Println("\nСтруктура протокола:")
	printFields(proto.Fields, 0)
}

func printUsage() {
	fmt.Println(`protoc-gen-go - Генератор протоколов из DSL в Go

Использование:
  protoc-gen-go <файл.dsl>       Сгенерировать Go код из DSL файла
  protoc-gen-go --save-bin <файл.dsl>  Сохранить схему в бинарный формат
  protoc-gen-go --load-bin <файл.bin>  Загрузить схему из бинарного формата
  protoc-gen-go -v, --version     Показать версию
  protoc-gen-go -h, --help        Показать справку

Примеры:
  protoc-gen-go protocol.dsl
  protoc-gen-go --save-bin protocol.dsl
  protoc-gen-go --load-bin protocol.bin`)
}

func printFields(fields []ast.Field, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	for _, field := range fields {
		switch f := field.(type) {
		case *ast.ScalarField:
			fmt.Printf("%s- %s: %s", prefix, f.Name, f.Type)
			if f.Condition != nil {
				fmt.Printf(" [если %s %s %d]", f.Condition.Field, f.Condition.Operator, f.Condition.Value)
			}
			fmt.Println()

		case *ast.StructField:
			fmt.Printf("%s- %s: struct {\n", prefix, f.Name)
			printFields(f.Struct.Fields, indent+1)
			fmt.Printf("%s  }", prefix)
			if f.Condition != nil {
				fmt.Printf(" [если %s %s %d]", f.Condition.Field, f.Condition.Operator, f.Condition.Value)
			}
			fmt.Println()

		case *ast.BitStructField:
			fmt.Printf("%s- %s: bitstruct {\n", prefix, f.Name)
			for _, bit := range f.Fields {
				if bit.IsRange {
					fmt.Printf("%s    %s: bits[%d:%d]\n", prefix, bit.Name, bit.HighBit, bit.LowBit)
				} else {
					fmt.Printf("%s    %s: bit(%d)\n", prefix, bit.Name, bit.Bit)
				}
			}
			fmt.Printf("%s  }", prefix)
			if f.Condition != nil {
				fmt.Printf(" [если %s %s %d]", f.Condition.Field, f.Condition.Operator, f.Condition.Value)
			}
			fmt.Println()

		case *ast.ArrayField:
			if f.FixedLength > 0 {
				fmt.Printf("%s- %s: [%d]", prefix, f.Name, f.FixedLength)
			} else {
				fmt.Printf("%s- %s: []", prefix, f.Name)
			}
			switch elem := f.ElementType.(type) {
			case *ast.ScalarField:
				fmt.Printf("%s", elem.Type)
			case *ast.StructField:
				fmt.Printf("struct")
			}
			if f.LengthFrom != "" {
				fmt.Printf(" (длина из: %s)", f.LengthFrom)
			}
			if f.Condition != nil {
				fmt.Printf(" [если %s %s %d]", f.Condition.Field, f.Condition.Operator, f.Condition.Value)
			}
			fmt.Println()

		case *ast.BytesField:
			fmt.Printf("%s- %s: bytes", prefix, f.Name)
			if f.LengthFrom != "" {
				fmt.Printf(" (длина из: %s)", f.LengthFrom)
			}
			if f.Condition != nil {
				fmt.Printf(" [если %s %s %d]", f.Condition.Field, f.Condition.Operator, f.Condition.Value)
			}
			fmt.Println()
		}
	}
}
