package mint

import (
	"os"
	"testing"
	"time"

	abci "github.com/ColorPlatform/prism/abci/types"
	"github.com/ColorPlatform/prism/libs/log"

	"github.com/stretchr/testify/require"

	"github.com/ColorPlatform/color-sdk/codec"
	"github.com/ColorPlatform/color-sdk/store"
	sdk "github.com/ColorPlatform/color-sdk/types"
	"github.com/ColorPlatform/color-sdk/x/auth"
	"github.com/ColorPlatform/color-sdk/x/bank"
	"github.com/ColorPlatform/color-sdk/x/params"
	"github.com/ColorPlatform/color-sdk/x/staking"
	dbm "github.com/ColorPlatform/prism/libs/db"
)

type testInput struct {
	ctx        sdk.Context
	cdc        *codec.Codec
	mintKeeper Keeper
}

func createTestCodec() *codec.Codec {
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	return cdc
}

func newTestInput(t *testing.T) testInput {
	cdc := createTestCodec()
	db := dbm.NewMemDB()

	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keyFeeCollection := sdk.NewKVStoreKey(auth.FeeStoreKey)
	keyMint := sdk.NewKVStoreKey(StoreKey)

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyStaking, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyFeeCollection, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyMint, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	paramsKeeper := params.NewKeeper(cdc, keyParams, tkeyParams)
	feeCollectionKeeper := auth.NewFeeCollectionKeeper(cdc, keyFeeCollection)
	accountKeeper := auth.NewAccountKeeper(cdc, keyAcc, paramsKeeper.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bankKeeper := bank.NewBaseKeeper(accountKeeper, paramsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	stakingKeeper := staking.NewKeeper(
		cdc, keyStaking, tkeyStaking, bankKeeper, paramsKeeper.Subspace(staking.DefaultParamspace), staking.DefaultCodespace,
	)
	mintKeeper := NewKeeper(
		cdc, keyMint, paramsKeeper.Subspace(DefaultParamspace), &stakingKeeper, feeCollectionKeeper,
	)

	ctx := sdk.NewContext(ms, abci.Header{Time: time.Unix(0, 0)}, false, log.NewTMLogger(os.Stdout))

	mintKeeper.SetParams(ctx, DefaultParams())
	mintKeeper.SetMinter(ctx, DefaultInitialMinter())

	return testInput{ctx, cdc, mintKeeper}
}
