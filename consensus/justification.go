package consensus

import "math"

// IsJustifiableAfter checks if a slot can be justified given the finalized slot.
// Implements 3SF-mini rules:
//   - delta <= 5: always justifiable
//   - delta is a perfect square (9, 16, 25...)
//   - delta is a pronic number (6, 12, 20, 30...) i.e. x*(x+1)
func (s Slot) IsJustifiableAfter(finalizedSlot Slot) bool {
	if s < finalizedSlot {
		return false
	}

	delta := uint64(s - finalizedSlot)

	// Rule 1: first few slots after finalization
	if delta <= 5 {
		return true
	}

	// Rule 2: perfect square
	sqrt := math.Sqrt(float64(delta))
	if sqrt == math.Floor(sqrt) {
		return true
	}

	// Rule 3: pronic number (x² + x)
	// A number n is pronic if sqrt(n + 0.25) has fractional part 0.5
	sqrtPronic := math.Sqrt(float64(delta) + 0.25)
	if math.Mod(sqrtPronic, 1) == 0.5 {
		return true
	}

	return false
}
