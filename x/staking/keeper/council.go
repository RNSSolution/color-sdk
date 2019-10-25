package keeper

import (
	// "bytes"
	// "fmt"
	// "time"

	sdk "github.com/ColorPlatform/color-sdk/types"
	"github.com/ColorPlatform/color-sdk/x/staking/types"
)

// set a council member
func (k Keeper) SetCouncilMember(ctx sdk.Context, member types.CouncilMember) {
	// store := ctx.KVStore(k.storeKey)
	// b := types.MustMarshalCouncilMember(k.cdc, member)
	// store.Set(GetDelegationKey(delegation.DelegatorAddress, delegation.ValidatorAddress), b)
}

