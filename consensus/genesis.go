package consensus

// GenerateGenesis creates a genesis state with the given parameters.
func GenerateGenesis(genesisTime, numValidators uint64) *State {
	emptyBody := &BlockBody{Attestations: []SignedVote{}}
	bodyRoot, _ := emptyBody.HashTreeRoot()

	genesisHeader := BlockHeader{
		Slot:          0,
		ProposerIndex: 0,
		ParentRoot:    Root{},
		StateRoot:     Root{},
		BodyRoot:      bodyRoot,
	}

	return &State{
		Config: Config{
			NumValidators: numValidators,
			GenesisTime:   genesisTime,
		},
		Slot:                    0,
		LatestBlockHeader:       genesisHeader,
		LatestJustified:         Checkpoint{Root: Root{}, Slot: 0},
		LatestFinalized:         Checkpoint{Root: Root{}, Slot: 0},
		HistoricalBlockHashes:   []Root{},
		JustifiedSlots:          []byte{},
		JustificationRoots:      []Root{},
		JustificationValidators: []byte{},
	}
}

// IsProposer checks if a validator is the proposer for the current slot.
func (s *State) IsProposer(validatorIndex ValidatorIndex) bool {
	return uint64(s.Slot)%s.Config.NumValidators == uint64(validatorIndex)
}
