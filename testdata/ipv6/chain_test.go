package ipv6

import (
	"testing"

	"github.com/feronski-bkpk/protoc-gen-go/testdata/ipv6/base"
	"github.com/feronski-bkpk/protoc-gen-go/testdata/ipv6/fragment"
	"github.com/feronski-bkpk/protoc-gen-go/testdata/ipv6/hopbyhop"
)

func TestIPv6Chain(t *testing.T) {
	baseHdr := base.Ipv6BaseHeader{
		Next_header:    0,
		Hop_limit:      64,
		Payload_length: 16,
	}
	baseHdr.SetVersion(6)

	hopHdr := hopbyhop.Ipv6HopByHop{
		Next_header: 44,
		Options_len: 4,
		Options:     []byte{0, 0, 0, 0},
	}

	fragHdr := fragment.Ipv6Fragment{
		Next_header:    6, // TCP
		Identification: 1,
	}

	baseBytes, err := baseHdr.MarshalBinary()
	if err != nil {
		t.Fatal("base marshal:", err)
	}
	hopBytes, err := hopHdr.MarshalBinary()
	if err != nil {
		t.Fatal("hop marshal:", err)
	}
	fragBytes, err := fragHdr.MarshalBinary()
	if err != nil {
		t.Fatal("frag marshal:", err)
	}

	t.Logf("Base: %d байт", len(baseBytes))
	t.Logf("HopByHop: %d байт", len(hopBytes))
	t.Logf("Fragment: %d байт", len(fragBytes))

	if len(baseBytes) != 40 {
		t.Errorf("base size = %d, expected 40", len(baseBytes))
	}
	if len(hopBytes) != 8 {
		t.Errorf("hop size = %d, expected 8", len(hopBytes))
	}
	if len(fragBytes) != 8 {
		t.Errorf("frag size = %d, expected 8", len(fragBytes))
	}

	chain := append(baseBytes, hopBytes...)
	chain = append(chain, fragBytes...)

	if len(chain) != 56 {
		t.Errorf("chain size = %d, expected 56", len(chain))
	}

	var restoredBase base.Ipv6BaseHeader
	if err := restoredBase.UnmarshalBinary(chain[0:40]); err != nil {
		t.Fatal("base unmarshal:", err)
	}
	if restoredBase.Next_header != 0 {
		t.Errorf("base next_header = %d, expected 0", restoredBase.Next_header)
	}

	var restoredHop hopbyhop.Ipv6HopByHop
	if err := restoredHop.UnmarshalBinary(chain[40:48]); err != nil {
		t.Fatal("hop unmarshal:", err)
	}
	if restoredHop.Next_header != 44 {
		t.Errorf("hop next_header = %d, expected 44", restoredHop.Next_header)
	}

	var restoredFrag fragment.Ipv6Fragment
	if err := restoredFrag.UnmarshalBinary(chain[48:56]); err != nil {
		t.Fatal("frag unmarshal:", err)
	}
	if restoredFrag.Next_header != 6 {
		t.Errorf("frag next_header = %d, expected 6", restoredFrag.Next_header)
	}

	t.Log("IPv6 chain roundtrip: OK")
}
