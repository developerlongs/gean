package types

import "fmt"

// Bitvector is a fixed-length bit array.
type Bitvector struct {
	data   []byte
	length int
}

func NewBitvector(length int) *Bitvector {
	return &Bitvector{
		data:   make([]byte, (length+7)/8),
		length: length,
	}
}

func BitvectorFromBytes(data []byte, length int) (*Bitvector, error) {
	expectedLen := (length + 7) / 8
	if len(data) != expectedLen {
		return nil, fmt.Errorf("expected %d bytes for %d bits, got %d", expectedLen, length, len(data))
	}
	copied := make([]byte, len(data))
	copy(copied, data)
	return &Bitvector{data: copied, length: length}, nil
}

func (b *Bitvector) Len() int     { return b.length }
func (b *Bitvector) ByteLen() int { return len(b.data) }

func (b *Bitvector) Get(index int) bool {
	if index < 0 || index >= b.length {
		return false
	}
	return (b.data[index/8] & (1 << (index % 8))) != 0
}

func (b *Bitvector) Set(index int, value bool) {
	if index < 0 || index >= b.length {
		return
	}
	if value {
		b.data[index/8] |= 1 << (index % 8)
	} else {
		b.data[index/8] &^= 1 << (index % 8)
	}
}

func (b *Bitvector) Bytes() []byte {
	result := make([]byte, len(b.data))
	copy(result, b.data)
	return result
}

// Bitlist is a variable-length bit array with a maximum capacity.
type Bitlist struct {
	data  []byte
	len   int
	limit int
}

func NewBitlist(limit int) *Bitlist {
	return &Bitlist{limit: limit}
}

func BitlistFromBits(bits []bool, limit int) (*Bitlist, error) {
	if len(bits) > limit {
		return nil, fmt.Errorf("bitlist exceeds limit of %d, got %d", limit, len(bits))
	}
	byteLen := (len(bits) + 7) / 8
	data := make([]byte, byteLen)
	for i, bit := range bits {
		if bit {
			data[i/8] |= 1 << (i % 8)
		}
	}
	return &Bitlist{data: data, len: len(bits), limit: limit}, nil
}

func (b *Bitlist) Len() int   { return b.len }
func (b *Bitlist) Limit() int { return b.limit }

func (b *Bitlist) Get(index int) bool {
	if index < 0 || index >= b.len {
		return false
	}
	return (b.data[index/8] & (1 << (index % 8))) != 0
}

func (b *Bitlist) Set(index int, value bool) {
	if index < 0 || index >= b.len {
		return
	}
	if value {
		b.data[index/8] |= 1 << (index % 8)
	} else {
		b.data[index/8] &^= 1 << (index % 8)
	}
}

// Bytes returns SSZ-encoded bytes with delimiter bit.
func (b *Bitlist) Bytes() []byte {
	if b.len == 0 {
		return []byte{0x01}
	}
	byteLen := (b.len + 7) / 8
	result := make([]byte, byteLen)
	copy(result, b.data)
	if b.len%8 == 0 {
		return append(result, 0x01)
	}
	result[b.len/8] |= 1 << (b.len % 8)
	return result
}
