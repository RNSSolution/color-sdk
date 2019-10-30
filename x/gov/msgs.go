package gov

import (
	"fmt"

	sdk "github.com/ColorPlatform/color-sdk/types"
)

// Governance message types and routes
const (
	TypeMsgDeposit        = "deposit"
	TypeMsgVote           = "vote"
	TypeMsgSubmitProposal = "submit_proposal"

	MaxDescriptionLength int = 5000
	MaxTitleLength       int = 140
)

var _, _, _ sdk.Msg = MsgSubmitProposal{}, MsgDeposit{}, MsgVote{}

// MsgSubmitProposal
type MsgSubmitProposal struct {
	Title          string         `json:"title"`           //  Title of the proposal
	Description    string         `json:"description"`     //  Description of the proposal
	ProposalType   ProposalKind   `json:"proposal_type"`   //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
	Proposer       sdk.AccAddress `json:"proposer"`        //  Address of the proposer
	InitialDeposit sdk.Coins      `json:"initial_deposit"` //  Initial deposit paid by sender. Must be strictly positive.
	RequestedFund  sdk.Coins      `json:"requested_fund"`  //  Requested Proposal Fund
	FundingCycle   uint64         `json:"fund_cycle"`      //  Fund Cycle
}

func NewMsgSubmitProposal(title, description string, proposalType ProposalKind, proposer sdk.AccAddress, initialDeposit, requestedFund sdk.Coins, fundingcycle uint64) MsgSubmitProposal {
	return MsgSubmitProposal{
		Title:          title,
		Description:    description,
		ProposalType:   proposalType,
		Proposer:       proposer,
		InitialDeposit: initialDeposit,
		RequestedFund:  requestedFund,
		FundingCycle:   fundingcycle,
	}
}

//nolint
func (msg MsgSubmitProposal) Route() string { return RouterKey }
func (msg MsgSubmitProposal) Type() string  { return TypeMsgSubmitProposal }

// Implements Msg.
func (msg MsgSubmitProposal) ValidateBasic() sdk.Error {
	//	var MaxLimit sdk.Int = sdk.NewInt(1209600000000)
	//	var FeeLimit sdk.Int = sdk.NewInt(10000000000)
	if len(msg.Title) == 0 {
		return ErrInvalidTitle(DefaultCodespace, "No title present in proposal")
	}
	if len(msg.Title) > MaxTitleLength {
		return ErrInvalidTitle(DefaultCodespace, fmt.Sprintf("Proposal title is longer than max length of %d", MaxTitleLength))
	}
	if len(msg.Description) == 0 {
		return ErrInvalidDescription(DefaultCodespace, "No description present in proposal")
	}
	if len(msg.Description) > MaxDescriptionLength {
		return ErrInvalidDescription(DefaultCodespace, fmt.Sprintf("Proposal description is longer than max length of %d", MaxDescriptionLength))
	}
	if !validProposalType(msg.ProposalType) {
		return ErrInvalidProposalType(DefaultCodespace, msg.ProposalType)
	}
	if msg.Proposer.Empty() {
		return sdk.ErrInvalidAddress(msg.Proposer.String())
	}
	if !msg.InitialDeposit.IsValid() {
		return sdk.ErrInvalidCoins(msg.InitialDeposit.String())
	}
	if msg.InitialDeposit.IsAnyNegative() {
		return sdk.ErrInvalidCoins(msg.InitialDeposit.String())
	}
	// if len(msg.RequestedFund.String()) == 0 {
	// 	return sdk.ErrInvalidCoins(msg.RequestedFund.String())
	// }
	// if msg.RequestedFund.AmountOf(sdk.DefaultBondDenom).IsZero() {
	// 	return sdk.ErrUnauthorized("Value Should be in uclr")
	// }
	// if (msg.RequestedFund.AmountOf(sdk.DefaultBondDenom)).GT(MaxLimit) {
	// 	return sdk.ErrUnauthorized("Value Should not be greater than 1209600000000")
	// }
	// if !msg.RequestedFund.IsValid() {
	// 	return sdk.ErrInvalidCoins(msg.RequestedFund.String())
	// }
	// if msg.RequestedFund.IsAnyNegative() {
	// 	return sdk.ErrInvalidCoins(msg.RequestedFund.String())
	// }
	// if msg.FundingCycle == 0 {
	// 	return sdk.ErrUnauthorized("Zero cycle is not allowed")
	// }
	// if msg.FundingCycle > 6 {
	// 	return sdk.ErrUnauthorized("Fund cycle more than 6 is not allowed")
	// }
	// if msg.InitialDeposit.AmountOf(sdk.DefaultBondDenom).LT(FeeLimit) {
	// 	return sdk.ErrUnauthorized("Minimum Deposit fee should be 10,000 CLR")
	// }
	return nil
}

