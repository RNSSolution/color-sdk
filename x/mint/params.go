package mint

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// mint parameters
type Params struct {
	MintDenom     string `json:"mint_denom"`      // type of coin to mint
	BlocksPerWeek uint64 `json:"blocks_per_week"` // expected blocks per week
}

func NewParams(mintDenom string, inflationRateChange, inflationMax,
	inflationMin, goalBonded sdk.Dec, blocksPerWeek uint64) Params {

	return Params{
		MintDenom:     mintDenom,
		BlocksPerWeek: blocksPerWeek,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:     sdk.DefaultBondDenom,
		BlocksPerWeek: uint64(60 * 60 * 24 * 7 / 5), // assuming 5 second block times
	}
}

func validateParams(params Params) error {

	if params.MintDenom == "" {
		return fmt.Errorf("mint parameter MintDenom can't be an empty string")
	}
	return nil
}

func (p Params) String() string {
	return fmt.Sprintf(`Minting Params:
  Mint Denom:             %s
  Blocks Per Year:        %d
  `,
		p.MintDenom, p.BlocksPerWeek,
	)
}
