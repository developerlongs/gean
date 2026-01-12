package ssz

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/devlongs/gean/common/types"
)

const BytesPerChunk = 32

var ZeroHash = types.Root{}

func Hash(data []byte) types.Root {
	return types.Root(sha256.Sum256(data))
}

func HashNodes(a, b types.Root) types.Root {
	h := sha256.New()
	h.Write(a[:])
	h.Write(b[:])
	var result types.Root
	copy(result[:], h.Sum(nil))
	return result
}

func HashTreeRootUint64(value uint64) types.Root {
	var buf [32]byte
	binary.LittleEndian.PutUint64(buf[:8], value)
	return types.Root(buf)
}

func HashTreeRootBitvector(bv *types.Bitvector) types.Root {
	chunks := packBits(bv.Len(), func(i int) bool { return bv.Get(i) })
	limit := (bv.Len() + 255) / 256
	return Merkleize(chunks, limit)
}

func HashTreeRootBitlist(bl *types.Bitlist) types.Root {
	chunks := packBits(bl.Len(), func(i int) bool { return bl.Get(i) })
	limit := (bl.Limit() + 255) / 256
	root := Merkleize(chunks, limit)
	return MixInLength(root, uint64(bl.Len()))
}

func packBits(n int, get func(int) bool) []types.Root {
	if n == 0 {
		return nil
	}
	byteLen := (n + 7) / 8
	data := make([]byte, byteLen)
	for i := 0; i < n; i++ {
		if get(i) {
			data[i/8] |= 1 << (i % 8)
		}
	}
	// Pad to chunk boundary
	padded := make([]byte, ((byteLen+31)/32)*32)
	copy(padded, data)
	chunks := make([]types.Root, len(padded)/32)
	for i := range chunks {
		copy(chunks[i][:], padded[i*32:(i+1)*32])
	}
	return chunks
}

func Merkleize(chunks []types.Root, limit int) types.Root {
	n := len(chunks)
	if n == 0 {
		if limit > 0 {
			return zeroTreeRoot(nextPowerOfTwo(limit))
		}
		return ZeroHash
	}

	width := nextPowerOfTwo(n)
	if limit > 0 && limit >= n {
		width = nextPowerOfTwo(limit)
	}

	if width == 1 {
		return chunks[0]
	}

	level := make([]types.Root, width)
	copy(level, chunks)

	for len(level) > 1 {
		next := make([]types.Root, len(level)/2)
		for i := range next {
			next[i] = HashNodes(level[i*2], level[i*2+1])
		}
		level = next
	}
	return level[0]
}

func MixInLength(root types.Root, length uint64) types.Root {
	var lenChunk types.Root
	binary.LittleEndian.PutUint64(lenChunk[:8], length)
	return HashNodes(root, lenChunk)
}

func nextPowerOfTwo(x int) int {
	if x <= 1 {
		return 1
	}
	n := x - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	return n + 1
}

func zeroTreeRoot(width int) types.Root {
	if width <= 1 {
		return ZeroHash
	}
	h := ZeroHash
	for width > 1 {
		h = HashNodes(h, h)
		width /= 2
	}
	return h
}
