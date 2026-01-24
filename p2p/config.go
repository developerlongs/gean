// Package p2p implements networking for the Lean Ethereum consensus protocol.
package p2p

// Domain types for message ID isolation (per networking spec)
var (
	MessageDomainInvalidSnappy = [4]byte{0x00, 0x00, 0x00, 0x00}
	MessageDomainValidSnappy   = [4]byte{0x01, 0x00, 0x00, 0x00}
)
