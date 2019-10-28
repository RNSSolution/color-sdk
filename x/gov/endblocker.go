package gov

import (
	"fmt"

	sdk "github.com/ColorPlatform/color-sdk/types"
	"github.com/ColorPlatform/color-sdk/x/gov/tags"
)

// EndBlocker Called every block, process inflation, update validator set
func EndBlocker(ctx sdk.Context, keeper Keeper) sdk.Tags {
	resTags := sdk.NewTags()
	currentFundingCycle, empty := keeper.GetCurrentCycle(ctx)
	fmt.Println("====Current Cycle =====", empty, currentFundingCycle)
	if (keeper.GetDaysPassed(ctx) >= LimitFirstFundingCycle) && empty {
		keeper.AddFundingCycle(ctx)
	} else if currentFundingCycle.CheckEqualEndTime(ctx.BlockHeader().Time) {
		ExecuteProposal(ctx, keeper, resTags)
		keeper.AddFundingCycle(ctx)
	}

	resTags = UpdateInactiveProposals(ctx, keeper, resTags)
	resTags = UpdateActiveProposals(ctx, keeper, resTags)
	return resTags
}

// UpdateInactiveProposals iteratee proposal and delete inactive proposal by time
func UpdateInactiveProposals(ctx sdk.Context, keeper Keeper, resTags sdk.Tags) sdk.Tags {
	logger := ctx.Logger().With("module", "x/gov")
	inactiveIterator := keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	defer inactiveIterator.Close()
	for ; inactiveIterator.Valid(); inactiveIterator.Next() {
		var proposalID uint64

		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(inactiveIterator.Value(), &proposalID)
		inactiveProposal, ok := keeper.GetProposal(ctx, proposalID)
		if !ok {
			panic(fmt.Sprintf("proposal %d does not exist", proposalID))
		}

		keeper.DeleteProposal(ctx, proposalID)
		keeper.DeleteDeposits(ctx, proposalID) // delete any associated deposits (burned)

		resTags = resTags.AppendTag(tags.ProposalID, fmt.Sprintf("%d", proposalID))
		resTags = resTags.AppendTag(tags.ProposalResult, tags.ActionProposalDropped)

		logger.Info(
			fmt.Sprintf("proposal %d (%s) didn't meet minimum deposit of %s (had only %s); deleted",
				inactiveProposal.ProposalID,
				inactiveProposal.GetTitle(),
				keeper.GetDepositParams(ctx).MinDeposit,
				inactiveProposal.TotalDeposit,
			),
		)
	}

	return resTags

}

// UpdateActiveProposals updates the prosal status on every block
func UpdateActiveProposals(ctx sdk.Context, keeper Keeper, resTags sdk.Tags) sdk.Tags {
	logger := ctx.Logger().With("module", "x/gov")
	// fetch active proposals whose voting periods have ended (are passed the block time)
	activeIterator := keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	defer activeIterator.Close()
	for ; activeIterator.Valid(); activeIterator.Next() {
		var proposalID uint64

		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(activeIterator.Value(), &proposalID)
		activeProposal, ok := keeper.GetProposal(ctx, proposalID)
		if !ok {
			panic(fmt.Sprintf("proposal %d does not exist", proposalID))
		}
		passes, tallyResults := tally(ctx, keeper, activeProposal)

		var tagValue = "In Voting Cycle"
		activeProposal.FinalTallyResult = tallyResults
		keeper.SetProposal(ctx, activeProposal)
		//	keeper.RemoveFromActiveProposalQueue(ctx, activeProposal.VotingEndTime, activeProposal.ProposalID)

		logger.Info(
			fmt.Sprintf(
				"proposal %d (%s) tallied; passed: %v",
				activeProposal.ProposalID, activeProposal.GetTitle(), passes,
			),
		)

		resTags = resTags.AppendTag(tags.ProposalID, fmt.Sprintf("%d", proposalID))
		resTags = resTags.AppendTag(tags.ProposalResult, tagValue)
	}
	return resTags
}

// ExecuteProposal Transfer funds on active proposal if voting end time reach
func ExecuteProposal(ctx sdk.Context, keeper Keeper, resTags sdk.Tags) sdk.Tags {
	logger := ctx.Logger().With("module", "x/gov")
	// fetch active proposals whose voting periods have ended (are passed the block time)
	activeIterator := keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	defer activeIterator.Close()
	for ; activeIterator.Valid(); activeIterator.Next() {
		var proposalID uint64

		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(activeIterator.Value(), &proposalID)
		activeProposal, ok := keeper.GetProposal(ctx, proposalID)
		if !ok {
			panic(fmt.Sprintf("proposal %d does not exist", proposalID))
		}
		passes, tallyResults := tally(ctx, keeper, activeProposal)

		var tagValue string

		// ===TODO check if no remaining cycle left then refund Deposit
		if passes && activeProposal.IsZeroRemainingCycle() {

			keeper.RefundDeposits(ctx, activeProposal.ProposalID)
			activeProposal.Status = StatusPassed
			tagValue = tags.ActionProposalPassed
		} else if !passes && activeProposal.IsZeroRemainingCycle() {

			keeper.DeleteDeposits(ctx, activeProposal.ProposalID)
			activeProposal.Status = StatusRejected
			tagValue = tags.ActionProposalRejected
		}

		activeProposal.FinalTallyResult = tallyResults
		keeper.SetProposal(ctx, activeProposal)

		// TODO check if no remaining cycle left then delete proposal
		//	keeper.RemoveFromActiveProposalQueue(ctx, activeProposal.VotingEndTime, activeProposal.ProposalID)

		logger.Info(
			fmt.Sprintf(
				"proposal %d (%s) tallied; passed: %v",
				activeProposal.ProposalID, activeProposal.GetTitle(), passes,
			),
		)

		resTags = resTags.AppendTag(tags.ProposalID, fmt.Sprintf("%d", proposalID))
		resTags = resTags.AppendTag(tags.ProposalResult, tagValue)
	}
	return resTags
}
