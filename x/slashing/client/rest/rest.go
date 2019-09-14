package rest

import (
	"github.com/gorilla/mux"

	"github.com/ColorPlatform/color-sdk/client/context"
	"github.com/ColorPlatform/color-sdk/codec"
	"github.com/ColorPlatform/color-sdk/crypto/keys"
)

// RegisterRoutes registers staking-related REST handlers to a router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, kb keys.Keybase) {
	registerQueryRoutes(cliCtx, r, cdc)
	registerTxRoutes(cliCtx, r, cdc, kb)
}
