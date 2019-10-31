package gov

import (
	"bytes"
	"fmt"
	"time"

	sdk "github.com/ColorPlatform/color-sdk/types"
)

// Key for getting a the next available proposalID from the store
var (
	KeyDelimiter = []byte(":")

	KeyFirstBlock               = []byte("firstblock")
	KeyNextProposalID           = []byte("newProposalID")
	KeyNextFundingCycleID       = []byte("newFundingCycleID")
	PrefixActiveProposalQueue   = []byte("activeProposalQueue")
	PrefixInactiveProposalQueue = []byte("inactiveProposalQueue")
	PrefixFudingCycleQueue      = []byte("fundingCycles")
	PrefixEligibilityQueue      = []byte("proposalEligibility")
)

// Key for getting a specific proposal from the store
func KeyProposal(proposalID uint64) []byte {
	return []byte(fmt.Sprintf("proposals:%d", proposalID))
}

// Key for getting a specific proposal from the store
func PrefixFudingCycle() []byte {
	return bytes.Join([][]byte{
		PrefixFudingCycleQueue,
	}, KeyDelimiter)
}

// Key for getting a specific proposal from the store
func KeyFundingCycle(fundingCycle uint64) []byte {
	return bytes.Join([][]byte{
		PrefixFudingCycleQueue,
		[]byte(fmt.Sprintf("cycle:%d", fundingCycle)),
		sdk.Uint64ToBigEndian(fundingCycle),
	}, KeyDelimiter)
}

// Key for getting a specific proposal from the store
func KeyEligibility(eligibilityID uint64) []byte {
	return bytes.Join([][]byte{
		PrefixEligibilityQueue,
		[]byte(fmt.Sprintf("proposalEligibility:%d", eligibilityID)),
		sdk.Uint64ToBigEndian(eligibilityID),
	}, KeyDelimiter)
}

// Key for getting a specific deposit from the store
func KeyDeposit(proposalID uint64, depositorAddr sdk.AccAddress) []byte {
	return []byte(fmt.Sprintf("deposits:%d:%d", proposalID, depositorAddr))
}

// Key for getting a specific vote from the store
func KeyVote(proposalID uint64, voterAddr sdk.AccAddress) []byte {
	return []byte(fmt.Sprintf("votes:%d:%d", proposalID, voterAddr))
}

// Key for getting all deposits on a proposal from the store
func KeyDepositsSubspace(proposalID uint64) []byte {
	return []byte(fmt.Sprintf("deposits:%d:", proposalID))
}

// Key for getting all votes on a proposal from the store
func KeyVotesSubspace(proposalID uint64) []byte {
	return []byte(fmt.Sprintf("votes:%d:", proposalID))
}

// Returns the key for a proposalID in the activeProposalQueue
func PrefixActiveProposalQueueTime(endTime time.Time) []byte {
	return bytes.Join([][]byte{
		PrefixActiveProposalQueue,
		sdk.FormatTimeBytes(endTime),
	}, KeyDelimiter)
}

// Returns the key for a proposalID in the activeProposalQueue
func KeyActiveProposalQueueProposal(endTime time.Time, proposalID uint64) []byte {
	return bytes.Join([][]byte{
		PrefixActiveProposalQueue,
		sdk.FormatTimeBytes(endTime),
		sdk.Uint64ToBigEndian(proposalID),
	}, KeyDelimiter)
}

// Returns the key for a proposalID in the activeProposalQueue
func PrefixInactiveProposalQueueTime(endTime time.Time) []byte {
	return bytes.Join([][]byte{
		PrefixInactiveProposalQueue,
		sdk.FormatTimeBytes(endTime),
	}, KeyDelimiter)
}

// Returns the key for a proposalID in the activeProposalQueue
func KeyInactiveProposalQueueProposal(endTime time.Time, proposalID uint64) []byte {
	return bytes.Join([][]byte{
		PrefixInactiveProposalQueue,
		sdk.FormatTimeBytes(endTime),
		sdk.Uint64ToBigEndian(proposalID),
	}, KeyDelimiter)
}
