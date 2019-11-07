package gov

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/ColorPlatform/color-sdk/types"
	"github.com/ColorPlatform/color-sdk/x/staking"
	abci "github.com/ColorPlatform/prism/abci/types"
	"github.com/stretchr/testify/require"
)

func TestProposalKind_Format(t *testing.T) {
	typeText, _ := ProposalTypeFromString("Text")
	tests := []struct {
		pt                   ProposalKind
		sprintFArgs          string
		expectedStringOutput string
	}{
		{typeText, "%s", "Text"},
		{typeText, "%v", "1"},
	}
	for _, tt := range tests {
		got := fmt.Sprintf(tt.sprintFArgs, tt.pt)
		require.Equal(t, tt.expectedStringOutput, got)
	}
}

func TestProposalStatus_Format(t *testing.T) {
	statusDepositPeriod, _ := ProposalStatusFromString("DepositPeriod")
	tests := []struct {
		pt                   ProposalStatus
		sprintFArgs          string
		expectedStringOutput string
	}{
		{statusDepositPeriod, "%s", "DepositPeriod"},
		{statusDepositPeriod, "%v", "1"},
	}
	for _, tt := range tests {
		got := fmt.Sprintf(tt.sprintFArgs, tt.pt)
		require.Equal(t, tt.expectedStringOutput, got)
	}
}

func TestProposalSubmission(t *testing.T) {

	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	SortAddresses(addrs)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	keeper.ck.SetSendEnabled(ctx, true)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 1
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	// Crating validators

	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{5, 7, 8})
	staking.EndBlocker(ctx, sk)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 3
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	// Should submit prosal when staking is less then limit

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1)
	govHandler := NewHandler(keeper)
	res := govHandler(ctx, newProposalMsg)

	fmt.Println(fmt.Println(res))
	require.True(t, res.IsOK())

	//_, err := keeper.GetProposal(ctx, 1)
	//require.False(t, err)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 4
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

}
