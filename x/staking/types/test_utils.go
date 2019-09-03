package types

import (
	"github.com/ColorPlatform/prism/crypto"
	"github.com/ColorPlatform/prism/crypto/ed25519"

	sdk "github.com/RNSSolution/color-sdk/types"
)

var (
	pk1   = ed25519.GenPrivKey().PubKey()
	pk2   = ed25519.GenPrivKey().PubKey()
	pk3   = ed25519.GenPrivKey().PubKey()
	addr1 = sdk.ValAddress(pk1.Address())
	addr2 = sdk.ValAddress(pk2.Address())
	addr3 = sdk.ValAddress(pk3.Address())

	emptyAddr   sdk.ValAddress
	emptyPubkey crypto.PubKey
)
