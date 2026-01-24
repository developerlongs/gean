package consensus

import (
	"fmt"
	"math"
)

var ZeroHash = Root{}

// ProcessSlot performs per-slot maintenance.
// If the latest block header has an empty state_root, fill it with the current state root.
func (s *State) ProcessSlot() (*State, error) {
	if s.LatestBlockHeader.StateRoot.IsZero() {
		stateRoot, err := s.HashTreeRoot()
		if err != nil {
			return nil, fmt.Errorf("hash state: %w", err)
		}
		newState := s.Copy()
		newState.LatestBlockHeader.StateRoot = stateRoot
		return newState, nil
	}
	return s, nil
}

// ProcessSlots advances the state through empty slots up to targetSlot.
func (s *State) ProcessSlots(targetSlot Slot) (*State, error) {
	if s.Slot >= targetSlot {
		return nil, fmt.Errorf("target slot %d must be greater than current slot %d", targetSlot, s.Slot)
	}

	state := s
	var err error
	for state.Slot < targetSlot {
		state, err = state.ProcessSlot()
		if err != nil {
			return nil, err
		}
		newState := state.Copy()
		newState.Slot++
		state = newState
	}
	return state, nil
}

// ProcessBlockHeader validates and applies a block header.
func (s *State) ProcessBlockHeader(block *Block) (*State, error) {
	// Validate slot matches
	if block.Slot != s.Slot {
		return nil, fmt.Errorf("block slot %d != state slot %d", block.Slot, s.Slot)
	}

	// Block must be newer than latest header
	if block.Slot <= s.LatestBlockHeader.Slot {
		return nil, fmt.Errorf("block slot %d <= latest header slot %d", block.Slot, s.LatestBlockHeader.Slot)
	}

	// Validate proposer (round-robin)
	expectedProposer := uint64(block.Slot) % s.Config.NumValidators
	if block.ProposerIndex != expectedProposer {
		return nil, fmt.Errorf("invalid proposer %d for slot %d, expected %d", block.ProposerIndex, block.Slot, expectedProposer)
	}

	// Validate parent root
	expectedParent, err := s.LatestBlockHeader.HashTreeRoot()
	if err != nil {
		return nil, fmt.Errorf("hash latest header: %w", err)
	}
	if block.ParentRoot != expectedParent {
		return nil, fmt.Errorf("parent root mismatch")
	}

	newState := s.Copy()

	// First block after genesis: mark genesis as justified and finalized
	if s.LatestBlockHeader.Slot == 0 {
		newState.LatestJustified.Root = block.ParentRoot
		newState.LatestFinalized.Root = block.ParentRoot
	}

	// Append parent root to history
	newState.HistoricalBlockHashes = append(newState.HistoricalBlockHashes, block.ParentRoot)

	// Track justified slot (genesis slot 0 is always justified)
	parentSlot := int(s.LatestBlockHeader.Slot)
	newState.JustifiedSlots = appendBitAt(newState.JustifiedSlots, parentSlot, s.LatestBlockHeader.Slot == 0)

	// Fill empty slots with zero hashes
	emptySlots := int(block.Slot - s.LatestBlockHeader.Slot - 1)
	for i := 0; i < emptySlots; i++ {
		newState.HistoricalBlockHashes = append(newState.HistoricalBlockHashes, ZeroHash)
		emptySlot := parentSlot + 1 + i
		newState.JustifiedSlots = appendBitAt(newState.JustifiedSlots, emptySlot, false)
	}

	// Create new block header (state_root left empty, filled by next ProcessSlot)
	bodyRoot, err := block.Body.HashTreeRoot()
	if err != nil {
		return nil, fmt.Errorf("hash body: %w", err)
	}
	newState.LatestBlockHeader = BlockHeader{
		Slot:          block.Slot,
		ProposerIndex: block.ProposerIndex,
		ParentRoot:    block.ParentRoot,
		StateRoot:     Root{},
		BodyRoot:      bodyRoot,
	}

	return newState, nil
}

