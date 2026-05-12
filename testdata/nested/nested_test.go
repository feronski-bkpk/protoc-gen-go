package nested

import (
	"testing"
)

func TestDeepNestedRoundtrip(t *testing.T) {
	p := DeepNested{
		Outer_id: 0x1234,
		Inner: Inner{
			Level1_id: 0xDEADBEEF,
			Deeper: Deeper{
				Level2_flag: 0x42,
				Deepest: Deepest{
					Value: 0x0123456789ABCDEF,
				},
			},
		},
		Trailer: 0xABCD,
	}

	data, err := p.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	var restored DeepNested
	if err := restored.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}

	if restored != p {
		t.Errorf("roundtrip mismatch:\n  original: %+v\n  restored: %+v", p, restored)
	}

	t.Logf("DeepNested roundtrip OK, size=%d", len(data))
}

func TestDeepNestedSize(t *testing.T) {
	p := DeepNested{
		Inner: Inner{
			Deeper: Deeper{
				Deepest: Deepest{},
			},
		},
	}
	expected := 2 + 4 + 1 + 8 + 2
	if p.Size() != expected {
		t.Errorf("Size() = %d, expected %d", p.Size(), expected)
	}
}
