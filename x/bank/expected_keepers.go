package bank

import (
	sdk "github.com/RNSSolution/color-sdk/types"
)

// expected crisis keeper
type CrisisKeeper interface {
	RegisterRoute(moduleName, route string, invar sdk.Invariant)
}
