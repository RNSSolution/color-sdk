package gov

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"testing"

	"github.com/ColorPlatform/color-sdk/codec"
	"github.com/ColorPlatform/color-sdk/store"
	sdk "github.com/ColorPlatform/color-sdk/types"
	"github.com/ColorPlatform/color-sdk/x/auth"
	"github.com/ColorPlatform/color-sdk/x/bank"
	distr "github.com/ColorPlatform/color-sdk/x/distribution"
	"github.com/ColorPlatform/color-sdk/x/distribution/types"
	"github.com/ColorPlatform/color-sdk/x/mint"
	"github.com/ColorPlatform/color-sdk/x/mock"
	"github.com/ColorPlatform/color-sdk/x/params"
	"github.com/ColorPlatform/color-sdk/x/staking"
	abci "github.com/ColorPlatform/prism/abci/types"
	"github.com/ColorPlatform/prism/crypto"
	"github.com/ColorPlatform/prism/crypto/ed25519"
	dbm "github.com/ColorPlatform/prism/libs/db"
	logm "github.com/ColorPlatform/prism/libs/log"
	"github.com/stretchr/testify/require"
)

// initialize the mock application for this module
func getMockApp(t *testing.T, numGenAccs int, genState GenesisState, genAccs []auth.Account) (
	mapp *mock.App, keeper Keeper, sk staking.Keeper, addrs []sdk.AccAddress,
	pubKeys []crypto.PubKey, privKeys []crypto.PrivKey) {

	mapp = mock.NewApp()

	staking.RegisterCodec(mapp.Cdc)
	RegisterCodec(mapp.Cdc)

	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyGov := sdk.NewKVStoreKey(StoreKey)
	keyDistr := sdk.NewKVStoreKey(distr.StoreKey)
	keyMinting := sdk.NewKVStoreKey(mint.StoreKey)

	pk := mapp.ParamsKeeper
	ck := bank.NewBaseKeeper(mapp.AccountKeeper, mapp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	sk = staking.NewKeeper(mapp.Cdc, keyStaking, tkeyStaking, ck, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	feeKeeper := auth.NewFeeCollectionKeeper(mapp.Cdc, mapp.KeyFeeCollection)
	distrKeeper := distr.NewKeeper(mapp.Cdc, keyDistr, pk.Subspace(distr.DefaultParamspace), ck, &sk, feeKeeper, distr.DefaultCodespace)

	minKeeper := mint.NewKeeper(mapp.Cdc, keyMinting, pk.Subspace(mint.DefaultParamspace), &sk, feeKeeper)
	keeper = NewKeeper(mapp.Cdc, distrKeeper, minKeeper, keyGov, pk, pk.Subspace("testgov"), ck, sk, sk, DefaultCodespace)

	mapp.Router().AddRoute(RouterKey, NewHandler(keeper))
	mapp.QueryRouter().AddRoute(QuerierRoute, NewQuerier(keeper))

	mapp.SetEndBlocker(getEndBlocker(keeper))
	mapp.SetInitChainer(getInitChainer(mapp, keeper, sk, genState))

	require.NoError(t, mapp.CompleteSetup(keyStaking, tkeyStaking, keyGov))

	valTokens := sdk.TokensFromTendermintPower(10000000000000)
	if genAccs == nil || len(genAccs) == 0 {
		genAccs, addrs, pubKeys, privKeys = mock.CreateGenAccounts(numGenAccs,
			sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, valTokens)})
	}

	mock.SetGenesis(mapp, genAccs)

	return mapp, keeper, sk, addrs, pubKeys, privKeys
}

// gov and staking endblocker
func getEndBlocker(keeper Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		tags := EndBlocker(ctx, keeper)
		return abci.ResponseEndBlock{
			Tags: tags,
		}
	}
}

// gov and staking initchainer
func getInitChainer(mapp *mock.App, keeper Keeper, stakingKeeper staking.Keeper, genState GenesisState) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)

		stakingGenesis := staking.DefaultGenesisState()
		tokens := sdk.TokensFromTendermintPower(2000000)
		stakingGenesis.Pool.NotBondedTokens = tokens

		validators, err := staking.InitGenesis(ctx, stakingKeeper, stakingGenesis)
		if err != nil {
			panic(err)
		}
		if genState.IsEmpty() {
			InitGenesis(ctx, keeper, DefaultGenesisState())
		} else {
			InitGenesis(ctx, keeper, genState)
		}
		return abci.ResponseInitChain{
			Validators: validators,
		}
	}
}