func (msg MsgSubmitProposal) String() string {
	return fmt.Sprintf("MsgSubmitProposal{%s, %s, %s, %v,%v}", msg.Title, msg.Description, msg.ProposalType, msg.InitialDeposit, msg.RequestedFund)
}

// Implements Msg.
func (msg MsgSubmitProposal) GetSignBytes() []byte {
	bz := msgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Proposer}
}

// MsgDeposit
type MsgDeposit struct {
	ProposalID uint64         `json:"proposal_id"` // ID of the proposal
	Depositor  sdk.AccAddress `json:"depositor"`   // Address of the depositor
	Amount     sdk.Coins      `json:"amount"`      // Coins to add to the proposal's deposit
}

func NewMsgDeposit(depositor sdk.AccAddress, proposalID uint64, amount sdk.Coins) MsgDeposit {
	return MsgDeposit{
		ProposalID: proposalID,
		Depositor:  depositor,
		Amount:     amount,
	}
}

// Implements Msg.
// nolint
func (msg MsgDeposit) Route() string { return RouterKey }
func (msg MsgDeposit) Type() string  { return TypeMsgDeposit }

// Implements Msg.
func (msg MsgDeposit) ValidateBasic() sdk.Error {
	if msg.Depositor.Empty() {
		return sdk.ErrInvalidAddress(msg.Depositor.String())
	}
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}
	if msg.Amount.IsAnyNegative() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}
	if msg.ProposalID < 0 {
		return ErrUnknownProposal(DefaultCodespace, msg.ProposalID)
	}
	return nil
}

func (msg MsgDeposit) String() string {
	return fmt.Sprintf("MsgDeposit{%s=>%v: %v}", msg.Depositor, msg.ProposalID, msg.Amount)
}

// Implements Msg.
func (msg MsgDeposit) GetSignBytes() []byte {
	bz := msgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Depositor}
}

// MsgVote
type MsgVote struct {
	ProposalID uint64         `json:"proposal_id"` // ID of the proposal
	Voter      sdk.AccAddress `json:"voter"`       //  address of the voter
	Option     VoteOption     `json:"option"`      //  option from OptionSet chosen by the voter
}

func NewMsgVote(voter sdk.AccAddress, proposalID uint64, option VoteOption) MsgVote {
	return MsgVote{
		ProposalID: proposalID,
		Voter:      voter,
		Option:     option,
	}
}

// Implements Msg.
// nolint
func (msg MsgVote) Route() string { return RouterKey }
func (msg MsgVote) Type() string  { return TypeMsgVote }

// Implements Msg.
func (msg MsgVote) ValidateBasic() sdk.Error {
	if msg.Voter.Empty() {
		return sdk.ErrInvalidAddress(msg.Voter.String())
	}
	if msg.ProposalID < 0 {
		return ErrUnknownProposal(DefaultCodespace, msg.ProposalID)
	}
	if !validVoteOption(msg.Option) {
		return ErrInvalidVote(DefaultCodespace, msg.Option)
	}
	return nil
}

func (msg MsgVote) String() string {
	return fmt.Sprintf("MsgVote{%v - %s}", msg.ProposalID, msg.Option)
}

// Implements Msg.
func (msg MsgVote) GetSignBytes() []byte {
	bz := msgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgVote) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Voter}
}
