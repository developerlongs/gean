package types

import "testing"

func TestBitvector(t *testing.T) {
	bv := NewBitvector(16)
	if bv.Len() != 16 {
		t.Error("length")
	}
	if bv.Get(0) {
		t.Error("should be false")
	}
	bv.Set(0, true)
	bv.Set(7, true)
	if !bv.Get(0) || !bv.Get(7) {
		t.Error("set bits")
	}
	if bv.Get(1) {
		t.Error("unset bit")
	}
}

func TestBitvectorFromBytes(t *testing.T) {
	bv, err := BitvectorFromBytes([]byte{0x81}, 8)
	if err != nil {
		t.Fatal(err)
	}
	if !bv.Get(0) || !bv.Get(7) {
		t.Error("bits from bytes")
	}
}

func TestBitlist(t *testing.T) {
	bl, err := BitlistFromBits([]bool{true, false, true}, 100)
	if err != nil {
		t.Fatal(err)
	}
	if bl.Len() != 3 {
		t.Error("length")
	}
	if !bl.Get(0) || bl.Get(1) || !bl.Get(2) {
		t.Error("bits")
	}
}

func TestBitlistBytes(t *testing.T) {
	bl, _ := BitlistFromBits([]bool{true, false, true}, 100)
	b := bl.Bytes()
	// bits: 101 + delimiter 1 = 1101 = 0x0D
	if len(b) != 1 || b[0] != 0x0D {
		t.Errorf("expected 0x0D, got %x", b)
	}
}

func TestEmptyBitlist(t *testing.T) {
	bl, _ := BitlistFromBits([]bool{}, 100)
	b := bl.Bytes()
	if len(b) != 1 || b[0] != 0x01 {
		t.Error("empty bitlist should be 0x01")
	}
}
