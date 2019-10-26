package gov

import (
	"time"
)

const (
	// FirstBlockHeight condation of block heigh reach to 1
	FirstBlockHeight = 1
	// LimitFirstFundingCycle condation first funding cycle should start after 4 weeks
	LimitFirstFundingCycle = 28
	// FourWeeksHours calculate total hours in 4 weeks
	FourWeeksHours = time.Hour * time.Duration(24*28)
)

// FundingCycle controlling proposal cycles
type FundingCycle struct {
	CycleID        uint64    `json:"proposal_id"`      //  ID of the proposal
	CycleStartTime time.Time `json:"cycle_start_time"` //  Time of the funding cycle to start
	CycleEndTime   time.Time `json:"cycle_end_time"`   //  Time that the funding cycle to end
}

// CheckEqualEndTime Peeks the next available ProposalID without incrementing it
func (fs FundingCycle) CheckEqualEndTime(currentTime time.Time) bool {
	if currentTime.Equal(fs.CycleEndTime) {
		return true
	}
	return false

}
