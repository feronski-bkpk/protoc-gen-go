package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/feronski-bkpk/protoc-gen-go/internal/ast"
	"github.com/feronski-bkpk/protoc-gen-go/internal/dsl"
	"github.com/feronski-bkpk/protoc-gen-go/internal/generator"
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

	filename := os.Args[1]

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Fatalf("Файл не найден: %s", filename)
	}

	fmt.Printf("Парсинг %s...\n", filename)
	parser := dsl.NewParser()
	proto, err := parser.ParseFile(filename)
	if err != nil {
		log.Fatalf("Ошибка парсинга %s: %v", filename, err)
	}

	fmt.Printf("Успешно разобран протокол: %s (ID: 0x%04X)\n", proto.Name, proto.PacketID)
	fmt.Printf("Всего полей: %d\n", len(proto.Fields))

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
  protoc-gen-go <файл.dsl>    Сгенерировать Go код из DSL файла
  protoc-gen-go -v, --version  Показать версию
  protoc-gen-go -h, --help     Показать справку

Примеры:
  protoc-gen-go protocol.dsl
  protoc-gen-go examples/simple/simple.dsl`)
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
				fmt.Printf("%s    %s: bit(%d)\n", prefix, bit.Name, bit.Bit)
			}
			fmt.Printf("%s  }", prefix)
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
