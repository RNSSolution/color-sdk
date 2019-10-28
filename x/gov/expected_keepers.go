package gov

import sdk "github.com/ColorPlatform/color-sdk/types"

// bank keeper expected
type BankKeeper interface {
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	// TODO remove once governance doesn't require use of accounts
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
	SetSendEnabled(ctx sdk.Context, enabled bool)
}

// StakingKeeper expected
type StakingKeeper interface {
	GetCouncilMemberIterator(ctx sdk.Context) sdk.Iterator
	GetCouncilMemberShares(ctx sdk.Context, memAddr sdk.AccAddress) (sdk.Dec,bool)
}