// ProcessAttestations processes attestation votes per Devnet 0 spec.
// Requires 2/3 supermajority to justify a target.
func (s *State) ProcessAttestations(attestations []SignedVote) (*State, error) {
	newState := s.Copy()

	// Get current justifications map
	justifications := newState.GetJustifications()

	for _, signed := range attestations {
		vote := signed.Data

		// Skip if source slot >= target slot
		if vote.Source.Slot >= vote.Target.Slot {
			continue
		}

		sourceSlot := int(vote.Source.Slot)
		targetSlot := int(vote.Target.Slot)

		// Skip if source is not justified
		if !getBit(newState.JustifiedSlots, sourceSlot) {
			continue
		}

		// Skip if target is already justified
		if getBit(newState.JustifiedSlots, targetSlot) {
			continue
		}

		// Skip if source root doesn't match historical
		if sourceSlot >= len(newState.HistoricalBlockHashes) {
			continue
		}
		if vote.Source.Root != newState.HistoricalBlockHashes[sourceSlot] {
			continue
		}

		// Skip if target root doesn't match historical
		if targetSlot >= len(newState.HistoricalBlockHashes) {
			continue
		}
		if vote.Target.Root != newState.HistoricalBlockHashes[targetSlot] {
			continue
		}

		// Skip if target slot is not justifiable after finalized
		if !IsJustifiableSlot(int(newState.LatestFinalized.Slot), targetSlot) {
			continue
		}

		// Track vote for this target root
		targetKey := vote.Target.Root
		if _, exists := justifications[targetKey]; !exists {
			justifications[targetKey] = make([]bool, newState.Config.NumValidators)
		}

		validatorID := int(vote.ValidatorID)
		if validatorID >= len(justifications[targetKey]) {
			continue
		}

		if !justifications[targetKey][validatorID] {
			justifications[targetKey][validatorID] = true
		}

		// Count votes for this target
		count := 0
		for _, voted := range justifications[targetKey] {
			if voted {
				count++
			}
		}

		// Check 2/3 supermajority: 3 * count >= 2 * num_validators
		if 3*count >= 2*int(newState.Config.NumValidators) {
			// Justify the target
			newState.LatestJustified = vote.Target
			newState.JustifiedSlots = setBit(newState.JustifiedSlots, targetSlot, true)

			// Remove from pending justifications
			delete(justifications, targetKey)

			// Check finalization: if no justifiable slots between source and target
			canFinalize := true
			for slot := int(vote.Source.Slot) + 1; slot < targetSlot; slot++ {
				if IsJustifiableSlot(int(newState.LatestFinalized.Slot), slot) {
					canFinalize = false
					break
				}
			}
			if canFinalize {
				newState.LatestFinalized = vote.Source
			}
		}
	}

	// Save justifications back to state
	newState.SetJustifications(justifications)

	return newState, nil
}

// ProcessBlock applies full block processing.
func (s *State) ProcessBlock(block *Block) (*State, error) {
	state, err := s.ProcessBlockHeader(block)
	if err != nil {
		return nil, err
	}
	return state.ProcessAttestations(block.Body.Attestations)
}

// StateTransition applies the complete state transition for a signed block.
// For Devnet 0, valid_signatures is always true (no signature verification).
func (s *State) StateTransition(signedBlock *SignedBlock, validateResult bool) (*State, error) {
	block := &signedBlock.Message

	// Process slots up to block slot
	state, err := s.ProcessSlots(block.Slot)
	if err != nil {
		return nil, err
	}

	// Process the block
	newState, err := state.ProcessBlock(block)
	if err != nil {
		return nil, err
	}

	// Validate state root
	if validateResult {
		computedRoot, err := newState.HashTreeRoot()
		if err != nil {
			return nil, fmt.Errorf("hash new state: %w", err)
		}
		if block.StateRoot != computedRoot {
			return nil, fmt.Errorf("state root mismatch: expected %x, got %x", block.StateRoot, computedRoot)
		}
	}

	return newState, nil
}

