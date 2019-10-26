package keeper

import (
	// "bytes"
	// "fmt"
	// "time"

	sdk "github.com/ColorPlatform/color-sdk/types"
	"github.com/ColorPlatform/color-sdk/x/staking/types"
)

// SetCouncilMember set a council member
func (k Keeper) SetCouncilMember(ctx sdk.Context, member types.CouncilMember) {
	 store := ctx.KVStore(k.storeKey)
	 b := types.MustMarshalCouncilMember(k.cdc, member)
	 store.Set(GetCouncilMemberKey(member.MemberAddress), b)
}

// GetCouncilMember gets council member
func (k Keeper) GetCouncilMember(ctx sdk.Context, 
	memAddr sdk.AccAddress) (cm types.CouncilMember,found bool) {

	store := ctx.KVStore(k.storeKey)
	key := GetCouncilMemberKey(memAddr)
	value := store.Get(key)

	if value == nil {
		return cm, false
	}

	cm = types.MustUnmarshalCouncilMember(k.cdc,value)
	return cm,true
}

// SetShares : Sets new shares value of the council member 
func (k Keeper) SetShares(ctx sdk.Context, memAddr sdk.AccAddress, newVal sdk.Dec) (bool){
	cm, found :=k.GetCouncilMember(ctx,memAddr)

	if found{
		cm.Shares = newVal
		k.SetCouncilMember(ctx,cm)
		return true
	}
	return false
}

func (k Keeper) DeleteCouncilMember(ctx sdk.Context, memAddr sdk.AccAddress) {
	
}