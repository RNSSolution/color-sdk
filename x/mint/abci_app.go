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

	updateWeeklySupply(params,&minter)
	k.SetMinter(ctx, minter)
	
	// mint coins, add to collected fees, update supply
	mintedCoin := minter.BlockProvision(params)
	k.fck.AddCollectedFees(ctx, sdk.Coins{mintedCoin})
	k.sk.InflateSupply(ctx, mintedCoin.Amount)

	minter.BlockTime= time.Now().UTC()
	k.SetMinter(ctx,minter)
}

// function to check  block height and time and update timestamps if needed.
func updateWeeklySupply(params Params,minter *Minter) {
	  if time.Now().UTC().After(minter.DeflationTime) {
		minter.DeflationTime = minter.DeflationTime.Add(time.Second * 10)
		minter.WeeklyProvisions, minter.MintingSpeed = minter.NewWeeklySupply(params)
	 }
}
// Add(time.Second * 10)
// minter.DeflationTime = minter.DeflationTime.AddDate(0, 0, 7 * 52)
