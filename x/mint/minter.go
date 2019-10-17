package mint

import (
	"fmt"
	"time"

	sdk "github.com/ColorPlatform/color-sdk/types"
)

// Minter represents the minting state.
type Minter struct {
	Deflation        sdk.Dec `json:"deflation"`         // current annual inflation rate
	WeeklyProvisions sdk.Dec `json:"weekly_provisions"` // current weekly expected provisions
	MintingSpeed sdk.Dec `json:"minting_speed"` //current minting speed per second
	DeflationTime time.Time `json:"deflation_time"` //next deflation time
	BlockTime time.Time 	`json:"block_time"` //timestamp of last block
}

// NewMinter returns a new Minter object with the given inflation and annual
// provisions values.
func NewMinter(deflation, weeklyProvisions sdk.Dec, mintingspeed sdk.Dec, 
	deflationtime time.Time, blocktime time.Time) Minter {
	return Minter{
		Deflation:        deflation,
		WeeklyProvisions: weeklyProvisions,
		MintingSpeed: mintingspeed,
		DeflationTime : deflationtime,
		BlockTime	: blocktime,
	}
}

// InitialMinter returns an initial Minter object with a given inflation value.
func InitialMinter(deflation sdk.Dec) Minter {
	return NewMinter(
		deflation,
		sdk.NewDec(362880000000),
		sdk.NewDec(60000),
		time.Now().UTC().AddDate(0, 0, 7 * 52),
		time.Now().UTC(),
	)
}

// DefaultInitialMinter returns a default initial Minter object for a new chain
// which uses an inflation rate of 13%.
func DefaultInitialMinter() Minter {
	return InitialMinter(
		sdk.NewDecWithPrec(3, 2),
	)
}

func validateMinter(minter Minter) error {
	if minter.Deflation.LT(sdk.ZeroDec()) {
		return fmt.Errorf("mint parameter Deflation should be positive, is %s",
			minter.Deflation.String())
	}
	return nil
}

// NewWeeklySupply reduces the amount of weekly supply by 5%
func (m Minter) NewWeeklySupply(params Params) (sdk.Dec,sdk.Dec) {
	
	provisionAmt := m.WeeklyProvisions.QuoInt(sdk.NewInt(int64(params.BlocksPerWeek)))
	return m.WeeklyProvisions.Sub(m.Deflation.Mul(m.WeeklyProvisions)),
	provisionAmt
}

// BlockProvision returns the provisions for a block based on the annual
// provisions rate.
func (m Minter) BlockProvision(params Params) sdk.Coin {
	// provisionAmt := m.WeeklyProvisions.QuoInt(sdk.NewInt(int64(params.BlocksPerWeek)))
	fmt.Println(time.Now().UTC().Sub(m.BlockTime))
	return sdk.NewCoin(params.MintDenom, m.MintingSpeed.TruncateInt())
}