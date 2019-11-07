package gov

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/ColorPlatform/prism/abci/types"

	sdk "github.com/ColorPlatform/color-sdk/types"
	"github.com/ColorPlatform/color-sdk/x/staking"
)

func TestTickExpiredDepositPeriod(t *testing.T) {

	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	keeper.ck.SetSendEnabled(ctx, true)
	govHandler := NewHandler(keeper)

	inactiveQueue := keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000000)}, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000000)}, 1)

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetDepositParams(ctx).MaxDepositPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	EndBlocker(ctx, keeper)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
}

func TestTickMultipleExpiredDepositPeriod(t *testing.T) {

	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	keeper.ck.SetSendEnabled(ctx, true)
	govHandler := NewHandler(keeper)

	inactiveQueue := keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1)

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(2) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg2 := NewMsgSubmitProposal("Test2", "test2", ProposalTypeText, addrs[1], sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1)
	res = govHandler(ctx, newProposalMsg2)
	require.True(t, res.IsOK())

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetDepositParams(ctx).MaxDepositPeriod).Add(time.Duration(-1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	EndBlocker(ctx, keeper)
	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(5) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	EndBlocker(ctx, keeper)
	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
}

func TestTickPassedDepositPeriod(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	keeper.ck.SetSendEnabled(ctx, true)
	govHandler := NewHandler(keeper)

	inactiveQueue := keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	activeQueue := keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1)

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	var proposalID uint64
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(res.Data, &proposalID)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newDepositMsg := NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)})
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	activeQueue = keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()
}

func TestTickPassedVotingPeriod(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	SortAddresses(addrs)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	keeper.ck.SetSendEnabled(ctx, true)
	govHandler := NewHandler(keeper)

	inactiveQueue := keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	activeQueue := keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromTendermintPower(5))}
	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], proposalCoins, proposalCoins, 1)

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	var proposalID uint64
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(res.Data, &proposalID)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	newDepositMsg := NewMsgDeposit(addrs[1], proposalID, proposalCoins)
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetDepositParams(ctx).MaxDepositPeriod).Add(keeper.GetVotingParams(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	activeQueue = keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, activeQueue.Valid())
	var activeProposalID uint64
	keeper.cdc.UnmarshalBinaryLengthPrefixed(activeQueue.Value(), &activeProposalID)
	proposal, ok := keeper.GetProposal(ctx, activeProposalID)
	require.True(t, ok)
	require.Equal(t, StatusVotingPeriod, proposal.Status)
	depositsIterator := keeper.GetDeposits(ctx, proposalID)
	require.True(t, depositsIterator.Valid())
	depositsIterator.Close()
	activeQueue.Close()

	EndBlocker(ctx, keeper)

	activeQueue = keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()
}

func TestTickPassedSubmitProposal(t *testing.T) {

	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
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

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1)
	govHandler := NewHandler(keeper)
	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	proposal, _ := keeper.GetProposal(ctx, 1)
	require.Nil(t, proposal)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 1
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	require.Empty(t, keeper.GetProposalEligibility(ctx))
	newProposalMsg = NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1)
	govHandler = NewHandler(keeper)
	res = govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	require.Equal(t, 1, len(keeper.GetProposalEligibility(ctx)))

}

func TestTickPassedLimitFirstFundingCycle(t *testing.T) {
	t.Log("Starting TestTickPassedLimitFirstFundingCycle")

	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)
	SortAddresses(addrs)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	keeper.ck.SetSendEnabled(ctx, true)

	keeper.ck.SetSendEnabled(ctx, true)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 1
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)
	_, err := keeper.GetCurrentCycle(ctx)
	require.NotNil(t, err)
	days, err := keeper.GetDaysPassed(ctx)
	require.Equal(t, 0, days)

	header = abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 2
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)
	days, err = keeper.GetDaysPassed(ctx)
	require.Equal(t, 28, days)

}

func TestTickSortingProposalEligibility(t *testing.T) {

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

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{50000, 7, 8})
	staking.EndBlocker(ctx, sk)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 3
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	//

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1)
	govHandler := NewHandler(keeper)
	res := govHandler(ctx, newProposalMsg)

	require.True(t, res.IsOK())

	msg := NewMsgVote(addrs[0], 1, OptionYes)
	res = govHandler(ctx, msg)

	require.True(t, res.IsOK())

	newProposalMsg2 := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1)
	res = govHandler(ctx, newProposalMsg2)
	require.True(t, res.IsOK())

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 4
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	eligibility_list := keeper.GetProposalEligibility(ctx)
	require.Equal(t, uint64(1), eligibility_list[0].ProposalID)

	//TODO working have to be done

}

func TestTickTransferFunds(t *testing.T) {

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

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{50000, 7, 8})
	staking.EndBlocker(ctx, sk)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 3
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	//

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1)
	govHandler := NewHandler(keeper)
	res := govHandler(ctx, newProposalMsg)

	require.True(t, res.IsOK())

	msg := NewMsgVote(addrs[0], 1, OptionYes)
	res = govHandler(ctx, msg)
	require.True(t, res.IsOK())

	newProposalMsg2 := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1)
	res = govHandler(ctx, newProposalMsg2)
	require.True(t, res.IsOK())

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 4
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	// Add time of end cycle

	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 4
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	//TODO working have to be done

}

func TestVotingofSubmittedProposalBasic(t *testing.T) {

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

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 2
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1)
	govHandler := NewHandler(keeper)
	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	_, _ = keeper.GetProposal(ctx, 1)

	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{50000, 7, 8})
	staking.EndBlocker(ctx, sk)

	msg := NewMsgVote(addrs[0], 1, OptionYes)
	res = govHandler(ctx, msg)
	vote, found := keeper.GetVote(ctx, msg.ProposalID, msg.Voter)
	require.True(t, found)
	require.Equal(t, addrs[0], vote.Voter)
	require.Equal(t, msg.ProposalID, vote.ProposalID)
	require.Equal(t, msg.Option, vote.Option)
	//require.True(t, res.IsOK())

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 3
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	require.Equal(t, 1, len(keeper.GetProposalEligibility(ctx)))

}

func TestVotingofSubmittedProposalAdvance(t *testing.T) {

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

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 2
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1)
	govHandler := NewHandler(keeper)
	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	_, _ = keeper.GetProposal(ctx, 1)

	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{50000, 7, 8})
	staking.EndBlocker(ctx, sk)

	msg := NewMsgVote(addrs[0], 1, OptionYes)
	res = govHandler(ctx, msg)
	vote, found := keeper.GetVote(ctx, msg.ProposalID, msg.Voter)
	require.True(t, found)
	require.Equal(t, addrs[0], vote.Voter)
	require.Equal(t, msg.ProposalID, vote.ProposalID)
	require.Equal(t, msg.Option, vote.Option)
	//require.True(t, res.IsOK())

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 3
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	require.Equal(t, 1, len(keeper.GetProposalEligibility(ctx)))

}
