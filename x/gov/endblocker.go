package gov

import (
	"fmt"

	sdk "github.com/ColorPlatform/color-sdk/types"
	"github.com/ColorPlatform/color-sdk/x/gov/tags"
)

// EndBlocker Called every block, process inflation, update validator set
func EndBlocker(ctx sdk.Context, keeper Keeper) sdk.Tags {

	resTags := sdk.NewTags()
	currentFundingCycle, err := keeper.GetCurrentCycle(ctx)
	if err != nil {
		days, err := keeper.GetDaysPassed(ctx)
		if err != nil {
			keeper.SetBlockTime(ctx)
		} else if days >= LimitFirstFundingCycle {
			keeper.RemoveFromInactiveProposalQueueIterator(ctx)
			keeper.AddFundingCycle(ctx)
		}

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
	proposals := []Proposal{}
	results := []TallyResult{}
	// fetch active proposals whose voting periods have ended (are passed the block time)
	activeIterator := keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	var tagValue = "In Voting Cycle"
	defer activeIterator.Close()
	for ; activeIterator.Valid(); activeIterator.Next() {
		var proposalID uint64
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(activeIterator.Value(), &proposalID)

		activeProposal, ok := keeper.GetProposal(ctx, proposalID)
		if !ok {
			panic(fmt.Sprintf("proposal %d does not exist", proposalID))
		}
		passes, tallyResults, netural := tally(ctx, keeper, activeProposal)

		if passes {
			proposals = append(proposals, activeProposal)
			results = append(results, tallyResults)

		} else if !passes || netural {
			activeProposal.Ranking = sdk.ZeroInt()

		}

		activeProposal.FinalTallyResult = tallyResults
		//	keeper.RemoveFromActiveProposalQueue(ctx, activeProposal.VotingEndTime, activeProposal.ProposalID)

		logger.Info(
			fmt.Sprintf(
				"proposal %d (%s) tallied; passed: %v",
				activeProposal.ProposalID, activeProposal.GetTitle(), passes,
			),
		)

		resTags = resTags.AppendTag(tags.ProposalID, fmt.Sprintf("%d", proposalID))
		resTags = resTags.AppendTag(tags.ProposalResult, tagValue)
		keeper.SetProposal(ctx, activeProposal)
	}
	proposals = SortProposalEligibility(proposals, results)
	keeper.SetEligibilityDetails(ctx, proposals)
	return resTags
}

// ExecuteProposal Transfer funds on active proposal if voting end time reach
func ExecuteProposal(ctx sdk.Context, keeper Keeper, resTags sdk.Tags) sdk.Tags {
	logger := ctx.Logger().With("module", "x/gov")
	proposals := []Proposal{}
	results := []TallyResult{}
	var tagValue string

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

		passes, tallyResults, netural := tally(ctx, keeper, activeProposal)

		if passes {
			proposals = append(proposals, activeProposal)
			results = append(results, tallyResults)

		} else if !passes && !netural {
			keeper.DeleteProposalEligibility(ctx, activeProposal)
			keeper.DeleteDeposits(ctx, activeProposal.ProposalID)
			keeper.RemoveFromInactiveProposalQueue(ctx, activeProposal.DepositEndTime, activeProposal.ProposalID)
			keeper.RemoveFromActiveProposalQueue(ctx, activeProposal.VotingEndTime, activeProposal.ProposalID)
			activeProposal.Status = StatusRejected
			tagValue = tags.ActionProposalRejected
			activeProposal.Ranking = sdk.ZeroInt()

		} else if !passes && netural {
			activeProposal.FundingCycleCount = activeProposal.FundingCycleCount + 1
			maxCycleLimit := activeProposal.CheckMaxCycleCount()
			if maxCycleLimit {
				keeper.DeleteProposalEligibility(ctx, activeProposal)
				keeper.DeleteDeposits(ctx, activeProposal.ProposalID)
				keeper.RemoveFromInactiveProposalQueue(ctx, activeProposal.DepositEndTime, activeProposal.ProposalID)
				keeper.RemoveFromActiveProposalQueue(ctx, activeProposal.VotingEndTime, activeProposal.ProposalID)
				activeProposal.Status = StatusRejected
				tagValue = tags.ActionProposalRejected
				activeProposal.Ranking = sdk.ZeroInt()

			}

		}

		activeProposal.FinalTallyResult = tallyResults
		logger.Info(
			fmt.Sprintf(
				"proposal %d (%s) tallied; passed: %v",
				activeProposal.ProposalID, activeProposal.GetTitle(), passes,
			),
		)

		keeper.SetProposal(ctx, activeProposal)

		// TODO check if no remaining cycle left then delete proposal
		//	keeper.RemoveFromActiveProposalQueue(ctx, activeProposal.VotingEndTime, activeProposal.ProposalID)

		resTags = resTags.AppendTag(tags.ProposalID, fmt.Sprintf("%d", proposalID))
		resTags = resTags.AppendTag(tags.ProposalResult, tagValue)

	}

	proposals = SortProposalEligibility(proposals, results)
	keeper.SetEligibilityDetails(ctx, proposals)
	keeper.TransferFunds(ctx, proposals)
	return resTags
}
