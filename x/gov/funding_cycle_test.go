package gov

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/ColorPlatform/color-sdk/types"
	abci "github.com/ColorPlatform/prism/abci/types"
)

func TestEligibility(t *testing.T) {

	mapp, keeper, _, _, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	keeper.ck.SetSendEnabled(ctx, true)

	require.Empty(t, keeper.GetProposalEligibility(ctx))
	keeper.AddProposalEligibility(ctx, 5)

	fmt.Println(keeper.GetProposalEligibility(ctx))

}

func TestEligibilitySorting(t *testing.T) {

	mapp, keeper, _, _, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	keeper.ck.SetSendEnabled(ctx, true)
	eligibilityQueue := []EligibilityDetails{}
	eligibility := NewEligibilityDetails(1, sdk.NewInt(1), sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)})
	eligibilityQueue = Append(eligibilityQueue, eligibility)

	eligibility2 := NewEligibilityDetails(2, sdk.NewInt(5), sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)})
	eligibilityQueue = Append(eligibilityQueue, eligibility2)

	eligibility3 := NewEligibilityDetails(3, sdk.NewInt(2), sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)})
	eligibilityQueue = Append(eligibilityQueue, eligibility3)

	eligibilityQueue = SortProposalEligibility(eligibilityQueue)
	require.Equal(t, uint64(2), eligibilityQueue[0].ProposalID)

}

func TestVerifyFunction(t *testing.T) {
	valTokens := sdk.TokensFromTendermintPower(6)
	coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, valTokens)}

	totalFundCount := sdk.NewCoins()
	//weeklyIncome := sdk.NewDec(10)
	limit := sdk.NewInt(5)

	result := VerifyAmount(totalFundCount, limit)
	require.True(t, result)
	totalFundCount = totalFundCount.Add(coins)
	result = VerifyAmount(totalFundCount, limit)
	require.False(t, result)

}