// TODO: Remove once address interface has been implemented (ref: #2186)
func SortValAddresses(addrs []sdk.ValAddress) {
	var byteAddrs [][]byte
	for _, addr := range addrs {
		byteAddrs = append(byteAddrs, addr.Bytes())
	}

	SortByteArrays(byteAddrs)

	for i, byteAddr := range byteAddrs {
		addrs[i] = byteAddr
	}
}

// Sorts Addresses
func SortAddresses(addrs []sdk.AccAddress) {
	var byteAddrs [][]byte
	for _, addr := range addrs {
		byteAddrs = append(byteAddrs, addr.Bytes())
	}
	SortByteArrays(byteAddrs)
	for i, byteAddr := range byteAddrs {
		addrs[i] = byteAddr
	}
}

// implement `Interface` in sort package.
type sortByteArrays [][]byte

func (b sortByteArrays) Len() int {
	return len(b)
}

func (b sortByteArrays) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(b[i], b[j]) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		log.Panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
		return false
	}
}

func (b sortByteArrays) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// Public
func SortByteArrays(src [][]byte) [][]byte {
	sorted := sortByteArrays(src)
	sort.Sort(sorted)
	return sorted
}

func testProposal() TextProposal {
	var priv = ed25519.GenPrivKey()
	var addr = sdk.AccAddress(priv.PubKey().Address())
	return NewTextProposal("Test", "description", sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}, 4, addr)
}

// checks if two proposals are equal (note: slow, for tests only)
func ProposalEqual(proposalA Proposal, proposalB Proposal) bool {
	return bytes.Equal(msgCdc.MustMarshalBinaryBare(proposalA), msgCdc.MustMarshalBinaryBare(proposalB))
}

// create a codec used only for testing
func MakeTestCodec() *codec.Codec {
	var cdc = codec.New()
	bank.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	types.RegisterCodec(cdc) // distr
	return cdc
}

type DummyFeeCollectionKeeper struct{}

var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delPk2   = ed25519.GenPrivKey().PubKey()
	delPk3   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())
	delAddr2 = sdk.AccAddress(delPk2.Address())
	delAddr3 = sdk.AccAddress(delPk3.Address())

	valOpPk1    = ed25519.GenPrivKey().PubKey()
	valOpPk2    = ed25519.GenPrivKey().PubKey()
	valOpPk3    = ed25519.GenPrivKey().PubKey()
	valOpAddr1  = sdk.ValAddress(valOpPk1.Address())
	valOpAddr2  = sdk.ValAddress(valOpPk2.Address())
	valOpAddr3  = sdk.ValAddress(valOpPk3.Address())
	valAccAddr1 = sdk.AccAddress(valOpPk1.Address()) // generate acc addresses for these validator keys too
	valAccAddr2 = sdk.AccAddress(valOpPk2.Address())
	valAccAddr3 = sdk.AccAddress(valOpPk3.Address())

	valConsPk1   = ed25519.GenPrivKey().PubKey()
	valConsPk2   = ed25519.GenPrivKey().PubKey()
	valConsPk3   = ed25519.GenPrivKey().PubKey()
	valConsAddr1 = sdk.ConsAddress(valConsPk1.Address())
	valConsAddr2 = sdk.ConsAddress(valConsPk2.Address())
	valConsAddr3 = sdk.ConsAddress(valConsPk3.Address())

	// test addresses
	TestAddrs = []sdk.AccAddress{
		delAddr1, delAddr2, delAddr3,
		valAccAddr1, valAccAddr2, valAccAddr3,
	}

	emptyDelAddr sdk.AccAddress
	emptyValAddr sdk.ValAddress
	emptyPubkey  crypto.PubKey
)

// testMsg is a mock transaction that has a validation which can fail.
type testMsg struct {
	signers     []sdk.AccAddress
	positiveNum int64
}

