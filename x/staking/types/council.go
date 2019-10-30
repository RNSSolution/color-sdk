package types

import (
	"fmt"
	
	"github.com/ColorPlatform/color-sdk/codec"
	sdk "github.com/ColorPlatform/color-sdk/types"
)


//CouncilMember : struct for council members
type CouncilMember struct{
	MemberAddress	sdk.AccAddress `json:"member_address"`
	Shares			sdk.Dec        `json:"shares"`
}

func NewCouncilMember(memberAddr sdk.AccAddress, shares sdk.Dec) CouncilMember {

	return CouncilMember{
		MemberAddress: memberAddr,
		Shares:           shares,
	}
}

// return the council member
func MustMarshalCouncilMember(cdc *codec.Codec, councilmember CouncilMember) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(councilmember)
}

// return the council member
func MustUnmarshalCouncilMember(cdc *codec.Codec, value []byte) CouncilMember {
	councilmember, err := UnmarshalCouncilMember(cdc, value)
	if err != nil {
		panic(err)
	}
	return councilmember
}

// return the council member
func UnmarshalCouncilMember(cdc *codec.Codec, value []byte) (councilmember CouncilMember, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &councilmember)
	return councilmember, err
}

func (cm CouncilMember) String() string {
	return fmt.Sprintf(`Council member :
  Address:                 %s
  Shares:          %v`, 
  cm.MemberAddress, cm.Shares)
}
	
