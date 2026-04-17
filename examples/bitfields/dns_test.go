package main

import (
	"fmt"

	"github.com/feronski-bkpk/protoc-gen-go/examples/bitfields/protocol"
)

func main() {
	var flags protocol.DNSFlags

	flags.SetQr(false)
	flags.SetOpcode(0)
	flags.SetAa(false)
	flags.SetTc(false)
	flags.SetRd(true)

	flags.SetRa(false)
	flags.SetZ(0)
	flags.SetRcode(0)

	fmt.Printf("DNS Flags 1: 0x%02X\n", flags.Flags)
	fmt.Printf("  QR: %v\n", flags.GetQr())
	fmt.Printf("  Opcode: %d\n", flags.GetOpcode())
	fmt.Printf("  RD: %v\n", flags.GetRd())

	fmt.Printf("DNS Flags 2: 0x%02X\n", flags.Flags2)

	flags.SetQr(true)
	flags.SetRa(true)
	flags.SetRcode(0)

	fmt.Printf("\nDNS Response Flags:\n")
	fmt.Printf("  Flags1: 0x%02X\n", flags.Flags)
	fmt.Printf("  Flags2: 0x%02X\n", flags.Flags2)
	fmt.Printf("  QR (Response): %v\n", flags.GetQr())
	fmt.Printf("  RA: %v\n", flags.GetRa())
	fmt.Printf("  Rcode: %d\n", flags.GetRcode())
}
