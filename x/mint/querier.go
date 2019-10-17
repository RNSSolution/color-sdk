package mint

import (
	"fmt"

	"github.com/ColorPlatform/color-sdk/codec"
	sdk "github.com/ColorPlatform/color-sdk/types"
	abci "github.com/ColorPlatform/prism/abci/types"
)

// Query endpoints supported by the minting querier
const (
	QueryParameters       = "parameters"
	QueryInflation        = "deflation"
	QueryWeeklyProvisions = "weekly_provisions"
	QueryMintingSpeed = "minting_speed"
)

// NewQuerier returns a minting Querier handler.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, _ abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryParameters:
			return queryParams(ctx, k)

		case QueryInflation:
			return queryInflation(ctx, k)

		case QueryWeeklyProvisions:
			return queryWeeklyProvisions(ctx, k)
		
		case QueryMintingSpeed:
			return queryMintingSpeed(ctx,k)

		default:
			return nil, sdk.ErrUnknownRequest(fmt.Sprintf("unknown minting query endpoint: %s", path[0]))
		}
	}
}

func queryParams(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(k.cdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal JSON", err.Error()))
	}

	return res, nil
}

func queryInflation(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	minter := k.GetMinter(ctx)

	res, err := codec.MarshalJSONIndent(k.cdc, minter.Deflation)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal JSON", err.Error()))
	}

	return res, nil
}

func queryWeeklyProvisions(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	minter := k.GetMinter(ctx)

	res, err := codec.MarshalJSONIndent(k.cdc, minter.WeeklyProvisions)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal JSON", err.Error()))
	}

	return res, nil
}

func queryMintingSpeed(ctx sdk.Context, k Keeper) ([]byte, sdk.Error){
	minter := k.GetMinter(ctx)

	res, err := codec.MarshalJSONIndent(k.cdc, minter.MintingSpeed)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal JSON", err.Error()))
	}

	return res, nil
}
