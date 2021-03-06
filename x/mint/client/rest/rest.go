package rest

import (
	"github.com/ColorPlatform/color-sdk/client/context"
	"github.com/ColorPlatform/color-sdk/codec"
	"github.com/gorilla/mux"
)

// RegisterRoutes registers minting module REST handlers on the provided router.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	registerQueryRoutes(cliCtx, r, cdc)
}
