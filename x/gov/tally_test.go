package gov

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/ColorPlatform/prism/abci/types"
	"github.com/ColorPlatform/prism/crypto"
	"github.com/ColorPlatform/prism/crypto/ed25519"

	sdk "github.com/ColorPlatform/color-sdk/types"
	"github.com/ColorPlatform/color-sdk/x/staking"
)

var (
	pubkeys = []crypto.PubKey{ed25519.GenPrivKey().PubKey(), ed25519.GenPrivKey().PubKey(), ed25519.GenPrivKey().PubKey()}

	testDescription   = staking.NewDescription("T", "E", "S", "T")
	testCommissionMsg = staking.NewCommissionMsg(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
)

func createValidators(t *testing.T, stakingHandler sdk.Handler, ctx sdk.Context, addrs []sdk.ValAddress, powerAmt []int64) {
	require.True(t, len(addrs) <= len(pubkeys), "Not enough pubkeys specified at top of file.")

	for i := 0; i < len(addrs); i++ {

		valTokens := sdk.TokensFromTendermintPower(powerAmt[i])
		valCreateMsg := staking.NewMsgCreateValidator(
			addrs[i], pubkeys[i], sdk.NewCoin(sdk.DefaultBondDenom, valTokens),
			testDescription, testCommissionMsg, sdk.OneInt(), sdk.OneInt(), sdk.OneInt(),
		)

		res := stakingHandler(ctx, valCreateMsg)
		require.True(t, res.IsOK())
	}
}

func TestTallyNoOneVotes(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	valAddrs := make([]sdk.ValAddress, len(addrs[:2]))
	for i, addr := range addrs[:2] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{50000, 50000})
	staking.EndBlocker(ctx, sk)

	tp := TextProposal{"Test", "test", sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1, addrs[0]}
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, tallyResults, neutral := tally(ctx, keeper, proposal)

	require.False(t, passes)
	require.True(t, neutral)
	require.True(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyAllYes(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 1
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	valAddrs := make([]sdk.ValAddress, len(addrs[:2]))
	for i, addr := range addrs[:2] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{50000, 50000})
	staking.EndBlocker(ctx, sk)

	tp := TextProposal{"Test", "test", sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1, addrs[0]}
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	// update blockchain time (add four weeks time)
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 2
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	//add votes on proposal
	err = keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, tallyResults, neutral := tally(ctx, keeper, proposal)

	require.True(t, passes)
	require.False(t, neutral)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTallyAllNo(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 1
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	valAddrs := make([]sdk.ValAddress, len(addrs[:2]))
	for i, addr := range addrs[:2] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{50000, 50000})
	staking.EndBlocker(ctx, sk)

	tp := TextProposal{"Test", "test", sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1, addrs[0]}
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	// update blockchain time (add four weeks time)
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 2
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	err = keeper.AddVote(ctx, proposalID, addrs[0], OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionNo)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, tallyResults, neutral := tally(ctx, keeper, proposal)

	require.False(t, passes)
	require.False(t, neutral)
	require.False(t, tallyResults.Equals(EmptyTallyResult()))
}

func TestTally66No(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 1
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{50000, 50000, 50000})
	staking.EndBlocker(ctx, sk)

	tp := TextProposal{"Test", "test", sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1, addrs[0]}
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	// update blockchain time (add four weeks time)
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 2
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	err = keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionNo)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, _, neutral := tally(ctx, keeper, proposal)

	require.False(t, passes)
	require.False(t, neutral)
}

func TestTally66Yes(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 1
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{50000, 50000, 50000})
	staking.EndBlocker(ctx, sk)

	tp := TextProposal{"Test", "test", sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1, addrs[0]}
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	// update blockchain time (add four weeks time)
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 2
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	err = keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionNo)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, _, neutral := tally(ctx, keeper, proposal)

	require.True(t, passes)
	require.False(t, neutral)
}

func TestTallyNeutral(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 1
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{50000, 50000, 50000})
	staking.EndBlocker(ctx, sk)

	tp := TextProposal{"Test", "test", sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1, addrs[0]}
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	// update blockchain time (add four weeks time)
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 2
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	err = keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionNo)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionAbstain)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, _, neutral := tally(ctx, keeper, proposal)

	require.False(t, passes)
	require.True(t, neutral)
}

func TestTallyAllAbstain(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 1
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{50000, 50000, 50000})
	staking.EndBlocker(ctx, sk)

	tp := TextProposal{"Test", "test", sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1, addrs[0]}
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	// update blockchain time (add four weeks time)
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 2
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	err = keeper.AddVote(ctx, proposalID, addrs[0], OptionAbstain)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionAbstain)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionAbstain)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, _, neutral := tally(ctx, keeper, proposal)

	require.False(t, passes)
	require.True(t, neutral)
}

func TestTallyQuorum(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 1
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{1000000, 50000, 50000})
	staking.EndBlocker(ctx, sk)

	tp := TextProposal{"Test", "test", sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1, addrs[0]}
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	// update blockchain time (add four weeks time)
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 2
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	//only one council member votes
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionNo)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, _, neutral := tally(ctx, keeper, proposal)

	require.False(t, passes)
	require.True(t, neutral)
}

func TestTallyExactQuorumPass(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 1
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	valAddrs := make([]sdk.ValAddress, len(addrs[:2]))
	for i, addr := range addrs[:2] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{850000, 150000})
	staking.EndBlocker(ctx, sk)

	tp := TextProposal{"Test", "test", sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1, addrs[0]}
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	// update blockchain time (add four weeks time)
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 2
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	//only one council member votes
	// err = keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	// require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, _, neutral := tally(ctx, keeper, proposal)

	require.True(t, passes)
	require.False(t, neutral)
}

func TestTallyExactQuorumFail(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 1
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	valAddrs := make([]sdk.ValAddress, len(addrs[:2]))
	for i, addr := range addrs[:2] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{850001, 150000})
	staking.EndBlocker(ctx, sk)

	tp := TextProposal{"Test", "test", sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1, addrs[0]}
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	// update blockchain time (add four weeks time)
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 2
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	//only one council member votes
	// err = keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	// require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, _, neutral := tally(ctx, keeper, proposal)

	require.False(t, passes)
	require.True(t, neutral)
}

// Test exact threshold value for yes and no votes
func TestTallyExactThreshold(t *testing.T) {
	mapp, keeper, sk, addrs, _, _ := getMockApp(t, 10, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	stakingHandler := staking.NewHandler(sk)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	newHeader.Height = 1
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	valAddrs := make([]sdk.ValAddress, len(addrs[:3]))
	for i, addr := range addrs[:3] {
		valAddrs[i] = sdk.ValAddress(addr)
	}

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{850000, 50001, 100000})
	staking.EndBlocker(ctx, sk)

	tp := TextProposal{"Test", "test", sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 1, addrs[0]}
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	// update blockchain time (add four weeks time)
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(FourWeeksHours)
	newHeader.Height = 2
	ctx = ctx.WithBlockHeader(newHeader)
	EndBlocker(ctx, keeper)

	//only one council member votes
	err = keeper.AddVote(ctx, proposalID, addrs[0], OptionAbstain)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.Nil(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[2], OptionAbstain)
	require.Nil(t, err)

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, _, neutral := tally(ctx, keeper, proposal)

	require.True(t, passes)
	require.False(t, neutral)

	// council member changes vote
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionNo)
	require.Nil(t, err)

	proposal, ok = keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, _, neutral = tally(ctx, keeper, proposal)

	require.False(t, passes)
	require.False(t, neutral)
}
