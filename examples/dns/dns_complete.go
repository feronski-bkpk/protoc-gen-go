package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
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

func (p *DNSFlags) Size() int { return 12 }

func (p *DNSFlags) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint16(buf[0:], p.Tx_id)
	buf[2] = p.Flags
	buf[3] = p.Flags2
	binary.BigEndian.PutUint16(buf[4:], p.Qdcount)
	binary.BigEndian.PutUint16(buf[6:], p.Ancount)
	binary.BigEndian.PutUint16(buf[8:], p.Nscount)
	binary.BigEndian.PutUint16(buf[10:], p.Arcount)
	return buf, nil
}

func (p *DNSFlags) UnmarshalBinary(data []byte) error {
	if len(data) < 12 {
		return fmt.Errorf("недостаточно данных")
	}
	p.Tx_id = binary.BigEndian.Uint16(data[0:])
	p.Flags = data[2]
	p.Flags2 = data[3]
	p.Qdcount = binary.BigEndian.Uint16(data[4:])
	p.Ancount = binary.BigEndian.Uint16(data[6:])
	p.Nscount = binary.BigEndian.Uint16(data[8:])
	p.Arcount = binary.BigEndian.Uint16(data[10:])
	return nil
}

func main() {
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("     ПОЛНОЦЕННАЯ ДЕМОНСТРАЦИЯ DNS С bits[high:low]")
	fmt.Println(strings.Repeat("═", 70))

	hexData := "12340100000100000000000003777777076578616d706c6503636f6d0000010001"
	data, _ := hex.DecodeString(hexData)

	var dns DNSFlags
	dns.UnmarshalBinary(data)

	fmt.Println("\nDNS ЗАПРОС (google.com):")
	fmt.Printf("  Transaction ID: 0x%04X\n", dns.Tx_id)
	fmt.Printf("  Flags: 0x%02X\n", dns.Flags)
	fmt.Printf("  QR: %v (Query)\n", dns.GetQr())
	fmt.Printf("  Opcode: %d\n", dns.GetOpcode())
	fmt.Printf("  RD: %v\n", dns.GetRd())

	dns.SetQr(true)
	dns.SetRa(true)
	dns.SetRcode(0)
	dns.Ancount = 1

	fmt.Println("\nСИМУЛИРОВАННЫЙ ОТВЕТ:")
	fmt.Printf("  Flags: 0x%02X, Flags2: 0x%02X\n", dns.Flags, dns.Flags2)
	fmt.Printf("  QR: %v (Response)\n", dns.GetQr())
	fmt.Printf("  Opcode: %d\n", dns.GetOpcode())
	fmt.Printf("  RA: %v\n", dns.GetRa())
	fmt.Printf("  Rcode: %d\n", dns.GetRcode())
	fmt.Printf("  Answers: %d\n", dns.Ancount)

	fmt.Println("\nТЕСТ ВСЕХ OPCODE (0-15):")
	opcodeNames := []string{"QUERY", "IQUERY", "STATUS", "RESERVED", "NOTIFY", "UPDATE", "RESERVED", "RESERVED"}
	for i := 0; i < 8; i++ {
		dns.SetOpcode(uint8(i))
		name := "RESERVED"
		if i < len(opcodeNames) {
			name = opcodeNames[i]
		}
		fmt.Printf("  %d: %s\n", i, name)
	}

	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("  DSL для этого протокола:")
	fmt.Println("  flags: bitstruct {")
	fmt.Println("      qr: bit(7); opcode: bits[6:3]; aa: bit(2); tc: bit(1); rd: bit(0)")
	fmt.Println("  }")
	fmt.Println(strings.Repeat("═", 70))
}
