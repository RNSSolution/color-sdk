package gov

import (
	"fmt"

	sdk "github.com/ColorPlatform/color-sdk/types"
	"github.com/ColorPlatform/color-sdk/x/gov/tags"
)

// Handle all "gov" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgDeposit:
			return handleMsgDeposit(ctx, keeper, msg)
		case MsgSubmitProposal:
			return handleMsgSubmitProposal(ctx, keeper, msg)
		case MsgVote:
			return handleMsgVote(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized gov msg type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSubmitProposal(ctx sdk.Context, keeper Keeper, msg MsgSubmitProposal) sdk.Result {
	var content ProposalContent
	if ExpectedTreasureIncome(keeper, ctx, msg.RequestedFund.AmountOf(sdk.DefaultBondDenom)) {
		return ErrInvalidTreasureIncome(keeper.codespace, msg.ProposalType).Result()
	}
	switch msg.ProposalType {
	case ProposalTypeText:
		content = NewTextProposal(msg.Title, msg.Description, msg.RequestedFund, msg.FundingCycle, msg.Proposer)
	case ProposalTypeSoftwareUpgrade:
		content = NewSoftwareUpgradeProposal(msg.Title, msg.Description, msg.RequestedFund, msg.FundingCycle, msg.Proposer)
	default:
		return ErrInvalidProposalType(keeper.codespace, msg.ProposalType).Result()
	}
	proposal, err := keeper.SubmitProposal(ctx, content)
	if err != nil {
		return err.Result()
	}
	proposalID := proposal.ProposalID
	proposalIDStr := fmt.Sprintf("%d", proposalID)

	err, votingStarted := keeper.AddDeposit(ctx, proposalID, msg.Proposer, msg.InitialDeposit)
	if err != nil {
		return err.Result()
	}

	resTags := sdk.NewTags(
		tags.Proposer, []byte(msg.Proposer.String()),
		tags.ProposalID, proposalIDStr,
	)

	if votingStarted {
		resTags = resTags.AppendTag(tags.VotingPeriodStart, proposalIDStr)
	}

	return sdk.Result{
		Data: keeper.cdc.MustMarshalBinaryLengthPrefixed(proposalID),
		Tags: resTags,
	}
}

func handleMsgDeposit(ctx sdk.Context, keeper Keeper, msg MsgDeposit) sdk.Result {
	err, votingStarted := keeper.AddDeposit(ctx, msg.ProposalID, msg.Depositor, msg.Amount)
	if err != nil {
		return err.Result()
	}

	proposalIDStr := fmt.Sprintf("%d", msg.ProposalID)
	resTags := sdk.NewTags(
		tags.Depositor, []byte(msg.Depositor.String()),
		tags.ProposalID, proposalIDStr,
	)

	if votingStarted {
		resTags = resTags.AppendTag(tags.VotingPeriodStart, proposalIDStr)
	}

	return sdk.Result{
		Tags: resTags,
	}
}

func handleMsgVote(ctx sdk.Context, keeper Keeper, msg MsgVote) sdk.Result {
	err := keeper.AddVote(ctx, msg.ProposalID, msg.Voter, msg.Option)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{
		Tags: sdk.NewTags(
			tags.Voter, msg.Voter.String(),
			tags.ProposalID, fmt.Sprintf("%d", msg.ProposalID),
		),
	}
}