// Copy creates a deep copy of the state.
func (s *State) Copy() *State {
	cp := *s
	cp.HistoricalBlockHashes = append([]Root{}, s.HistoricalBlockHashes...)
	cp.JustifiedSlots = append([]byte{}, s.JustifiedSlots...)
	cp.JustificationRoots = append([]Root{}, s.JustificationRoots...)
	cp.JustificationValidators = append([]byte{}, s.JustificationValidators...)
	return &cp
}

// GetJustifications returns a map of root -> validator votes from state.
func (s *State) GetJustifications() map[Root][]bool {
	justifications := make(map[Root][]bool)
	numValidators := int(s.Config.NumValidators)

	for i, root := range s.JustificationRoots {
		startIdx := i * numValidators

		votes := make([]bool, numValidators)
		for j := 0; j < numValidators && startIdx+j < len(s.JustificationValidators)*8; j++ {
			votes[j] = getBit(s.JustificationValidators, startIdx+j)
		}
		justifications[root] = votes
	}

	return justifications
}

// SetJustifications saves the justifications map back to state.
func (s *State) SetJustifications(justifications map[Root][]bool) {
	s.JustificationRoots = make([]Root, 0, len(justifications))
	numValidators := int(s.Config.NumValidators)

	// Collect and sort roots for deterministic ordering
	roots := make([]Root, 0, len(justifications))
	for root := range justifications {
		roots = append(roots, root)
	}
	// Sort by bytes (lexicographic)
	for i := 0; i < len(roots); i++ {
		for j := i + 1; j < len(roots); j++ {
			if compareRoots(roots[i], roots[j]) > 0 {
				roots[i], roots[j] = roots[j], roots[i]
			}
		}
	}

	totalBits := len(roots) * numValidators
	s.JustificationValidators = make([]byte, (totalBits+7)/8)

	for i, root := range roots {
		s.JustificationRoots = append(s.JustificationRoots, root)
		votes := justifications[root]
		startIdx := i * numValidators
		for j, voted := range votes {
			if voted {
				s.JustificationValidators = setBit(s.JustificationValidators, startIdx+j, true)
			}
		}
	}
}

// IsJustifiableSlot checks if a candidate slot is justifiable after a finalized slot.
// Per Devnet 0 spec: delta <= 5, or delta is x^2, or delta is x^2+x
func IsJustifiableSlot(finalizedSlot, candidate int) bool {
	if candidate < finalizedSlot {
		return false
	}
	delta := candidate - finalizedSlot
	if delta <= 5 {
		return true
	}
	// Check if delta is a perfect square (x^2)
	sqrtDelta := math.Sqrt(float64(delta))
	if sqrtDelta == math.Floor(sqrtDelta) {
		return true
	}
	// Check if delta is x^2+x (i.e., delta+0.25 has sqrt ending in 0.5)
	sqrtAdjusted := math.Sqrt(float64(delta) + 0.25)
	fractional := sqrtAdjusted - math.Floor(sqrtAdjusted)
	if math.Abs(fractional-0.5) < 1e-9 {
		return true
	}
	return false
}

// Bitlist helpers

// appendBitAt sets a bit at the given index, extending the slice if needed.
func appendBitAt(bits []byte, index int, val bool) []byte {
	byteIndex := index / 8
	for byteIndex >= len(bits) {
		bits = append(bits, 0)
	}
	if val {
		bits[byteIndex] |= 1 << (index % 8)
	}
	return bits
}

func getBit(bits []byte, index int) bool {
	if index/8 >= len(bits) {
		return false
	}
	return bits[index/8]&(1<<(index%8)) != 0
}

func setBit(bits []byte, index int, val bool) []byte {
	for index/8 >= len(bits) {
		bits = append(bits, 0)
	}
	if val {
		bits[index/8] |= 1 << (index % 8)
	} else {
		bits[index/8] &^= 1 << (index % 8)
	}
	return bits
}
