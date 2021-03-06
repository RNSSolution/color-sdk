package gov

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/ColorPlatform/color-sdk/types"
)

const (
	//MaxCycleCount store the max limit for proposal
	MaxCycleCount = 2
)

// Proposal is a struct used by gov module internally
// embedds ProposalContent with additional fields to record the status of the proposal process
type Proposal struct {
	ProposalContent `json:"proposal_content"` // Proposal content interface

	ProposalID uint64 `json:"proposal_id"` //  ID of the proposal

	Status                ProposalStatus `json:"proposal_status"`         //  Status of the Proposal {Pending, Active, Passed, Rejected}
	FinalTallyResult      TallyResult    `json:"final_tally_result"`      //  Result of Tallys
	SubmitTime            time.Time      `json:"submit_time"`             //  Time of the block where TxGovSubmitProposal was included
	DepositEndTime        time.Time      `json:"deposit_end_time"`        // Time that the Proposal would expire if deposit amount isn't met
	TotalDeposit          sdk.Coins      `json:"total_deposit"`           //  Current deposit on this proposal. Initial value is set at InitialDeposit	RequestedFund   sdk.Coins `json:"requested_fund"`    //  Fund Requested
	RemainingFundingCycle uint64         `json:"remaining_funding_cycle"` //   Remaining Funding Cycle
	FundingCycleCount     uint64         `json:"funding_cycle_count"`     //   Remaining Funding Cycle
	Ranking               sdk.Int        `json:"ranking"`                 //   Remaining Funding Cycle
	VotingStartTime       time.Time      `json:"voting_start_time"`       //  Time of the block where MinDeposit was reached. -1 if MinDeposit is not reached
	VotingEndTime         time.Time      `json:"voting_end_time"`         // Time that the VotingPeriod for this proposal will end and votes will be tallied
}

// nolint
func (p Proposal) String() string {
	return fmt.Sprintf(`Proposal %d:
  Title:              %s
  Ranking             %s
  Type:               %s
  Status:             %s
  Submit Time:        %s
  Deposit End Time:   %s
  Total Deposit:      %s
  Requested Fund:     %s
  Funding Cycle:      %d
  Voting Start Time:  %s
  Voting End Time:    %s
  Description:        %s`,
		p.ProposalID, p.GetTitle(), p.Ranking, p.ProposalType(),
		p.Status, p.SubmitTime, p.DepositEndTime,
		p.TotalDeposit, p.GetRequestedFund(), p.GetFundingCycle(), p.VotingStartTime, p.VotingEndTime, p.GetDescription(),
	)
}

func (p Proposal) IsZeroRemainingCycle() bool {
	return p.RemainingFundingCycle == 0
}

func (p Proposal) CheckMaxCycleCount() bool {
	return p.FundingCycleCount == MaxCycleCount
}
func (p Proposal) ReduceCycleCount() Proposal {

	if p.RemainingFundingCycle <= 0 {
		panic("Remaining funding Cycle cannot be less then zero")
	} else {
		p.RemainingFundingCycle = p.RemainingFundingCycle - 1
	}
	return p
}

// ProposalContent is an interface that has title, description, and proposaltype
// that the governance module can use to identify them and generate human readable messages
// ProposalContent can have additional fields, which will handled by ProposalHandlers
// via type assertion, e.g. parameter change amount in ParameterChangeProposal
type ProposalContent interface {
	GetTitle() string
	GetDescription() string
	ProposalType() ProposalKind
	GetRequestedFund() sdk.Coins
	GetFundingCycle() uint64
	GetProposer() sdk.AccAddress
}

// Proposals is an array of proposal
type Proposals []Proposal

// nolint
func (p Proposals) String() string {
	out := "ID - (Status) [Type] Title\n"
	for _, prop := range p {
		out += fmt.Sprintf("%d - (%s) [%s] %s\n",
			prop.ProposalID, prop.Status,
			prop.ProposalType(), prop.GetTitle())
	}
	return strings.TrimSpace(out)
}

// Text Proposals
type TextProposal struct {
	Title         string         `json:"title"`          //  Title of the proposal
	Description   string         `json:"description"`    //  Description of the proposal
	RequestedFund sdk.Coins      `json:"requested_fund"` // Requested Funds in Proposal
	FundingCycle  uint64         `json:"funding_cycle"`  // Funding Cycle
	Proposer      sdk.AccAddress `json:"proposer"`       //  Address of the proposer
}

