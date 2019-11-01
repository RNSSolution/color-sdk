package gov

import (
	"sort"
	"strconv"
	"time"

	sdk "github.com/ColorPlatform/color-sdk/types"
)

const (
	// FirstBlockHeight condation of block heigh reach to 1
	FirstBlockHeight = 1
	// LimitFirstFundingCycle condation first funding cycle should start after 4 weeks
	LimitFirstFundingCycle = 28
	// FourWeeksHours calculate total hours in 4 weeks
	FourWeeksHours = time.Hour * time.Duration(24*28)

	DefaultBondDenom = "uclr"
)

// FundingCycle controlling proposal cycles
type FundingCycle struct {
	CycleID        uint64    `json:"cycle_id"`         //  ID of the proposal
	CycleStartTime time.Time `json:"cycle_start_time"` //  Time of the funding cycle to start
	CycleEndTime   time.Time `json:"cycle_end_time"`   //  Time that the funding cycle to end
}

// CheckEqualEndTime Peeks the next available ProposalID without incrementing it
func (fs FundingCycle) CheckEqualEndTime(currentTime time.Time) bool {
	if currentTime.After(fs.CycleEndTime) {
		return true
	}
	return false

}
func GetPercentageAmount(amount sdk.Dec, percentage float64) sdk.Dec {
	num1, _ := strconv.ParseFloat(amount.String(), 64)
	percentage = percentage * num1
	return sdk.NewDec(int64(percentage))

}

type ProposalEligibility struct {
	ProposalID uint64 `json:"proposal_id"` //  ID of the proposal
	Rank       uint64 `json:"rank"`        //  rank of the proposal
}
type EligibilityDetails struct {
	ProposalID    uint64    `json:"proposal_id"` //  ID of the proposal
	VotesCount    sdk.Int   `json:"votes_count"` //  rank of the proposal
	RequestedFund sdk.Coins `json:"votes_count"` //  rank of the proposal
}

func Append(eligibilityList []EligibilityDetails, eligibility EligibilityDetails) []EligibilityDetails {
	return append(eligibilityList, eligibility)
}

func NewEligibilityDetails(proposalID uint64, votes sdk.Int, requestedFund sdk.Coins) EligibilityDetails {

	var e EligibilityDetails
	e.ProposalID = proposalID
	e.VotesCount = votes
	e.RequestedFund = requestedFund
	return e

}
func VerifyAmount(totalRequested sdk.Coins, limit sdk.Dec) bool {
	ts := totalRequested.AmountOf(DefaultBondDenom).Int64()
	l := limit.Int64()
	if l <= ts {
		return true
	} else {
		return false
	}

}

func SortProposalEligibility(eligibilityList []EligibilityDetails) []EligibilityDetails {
	sort.Slice(eligibilityList, func(i, j int) bool {
		return eligibilityList[i].VotesCount.GT(eligibilityList[j].VotesCount)
	})
	return eligibilityList
}
