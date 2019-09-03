package rpc

import (
	amino "github.com/tendermint/go-amino"
	ctypes "github.com/ColorPlatform/prism/rpc/core/types"
)

var cdc = amino.NewCodec()

func init() {
	ctypes.RegisterAmino(cdc)
}
