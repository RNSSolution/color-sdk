package app

import (
	cryptoAmino "github.com/ColorPlatform/prism/crypto/encoding/amino"
	cmn "github.com/ColorPlatform/prism/libs/common"
	dbm "github.com/ColorPlatform/prism/libs/db"
	"github.com/ColorPlatform/prism/libs/log"

	bapp "github.com/ColorPlatform/color-sdk/baseapp"
	"github.com/ColorPlatform/color-sdk/codec"
	sdk "github.com/ColorPlatform/color-sdk/types"
	"github.com/ColorPlatform/color-sdk/x/auth"
	"github.com/ColorPlatform/color-sdk/x/bank"
)

const (
	app3Name = "App3"
)

func NewApp3(logger log.Logger, db dbm.DB) *bapp.BaseApp {

	// Create the codec with registered Msg types
	cdc := UpdatedCodec()

	// Create the base application object.
	app := bapp.NewBaseApp(app3Name, logger, db, auth.DefaultTxDecoder(cdc))

	// Create a key for accessing the account store.
	keyAccount := sdk.NewKVStoreKey(auth.StoreKey)
	keyFees := sdk.NewKVStoreKey(auth.FeeStoreKey) // TODO

	// Set various mappers/keepers to interact easily with underlying stores
	accountKeeper := auth.NewAccountKeeper(cdc, keyAccount, auth.ProtoBaseAccount)
	bankKeeper := bank.NewBaseKeeper(accountKeeper)
	feeKeeper := auth.NewFeeCollectionKeeper(cdc, keyFees)

	app.SetAnteHandler(auth.NewAnteHandler(accountKeeper, feeKeeper))

	// Register message routes.
	// Note the handler gets access to
	app.Router().
		AddRoute("bank", bank.NewHandler(bankKeeper))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount, keyFees)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

// Update codec from app2 to register imported modules
func UpdatedCodec() *codec.Codec {
	cdc := codec.New()
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(MsgSend{}, "example/MsgSend", nil)
	cdc.RegisterConcrete(MsgIssue{}, "example/MsgIssue", nil)
	auth.RegisterCodec(cdc)
	cryptoAmino.RegisterAmino(cdc)
	return cdc
}
