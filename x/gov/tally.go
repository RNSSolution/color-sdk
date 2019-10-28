package gov

import (
	sdk "github.com/ColorPlatform/color-sdk/types"
)


// councilMemberGovInfo used for tallying
type councilMemberGovInfo struct {
	Address		sdk.AccAddress	//councilMember Address
	Power		sdk.Dec			//CouncilMember Power
}


// CalulcateCouncilPower : calculates total power of council members
func (keeper Keeper) CalulcateCouncilPower(ctx sdk.Context) sdk.Dec {

	councilmemberIterator := keeper.stk.GetCouncilMemberIterator(ctx)
	defer councilmemberIterator.Close()
	total := sdk.NewDec(0)

	for ; councilmemberIterator.Valid(); councilmemberIterator.Next(){
		cm := &councilMemberGovInfo{}
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(councilmemberIterator.Value(),cm)
		total= total.Add(cm.Power)
	}
	return total
}

func tally(ctx sdk.Context, keeper Keeper, 
	proposal Proposal) (passes bool, tallyResults TallyResult,neutral bool) {

	results := make(map[VoteOption]sdk.Dec)
	results[OptionYes] = sdk.ZeroDec()
	results[OptionAbstain] = sdk.ZeroDec()
	results[OptionNo] = sdk.ZeroDec()
	totalVotingPower := sdk.ZeroDec()

	// iterate over all the votes
	votesIterator := keeper.GetVotes(ctx, proposal.ProposalID)
	defer votesIterator.Close()
	for ; votesIterator.Valid(); votesIterator.Next() {
		vote := &Vote{}
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(votesIterator.Value(), vote)

		if cmpower,found:=keeper.stk.GetCouncilMemberShares(ctx,vote.Voter); found{
			results[vote.Option] = results[vote.Option].Add(cmpower)
			totalVotingPower = totalVotingPower.Add(cmpower)
		}
	}
	
	tallyParams := keeper.GetTallyParams(ctx)
	tallyResults = NewTallyResultFromMap(results)
	totalCouncilPower := keeper.CalulcateCouncilPower(ctx)
	threshold := totalCouncilPower.Mul(tallyParams.Threshold)


	// If there is not enough quorum of votes, return neutral signal
	percentVoting := totalVotingPower.Quo(totalCouncilPower)
	if percentVoting.LT(tallyParams.Quorum) {
		return false, tallyResults,true
	}

	// If yes_votes minus no_votes is greater than the threshold, the proposal passes
	if (results[OptionYes].Sub(results[OptionNo])).GT(threshold) {
		return true, tallyResults, false
	}

	// If no_votes minus yes_votes is greater than the threshold, the proposal fails
	if (results[OptionNo].Sub(results[OptionYes])).GT(threshold) {
		return false, tallyResults, false
	}

	// If the voting is neutral, return neutral signal
	return false,tallyResults, true
	}