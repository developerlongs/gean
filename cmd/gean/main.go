package main

import (
	"fmt"

	"github.com/devlongs/gean/common/ssz"
	"github.com/devlongs/gean/common/types"
)

func main() {
	fmt.Println("Gean - Go Lean Ethereum Client")
	fmt.Println()

	// Primitives
	slot := types.Slot(100)
	idx := types.ValidatorIndex(42)
	var root types.Root
	root[0] = 0xde

	fmt.Printf("Slot: %d\n", slot)
	fmt.Printf("ValidatorIndex: %d\n", idx)
	fmt.Printf("Root: %x...\n", root[:4])
	fmt.Println()

	// Time
	genesis := uint64(1700000000)
	fmt.Printf("Slot %d -> Time %d\n", slot, types.SlotToTime(slot, genesis))
	fmt.Println()

	// Byte arrays
	var b4 types.Bytes4
	var b52 types.Bytes52
	b4[0] = 0x01
	b52[0] = 0x02
	fmt.Printf("Bytes4: %x\n", b4[:])
	fmt.Printf("Bytes52: %x... (%d bytes)\n", b52[:4], len(b52))
	fmt.Println()

	// Bitvector
	bv := types.NewBitvector(16)
	bv.Set(0, true)
	bv.Set(7, true)
	fmt.Printf("Bitvector(16): bits 0,7 set -> %x\n", bv.Bytes())
	fmt.Println()

	// Bitlist
	bl, _ := types.BitlistFromBits([]bool{true, false, true}, 100)
	fmt.Printf("Bitlist [t,f,t]: SSZ bytes -> %x\n", bl.Bytes())
	fmt.Println()

	// SSZ
	h1 := ssz.HashTreeRootUint64(uint64(slot))
	h2 := ssz.HashTreeRootBitvector(bv)
	h3 := ssz.HashTreeRootBitlist(bl)
	fmt.Printf("HashTreeRoot(Slot): %x...\n", h1[:8])
	fmt.Printf("HashTreeRoot(Bitvector): %x...\n", h2[:8])
	fmt.Printf("HashTreeRoot(Bitlist): %x...\n", h3[:8])
	fmt.Println()

	fmt.Println("Milestone 1 complete!")
}
