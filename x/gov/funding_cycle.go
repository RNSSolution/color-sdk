package gov

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	sdk "github.com/ColorPlatform/color-sdk/types"
)

const (
	// FirstBlockHeight condation of block heigh reach to 1
	FirstBlockHeight = 1
	// LimitFirstFundingCycle condation first funding cycle should start after 4 weeks
	LimitFirstFundingCycle = 0
	// FourWeeksHours calculate total hours in 4 weeks
	FourWeeksHours        = time.Hour * time.Duration(24*28)
	StopFundingBeforeDays = 0 //stop on last two days of funding cycle
	DefaultBondDenom      = sdk.DefaultBondDenom
)

// FundingCycle controlling proposal cycles
type FundingCycle struct {
	CycleID         uint64    `json:"cycle_id"`         //  ID of the proposal
	CycleStartTime  time.Time `json:"cycle_start_time"` //  Time of the funding cycle to start
	CycleEndTime    time.Time `json:"cycle_end_time"`   //  Time that the funding cycle to end
	FundedProposals []uint64  `json:"funded_proposals"` // Funded proposals in a funding cycle
}

func (fs FundingCycle) String() string {
	return fmt.Sprintf(`
	CycleID:                    %d
	Cycle Start Time:           %s
	Cycle End Time:             %s
	Funded Proposals: 	%s
`,
		fs.CycleID, fs.CycleStartTime, fs.CycleEndTime, fs.FundedProposals,
	)
}

// CheckEqualEndTime Check Current Time of Blockchain
func (fs FundingCycle) CheckEqualEndTime(currentTime time.Time) bool {
	if currentTime.After(fs.CycleEndTime) {
		return true
	}
	return false

}
func GetPercentageAmount(amount sdk.Dec, percentage float64) int64 {
	num1, _ := strconv.ParseFloat(amount.String(), 64)
	percentage = percentage * num1
	return int64(percentage)

}

type FundingCycles []FundingCycle

// nolint
func (fs FundingCycles) String() string {
	out := "ID - [StartTime] [EndTime]\n"
	for _, cycle := range fs {
		out += fmt.Sprintf("%d - [%s] [%s]\n",
			cycle.CycleID, cycle.CycleStartTime, cycle.CycleEndTime)
	}
	return strings.TrimSpace(out)
}

type ProposalEligibility struct {
	ProposalID    uint64    `json:"proposal_id"` //  ID of the proposal
	VotesCount    sdk.Int   `json:"votes_count"` //  rank of the proposal
	RequestedFund sdk.Coins `json:"votes_count"` //  rank of the proposal
}

func NewEligibilityDetails(proposalID uint64, votes sdk.Int, requestedFund sdk.Coins) ProposalEligibility {

	var e ProposalEligibility
	e.ProposalID = proposalID
	e.VotesCount = votes
	e.RequestedFund = requestedFund
	return e

}
func VerifyAmount(totalRequested sdk.Coins, limit sdk.Int) bool {
	ts := totalRequested.AmountOf(DefaultBondDenom)
	if ts.LTE(limit) {
		return true
	} else {
		return false
	}

}

func SortProposalEligibility(proposals []Proposal, results []TallyResult) []Proposal {
	sort.Slice(proposals, func(i, j int) bool {
		votesA := results[i].Yes.Sub(results[i].No)
		votesB := results[i].Yes.Sub(results[i].No)

		return votesA.GT(votesB)
	})
	return proposals
}

//CheckCycleActive Stop Funding on last two days of Funding Cycle
func (keeper Keeper) CheckCycleActive(ctx sdk.Context) bool {
	currentFundingCycle, err := keeper.GetCurrentCycle(ctx)
	if err == nil {
		timeblock := ctx.BlockHeader().Time
		if !timeblock.After(currentFundingCycle.CycleEndTime.AddDate(0, 0, StopFundingBeforeDays)) {
			return true
		}
		return false
	}
	return false
}
