package gov

import (
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

}

func TestEligibilitySorting(t *testing.T) {

	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	keeper.ck.SetSendEnabled(ctx, true)

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1)
	govHandler := NewHandler(keeper)
	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	_, _ = keeper.GetProposal(ctx, 1)

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

func TestEligibilityDelation(t *testing.T) {

	mapp, keeper, _, _, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	keeper.ck.SetSendEnabled(ctx, true)

}
