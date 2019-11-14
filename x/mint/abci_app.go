package mint

import (
	"time"

	sdk "github.com/ColorPlatform/color-sdk/types"
)

// Inflate every block, update inflation parameters once per hour
func BeginBlocker(ctx sdk.Context, k Keeper) {

	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	updateWeeklySupply(params,&minter,ctx.BlockHeader().Time)
	k.SetMinter(ctx, minter)

	// mint coins, add to collected fees, update supply
	mintedCoin := minter.BlockProvision(params,ctx.BlockHeader().Time)
	k.fck.AddCollectedFees(ctx, sdk.Coins{mintedCoin})
	k.sk.InflateSupply(ctx, mintedCoin.Amount)

	minter.BlockTime= ctx.BlockHeader().Time
	k.SetMinter(ctx,minter)
}

// function to check  block height and time and update timestamps if needed.
func updateWeeklySupply(params Params,minter *Minter, currentTime time.Time) {
	  if currentTime.After(minter.DeflationTime) {
		minter.DeflationTime = minter.DeflationTime.Add(time.Second* deflationtime)
		minter.WeeklyProvisions, minter.MintingSpeed = minter.NewWeeklySupply(params)
	 }
}
