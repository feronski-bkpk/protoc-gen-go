package main

import (
	"fmt"
)

type DNSFlags struct {
	Tx_id   uint16
	Flags   uint8
	Flags2  uint8
	Qdcount uint16
	Ancount uint16
	Nscount uint16
	Arcount uint16
}

func (p *DNSFlags) GetQr() bool { return (p.Flags & (1 << 7)) != 0 }
func (p *DNSFlags) SetQr(v bool) {
	if v {
		p.Flags |= 1 << 7
	} else {
		p.Flags &^= 1 << 7
	}
}
func (p *DNSFlags) GetOpcode() uint8  { return (p.Flags >> 3) & 0x0F }
func (p *DNSFlags) SetOpcode(v uint8) { p.Flags = (p.Flags & 0x87) | ((v & 0x0F) << 3) }
func (p *DNSFlags) GetAa() bool       { return (p.Flags & (1 << 2)) != 0 }
func (p *DNSFlags) SetAa(v bool) {
	if v {
		p.Flags |= 1 << 2
	} else {
		p.Flags &^= 1 << 2
	}
}
func (p *DNSFlags) GetTc() bool { return (p.Flags & (1 << 1)) != 0 }
func (p *DNSFlags) SetTc(v bool) {
	if v {
		p.Flags |= 1 << 1
	} else {
		p.Flags &^= 1 << 1
	}
}
func (p *DNSFlags) GetRd() bool { return (p.Flags & (1 << 0)) != 0 }
func (p *DNSFlags) SetRd(v bool) {
	if v {
		p.Flags |= 1 << 0
	} else {
		p.Flags &^= 1 << 0
	}
}

func (p *DNSFlags) GetRa() bool { return (p.Flags2 & (1 << 7)) != 0 }
func (p *DNSFlags) SetRa(v bool) {
	if v {
		p.Flags2 |= 1 << 7
	} else {
		p.Flags2 &^= 1 << 7
	}
}
func (p *DNSFlags) GetZ() uint8      { return (p.Flags2 >> 4) & 0x07 }
func (p *DNSFlags) SetZ(v uint8)     { p.Flags2 = (p.Flags2 & 0x8F) | ((v & 0x07) << 4) }
func (p *DNSFlags) GetRcode() uint8  { return p.Flags2 & 0x0F }
func (p *DNSFlags) SetRcode(v uint8) { p.Flags2 = (p.Flags2 & 0xF0) | (v & 0x0F) }

func main() {
	fmt.Println("=== ДЕМОНСТРАЦИЯ DNS ФЛАГОВ ===\n")

	var dns DNSFlags

	dns.SetQr(false)
	dns.SetOpcode(0)
	dns.SetAa(false)
	dns.SetTc(false)
	dns.SetRd(true)

	fmt.Println("DNS ЗАПРОС:")
	fmt.Printf("  Flags:  0x%02X\n", dns.Flags)
	fmt.Printf("  QR:     %v (Query)\n", dns.GetQr())
	fmt.Printf("  Opcode: %d\n", dns.GetOpcode())
	fmt.Printf("  RD:     %v\n", dns.GetRd())

	dns.SetQr(true)
	dns.SetRa(true)
	dns.SetRcode(0)

	fmt.Println("\nDNS ОТВЕТ:")
	fmt.Printf("  Flags:  0x%02X\n", dns.Flags)
	fmt.Printf("  Flags2: 0x%02X\n", dns.Flags2)
	fmt.Printf("  QR:     %v (Response)\n", dns.GetQr())
	fmt.Printf("  RA:     %v\n", dns.GetRa())
	fmt.Printf("  Rcode:  %d\n", dns.GetRcode())

	fmt.Println("\nТЕСТ OPCODE (4 бита, значения 0-15):")
	for i := uint8(0); i <= 15; i++ {
		dns.SetOpcode(i)
		if dns.GetOpcode() != i {
			fmt.Printf("  ОШИБКА: установлено %d, получено %d\n", i, dns.GetOpcode())
		}
	}
	fmt.Println("Все значения Opcode корректны")

	fmt.Println("\nТЕСТ RCODE (4 бита, значения 0-15):")
	for i := uint8(0); i <= 15; i++ {
		dns.SetRcode(i)
		if dns.GetRcode() != i {
			fmt.Printf("  ОШИБКА: установлено %d, получено %d\n", i, dns.GetRcode())
		}
	}
	fmt.Println("Все значения Rcode корректны")

	fmt.Println("\n=== ДЕМОНСТРАЦИЯ ЗАВЕРШЕНА ===")
}
