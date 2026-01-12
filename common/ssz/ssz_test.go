package ssz

import (
	"testing"

	"github.com/devlongs/gean/common/types"
)

func TestHash(t *testing.T) {
	h := Hash([]byte("hello"))
	if h.IsZero() {
		t.Error("should not be zero")
	}
	if Hash([]byte("hello")) != h {
		t.Error("should be deterministic")
	}
}

func TestHashNodes(t *testing.T) {
	a, b := types.Root{1}, types.Root{2}
	h := HashNodes(a, b)
	if h.IsZero() {
		t.Error("should not be zero")
	}
	if HashNodes(b, a) == h {
		t.Error("order should matter")
	}
}

func TestHashTreeRootUint64(t *testing.T) {
	r := HashTreeRootUint64(100)
	if r[0] != 100 {
		t.Error("first byte")
	}
}

func TestMerkleize(t *testing.T) {
	chunk := types.Root{1}
	if Merkleize([]types.Root{chunk}, 0) != chunk {
		t.Error("single chunk")
	}

	a, b := types.Root{1}, types.Root{2}
	if Merkleize([]types.Root{a, b}, 0) != HashNodes(a, b) {
		t.Error("two chunks")
	}

	if Merkleize(nil, 0) != ZeroHash {
		t.Error("empty")
	}
}

func TestMixInLength(t *testing.T) {
	root := types.Root{1}
	mixed := MixInLength(root, 42)
	if mixed == root {
		t.Error("should change")
	}
}

func TestHashTreeRootBitvector(t *testing.T) {
	bv := types.NewBitvector(8)
	bv.Set(0, true)
	r := HashTreeRootBitvector(bv)
	if r.IsZero() {
		t.Error("should not be zero")
	}
}

func TestHashTreeRootBitlist(t *testing.T) {
	bl, _ := types.BitlistFromBits([]bool{true, false, true}, 100)
	r := HashTreeRootBitlist(bl)
	if r.IsZero() {
		t.Error("should not be zero")
	}
}

func TestNextPowerOfTwo(t *testing.T) {
	tests := [][2]int{{0, 1}, {1, 1}, {2, 2}, {3, 4}, {5, 8}, {8, 8}, {9, 16}}
	for _, tt := range tests {
		if nextPowerOfTwo(tt[0]) != tt[1] {
			t.Errorf("nextPowerOfTwo(%d) = %d, want %d", tt[0], nextPowerOfTwo(tt[0]), tt[1])
		}
	}
}