func NewTextProposal(title, description string, requestfund sdk.Coins, fundingcycle uint64, proposer sdk.AccAddress) TextProposal {
	return TextProposal{
		Title:         title,
		Description:   description,
		RequestedFund: requestfund,
		FundingCycle:  fundingcycle,
		Proposer:      proposer,
	}
}

// Implements Proposal Interface
var _ ProposalContent = TextProposal{}

// nolint
func (tp TextProposal) GetTitle() string            { return tp.Title }
func (tp TextProposal) GetDescription() string      { return tp.Description }
func (tp TextProposal) ProposalType() ProposalKind  { return ProposalTypeText }
func (tp TextProposal) GetRequestedFund() sdk.Coins { return tp.RequestedFund }
func (tp TextProposal) GetFundingCycle() uint64     { return tp.FundingCycle }
func (tp TextProposal) GetProposer() sdk.AccAddress { return tp.Proposer }

// Software Upgrade Proposals
type SoftwareUpgradeProposal struct {
	TextProposal
}

func NewSoftwareUpgradeProposal(title, description string, requestfund sdk.Coins, fundingcycle uint64, proposer sdk.AccAddress) SoftwareUpgradeProposal {
	return SoftwareUpgradeProposal{
		TextProposal: NewTextProposal(title, description, requestfund, fundingcycle, proposer),
	}
}

// Implements Proposal Interface
var _ ProposalContent = SoftwareUpgradeProposal{}

// nolint
func (sup SoftwareUpgradeProposal) ProposalType() ProposalKind { return ProposalTypeSoftwareUpgrade }

// ProposalQueue
type ProposalQueue []uint64

// ProposalKind

// Type that represents Proposal Type as a byte
type ProposalKind byte

//nolint
const (
	ProposalTypeNil             ProposalKind = 0x00
	ProposalTypeText            ProposalKind = 0x01
	ProposalTypeParameterChange ProposalKind = 0x02
	ProposalTypeSoftwareUpgrade ProposalKind = 0x03
)

// String to proposalType byte. Returns 0xff if invalid.
func ProposalTypeFromString(str string) (ProposalKind, error) {
	switch str {
	case "Text":
		return ProposalTypeText, nil
	case "ParameterChange":
		return ProposalTypeParameterChange, nil
	case "SoftwareUpgrade":
		return ProposalTypeSoftwareUpgrade, nil
	default:
		return ProposalKind(0xff), fmt.Errorf("'%s' is not a valid proposal type", str)
	}
}

// is defined ProposalType?
func validProposalType(pt ProposalKind) bool {
	if pt == ProposalTypeText ||
		pt == ProposalTypeParameterChange ||
		pt == ProposalTypeSoftwareUpgrade {
		return true
	}
	return false
}

// Marshal needed for protobuf compatibility
func (pt ProposalKind) Marshal() ([]byte, error) {
	return []byte{byte(pt)}, nil
}

// Unmarshal needed for protobuf compatibility
func (pt *ProposalKind) Unmarshal(data []byte) error {
	*pt = ProposalKind(data[0])
	return nil
}

// Marshals to JSON using string
func (pt ProposalKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(pt.String())
}

// Unmarshals from JSON assuming Bech32 encoding
func (pt *ProposalKind) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bz2, err := ProposalTypeFromString(s)
	if err != nil {
		return err
	}
	*pt = bz2
	return nil
}

// Turns VoteOption byte to String
func (pt ProposalKind) String() string {
	switch pt {
	case ProposalTypeText:
		return "Text"
	case ProposalTypeParameterChange:
		return "ParameterChange"
	case ProposalTypeSoftwareUpgrade:
		return "SoftwareUpgrade"
	default:
		return ""
	}
}

// For Printf / Sprintf, returns bech32 when using %s
// nolint: errcheck
func (pt ProposalKind) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(pt.String()))
	default:
		// TODO: Do this conversion more directly
		s.Write([]byte(fmt.Sprintf("%v", byte(pt))))
	}
}

// ProposalStatus

// Type that represents Proposal Status as a byte
type ProposalStatus byte

