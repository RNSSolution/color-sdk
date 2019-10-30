package keeper

import (
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
	memAddr sdk.AccAddress) (cm types.CouncilMember, found bool) {

	store := ctx.KVStore(k.storeKey)
	key := GetCouncilMemberKey(memAddr)
	value := store.Get(key)

	if value == nil {
		return cm, false
	}

	cm = types.MustUnmarshalCouncilMember(k.cdc, value)
	return cm, true
}

// SetCouncilMemberShares : Sets new shares value of the council member
func (k Keeper) SetCouncilMemberShares(ctx sdk.Context, memAddr sdk.AccAddress, newVal sdk.Dec) bool {
	cm, found := k.GetCouncilMember(ctx, memAddr)

	if found {
		cm.Shares = newVal
		k.SetCouncilMember(ctx, cm)
		return true
	}
	return false
}

// DeleteCouncilMember : delete a council member
func (k Keeper) DeleteCouncilMember(ctx sdk.Context, memAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := GetCouncilMemberKey(memAddr)
	store.Delete(key)
}

// GetCouncilMemberIterator :
func (k Keeper) GetCouncilMemberIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, CouncilMembersKey)
}

// GetCouncilMemberShares : get the shares of a council member
func (k Keeper) GetCouncilMemberShares(ctx sdk.Context, memAddr sdk.AccAddress) (sdk.Dec, bool) {
	cm, found := k.GetCouncilMember(ctx, memAddr)
	if found {
		return cm.Shares, true
	}
	return sdk.ZeroDec(), false

}

// GetAllCouncilMembers : get all the council members
func (k Keeper) GetAllCouncilMembers(ctx sdk.Context) (councilMembers []types.CouncilMember) {
	iterator := k.GetCouncilMemberIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		cm := types.MustUnmarshalCouncilMember(k.cdc, iterator.Value())
		councilMembers = append(councilMembers, cm)
	}
	return councilMembers
}
