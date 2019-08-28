package mint

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var t1 time.Time
var t2 time.Time

// Inflate every block, update inflation parameters once per hour
func BeginBlocker(ctx sdk.Context, k Keeper) {

	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	minter.WeeklyProvisions = updateWeeklySupply(ctx.BlockHeight(), minter)

	k.SetMinter(ctx, minter)

	// mint coins, add to collected fees, update supply
	mintedCoin := minter.BlockProvision(params)
	k.fck.AddCollectedFees(ctx, sdk.Coins{mintedCoin})
	k.sk.InflateSupply(ctx, mintedCoin.Amount)
}

// function to check  block height and time and update timestamps if needed.
func updateWeeklySupply(height int64, minter Minter) sdk.Dec {
	if height == 1 {
		yeartimeStamp(&t1, &t2)
	}
	if time.Now().After(t2) {
		yeartimeStamp(&t1, &t2)
		return minter.NextWeeklySupply()
	}
	return minter.WeeklyProvisions
}

// function to update time stamps
func yeartimeStamp(t1 *time.Time, t2 *time.Time) {
	*t1 = time.Now().UTC()
	*t2 = t1.AddDate(1, 0, 0)
}