// hogpodge of all sorts of input required for testing
func CreateTestInputAdvanced1(t *testing.T, isCheckTx bool, initPower int64,
	communityTax sdk.Dec, numGenAccs int, genState GenesisState, genAccs []auth.Account) (mapp *mock.App, ctx sdk.Context, keeper Keeper, sk staking.Keeper, addrs []sdk.AccAddress,
	pubKeys []crypto.PubKey, privKeys []crypto.PrivKey) {
	mapp = mock.NewApp()

	bank.RegisterCodec(mapp.Cdc)
	staking.RegisterCodec(mapp.Cdc)
	types.RegisterCodec(mapp.Cdc)
	RegisterCodec(mapp.Cdc)

	initCoins := sdk.TokensFromTendermintPower(initPower)
	keyDistr := sdk.NewKVStoreKey(distr.StoreKey)
	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyFeeCollection := sdk.NewKVStoreKey(auth.FeeStoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keyGov := sdk.NewKVStoreKey(StoreKey)
	keyMinting := sdk.NewKVStoreKey(mint.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	ms.MountStoreWithDB(keyDistr, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyStaking, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyFeeCollection, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	pk := params.NewKeeper(mapp.Cdc, keyParams, tkeyParams)
	ctx = sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, isCheckTx, logm.NewNopLogger())
	accountKeeper := auth.NewAccountKeeper(mapp.Cdc, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bankKeeper := bank.NewBaseKeeper(accountKeeper, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	sk = staking.NewKeeper(mapp.Cdc, keyStaking, tkeyStaking, bankKeeper, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	feeKeeper := auth.NewFeeCollectionKeeper(mapp.Cdc, keyFeeCollection)
	distrKeeper := distr.NewKeeper(mapp.Cdc, keyDistr, pk.Subspace(DefaultParamspace), bankKeeper, &sk, feeKeeper, distr.DefaultCodespace)
	minKeeper := mint.NewKeeper(mapp.Cdc, keyMinting, pk.Subspace(mint.DefaultParamspace), &sk, feeKeeper)

	keeper = NewKeeper(mapp.Cdc, distrKeeper, minKeeper, keyGov, pk, pk.Subspace("testgov"), bankKeeper, sk, sk, DefaultCodespace)

	sk.SetPool(ctx, staking.InitialPool())
	sk.SetParams(ctx, staking.DefaultParams())

	sk.SetHooks(distrKeeper.Hooks())

	// set genesis items required for distribution
	distrKeeper.SetFeePool(ctx, types.InitialFeePool())
	distrKeeper.SetCommunityTax(ctx, communityTax)
	distrKeeper.SetBaseProposerReward(ctx, sdk.NewDecWithPrec(1, 2))
	distrKeeper.SetBonusProposerReward(ctx, sdk.NewDecWithPrec(4, 2))
	fmt.Println("community tax", distrKeeper.GetCommunityTax(ctx))

	mapp.Router().AddRoute(RouterKey, NewHandler(keeper))
	mapp.QueryRouter().AddRoute(QuerierRoute, NewQuerier(keeper))

	mapp.SetEndBlocker(getEndBlocker(keeper))
	mapp.SetInitChainer(getInitChainer(mapp, keeper, sk, genState))

	require.NoError(t, mapp.CompleteSetup(keyStaking, tkeyStaking, keyGov))

	// fill all the addresses with some coins, set the loose pool tokens simultaneously

	// set the distribution hooks on staking

	valTokens := sdk.TokensFromTendermintPower(10000000000000)
	if genAccs == nil || len(genAccs) == 0 {
		genAccs, addrs, pubKeys, privKeys = mock.CreateGenAccounts(numGenAccs,
			sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, valTokens)})
	}
	mock.SetGenesis(mapp, genAccs)

	for _, addr := range TestAddrs {
		pool := sk.GetPool(ctx)
		_, _, err := bankKeeper.AddCoins(ctx, addr, sdk.Coins{
			sdk.NewCoin(sk.GetParams(ctx).BondDenom, initCoins),
		})
		require.Nil(t, err)
		pool.NotBondedTokens = pool.NotBondedTokens.Add(initCoins)
		sk.SetPool(ctx, pool)
	}

	return mapp, ctx, keeper, sk, addrs, pubKeys, privKeys
}

// hogpodge of all sorts of input required for testing
func CreateTestInputAdvanced(t *testing.T, isCheckTx bool, initPower int64,
	communityTax sdk.Dec, numGenAccs int, genState GenesisState, genAccs []auth.Account) sdk.Context {

	initCoins := sdk.TokensFromTendermintPower(initPower)

	keyDistr := sdk.NewKVStoreKey(types.StoreKey)
	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyFeeCollection := sdk.NewKVStoreKey(auth.FeeStoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	ms.MountStoreWithDB(keyDistr, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyStaking, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyFeeCollection, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	cdc := MakeTestCodec()
	pk := params.NewKeeper(cdc, keyParams, tkeyParams)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, isCheckTx, logm.NewNopLogger())
	accountKeeper := auth.NewAccountKeeper(cdc, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bankKeeper := bank.NewBaseKeeper(accountKeeper, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	sk := staking.NewKeeper(cdc, keyStaking, tkeyStaking, bankKeeper, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	sk.SetPool(ctx, staking.InitialPool())
	sk.SetParams(ctx, staking.DefaultParams())

	// fill all the addresses with some coins, set the loose pool tokens simultaneously
	for _, addr := range TestAddrs {
		pool := sk.GetPool(ctx)
		_, _, err := bankKeeper.AddCoins(ctx, addr, sdk.Coins{
			sdk.NewCoin(sk.GetParams(ctx).BondDenom, initCoins),
		})
		require.Nil(t, err)
		pool.NotBondedTokens = pool.NotBondedTokens.Add(initCoins)
		sk.SetPool(ctx, pool)
	}

	feeKeeper := auth.NewFeeCollectionKeeper(cdc, keyFeeCollection)
	keeper := distr.NewKeeper(cdc, keyDistr, pk.Subspace(DefaultParamspace), bankKeeper, sk, feeKeeper, types.DefaultCodespace)

	// set the distribution hooks on staking
	sk.SetHooks(keeper.Hooks())

	// set genesis items required for distribution
	keeper.SetFeePool(ctx, types.InitialFeePool())
	keeper.SetCommunityTax(ctx, communityTax)
	keeper.SetBaseProposerReward(ctx, sdk.NewDecWithPrec(1, 2))
	keeper.SetBonusProposerReward(ctx, sdk.NewDecWithPrec(4, 2))
	fmt.Println(keeper.GetCommunityTax(ctx))

	return ctx
}

func getMockApp1(t *testing.T, isCheckTx bool, initPower int64,
	communityTax sdk.Dec, numGenAccs int, genState GenesisState, genAccs []auth.Account) (
	mapp *mock.App, keeper Keeper, sk staking.Keeper, addrs []sdk.AccAddress,
	pubKeys []crypto.PubKey, privKeys []crypto.PrivKey) {

	mapp = mock.NewApp()

	staking.RegisterCodec(mapp.Cdc)
	bank.RegisterCodec(mapp.Cdc)
	types.RegisterCodec(mapp.Cdc)
	RegisterCodec(mapp.Cdc)

	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyGov := sdk.NewKVStoreKey(StoreKey)
	keyDistr := sdk.NewKVStoreKey(distr.StoreKey)
	keyMinting := sdk.NewKVStoreKey(mint.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyDistr, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyStaking, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)

	pk := mapp.ParamsKeeper
	accountKeeper := auth.NewAccountKeeper(mapp.Cdc, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	ck := bank.NewBaseKeeper(accountKeeper, mapp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	sk = staking.NewKeeper(mapp.Cdc, keyStaking, tkeyStaking, ck, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	feeKeeper := auth.NewFeeCollectionKeeper(mapp.Cdc, mapp.KeyFeeCollection)
	distrKeeper := distr.NewKeeper(mapp.Cdc, keyDistr, pk.Subspace(distr.DefaultParamspace), ck, &sk, feeKeeper, distr.DefaultCodespace)

	minKeeper := mint.NewKeeper(mapp.Cdc, keyMinting, pk.Subspace(mint.DefaultParamspace), &sk, feeKeeper)
	keeper = NewKeeper(mapp.Cdc, distrKeeper, minKeeper, keyGov, pk, pk.Subspace("testgov"), ck, sk, sk, DefaultCodespace)

	pk = params.NewKeeper(mapp.Cdc, keyParams, tkeyParams)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, isCheckTx, logm.NewNopLogger())

	sk.SetPool(ctx, staking.InitialPool())
	sk.SetParams(ctx, staking.DefaultParams())
	sk.SetHooks(distrKeeper.Hooks())
	distrKeeper.SetFeePool(ctx, types.InitialFeePool())
	distrKeeper.SetCommunityTax(ctx, communityTax)
	distrKeeper.SetBaseProposerReward(ctx, sdk.NewDecWithPrec(1, 2))
	distrKeeper.SetBonusProposerReward(ctx, sdk.NewDecWithPrec(4, 2))
	fmt.Println("community tax", distrKeeper.GetCommunityTax(ctx))

	mapp.Router().AddRoute(RouterKey, NewHandler(keeper))
	mapp.QueryRouter().AddRoute(QuerierRoute, NewQuerier(keeper))

	mapp.SetEndBlocker(getEndBlocker(keeper))
	mapp.SetInitChainer(getInitChainer(mapp, keeper, sk, genState))

	require.NoError(t, mapp.CompleteSetup(keyStaking, tkeyStaking, keyGov))

	valTokens := sdk.TokensFromTendermintPower(10000000000000)
	if genAccs == nil || len(genAccs) == 0 {
		genAccs, addrs, pubKeys, privKeys = mock.CreateGenAccounts(numGenAccs,
			sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, valTokens)})
	}

	mock.SetGenesis(mapp, genAccs)

	return mapp, keeper, sk, addrs, pubKeys, privKeys
}