//nolint
const (
	StatusNil           ProposalStatus = 0x00
	StatusDepositPeriod ProposalStatus = 0x01
	StatusVotingPeriod  ProposalStatus = 0x02
	StatusPassed        ProposalStatus = 0x03
	StatusRejected      ProposalStatus = 0x04
)

// ProposalStatusToString turns a string into a ProposalStatus
func ProposalStatusFromString(str string) (ProposalStatus, error) {
	switch str {
	case "DepositPeriod":
		return StatusDepositPeriod, nil
	case "VotingPeriod":
		return StatusVotingPeriod, nil
	case "Passed":
		return StatusPassed, nil
	case "Rejected":
		return StatusRejected, nil
	case "":
		return StatusNil, nil
	default:
		return ProposalStatus(0xff), fmt.Errorf("'%s' is not a valid proposal status", str)
	}
}

// is defined ProposalType?
func validProposalStatus(status ProposalStatus) bool {
	if status == StatusDepositPeriod ||
		status == StatusVotingPeriod ||
		status == StatusPassed ||
		status == StatusRejected {
		return true
	}
	return false
}

// Marshal needed for protobuf compatibility
func (status ProposalStatus) Marshal() ([]byte, error) {
	return []byte{byte(status)}, nil
}

// Unmarshal needed for protobuf compatibility
func (status *ProposalStatus) Unmarshal(data []byte) error {
	*status = ProposalStatus(data[0])
	return nil
}

// Marshals to JSON using string
func (status ProposalStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(status.String())
}

// Unmarshals from JSON assuming Bech32 encoding
func (status *ProposalStatus) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bz2, err := ProposalStatusFromString(s)
	if err != nil {
		return err
	}
	*status = bz2
	return nil
}

// Turns VoteStatus byte to String
func (status ProposalStatus) String() string {
	switch status {
	case StatusDepositPeriod:
		return "DepositPeriod"
	case StatusVotingPeriod:
		return "VotingPeriod"
	case StatusPassed:
		return "Passed"
	case StatusRejected:
		return "Rejected"
	default:
		return ""
	}
}

// For Printf / Sprintf, returns bech32 when using %s
// nolint: errcheck
func (status ProposalStatus) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(status.String()))
	default:
		// TODO: Do this conversion more directly
		s.Write([]byte(fmt.Sprintf("%v", byte(status))))
	}
}

// Tally Results
type TallyResult struct {
	Yes     sdk.Int `json:"yes"`
	Abstain sdk.Int `json:"abstain"`
	No      sdk.Int `json:"no"`
}

func NewTallyResult(yes, abstain, no sdk.Int) TallyResult {
	return TallyResult{
		Yes:     yes,
		Abstain: abstain,
		No:      no,
	}
}

func NewTallyResultFromMap(results map[VoteOption]sdk.Dec) TallyResult {
	return TallyResult{
		Yes:     results[OptionYes].TruncateInt(),
		Abstain: results[OptionAbstain].TruncateInt(),
		No:      results[OptionNo].TruncateInt(),
	}
}

// checks if two proposals are equal
func EmptyTallyResult() TallyResult {
	return TallyResult{
		Yes:     sdk.ZeroInt(),
		Abstain: sdk.ZeroInt(),
		No:      sdk.ZeroInt(),
	}
}

// checks if two proposals are equal
func (tr TallyResult) Equals(comp TallyResult) bool {
	return (tr.Yes.Equal(comp.Yes) &&
		tr.Abstain.Equal(comp.Abstain) &&
		tr.No.Equal(comp.No))
}

func (tr TallyResult) String() string {
	return fmt.Sprintf(`Tally Result:
  Yes:        %s
  Abstain:    %s
  No:         %s`, tr.Yes, tr.Abstain, tr.No)
}

///ExpectedTreasureIncome Calculate Funding requested must be no more than 50% of Treasury income per cycle
func ExpectedTreasureIncome(keeper Keeper, ctx sdk.Context, Requestedfund sdk.Int) bool {
	limit := keeper.GetTreasuryWeeklyIncome(ctx)
	formula := 0.5 //Deduct 5%
	num1, _ := strconv.ParseFloat(limit.String(), 64)
	formula = formula * num1
	var treasuryIncome sdk.Int = sdk.NewInt(int64(formula))
	if !(Requestedfund.LT(treasuryIncome)) {
		return true
	}
	return false
}
