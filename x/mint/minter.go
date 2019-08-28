package mint

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Minter represents the minting state.
type Minter struct {
	Inflation        sdk.Dec `json:"inflation"`         // current annual inflation rate
	WeeklyProvisions sdk.Dec `json:"weekly_provisions"` // current Weekly expected provisions
}

// NewMinter returns a new Minter object with the given inflation and annual
// provisions values.
func NewMinter(inflation, weeklyProvisions sdk.Dec) Minter {
	return Minter{
		Inflation:        inflation,
		WeeklyProvisions: weeklyProvisions,
	}
}

// InitialMinter returns an initial Minter object with a given inflation value.
func InitialMinter(inflation sdk.Dec) Minter {
	return NewMinter(
		inflation,
		sdk.NewDec(300000000000),
	)
}

// DefaultInitialMinter returns a default initial Minter object for a new chain
// which uses an inflation rate of 13%.
func DefaultInitialMinter() Minter {
	return InitialMinter(
		sdk.NewDecWithPrec(5, 2),
	)
}

func validateMinter(minter Minter) error {
	if minter.Inflation.LT(sdk.ZeroDec()) {
		return fmt.Errorf("mint parameter Inflation should be positive, is %s",
			minter.Inflation.String())
	}
	return nil
}

// NextWeeklySupply reduces the amount of weekly supply by 5%
func (m Minter) NextWeeklySupply() sdk.Dec {
	return m.WeeklyProvisions.Sub(m.Inflation.Mul(m.WeeklyProvisions))
}

// BlockProvision returns the provisions for a block based on the annual
// provisions rate.
func (m Minter) BlockProvision(params Params) sdk.Coin {
	provisionAmt := m.WeeklyProvisions.QuoInt(sdk.NewInt(int64(params.BlocksPerWeek)))
	return sdk.NewCoin(params.MintDenom, provisionAmt.TruncateInt())
}
