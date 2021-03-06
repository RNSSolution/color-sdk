package cli

import (
	"fmt"

	"github.com/ColorPlatform/color-sdk/client/context"
	"github.com/ColorPlatform/color-sdk/codec"
	sdk "github.com/ColorPlatform/color-sdk/types"
	"github.com/ColorPlatform/color-sdk/x/mint"
	"github.com/spf13/cobra"
)

// GetCmdQueryParams implements a command to return the current minting
// parameters.
func GetCmdQueryParams(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "Query the current minting parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", mint.QuerierRoute, mint.QueryParameters)
			res, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var params mint.Params
			if err := cdc.UnmarshalJSON(res, &params); err != nil {
				return err
			}

			return cliCtx.PrintOutput(params)
		},
	}
}

// GetCmdQueryInflation implements a command to return the current minting
// inflation value.
func GetCmdQueryInflation(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "deflation",
		Short: "Query the current minting deflation value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", mint.QuerierRoute, mint.QueryInflation)
			res, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var inflation sdk.Dec
			if err := cdc.UnmarshalJSON(res, &inflation); err != nil {
				return err
			}

			return cliCtx.PrintOutput(inflation)
		},
	}
}

// GetCmdQueryAnnualProvisions implements a command to return the current minting
// annual provisions value.
func GetCmdQueryAnnualProvisions(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "weekly-provisions",
		Short: "Query the current minting weekly provisions value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", mint.QuerierRoute, mint.QueryWeeklyProvisions)
			res, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var weekly_prov sdk.Dec
			if err := cdc.UnmarshalJSON(res, &weekly_prov); err != nil {
				return err
			}

			return cliCtx.PrintOutput(weekly_prov)
		},
	}
}

// GetCmdQueryAnnualProvisions implements a command to return the current minting
// annual provisions value.
func GetCmdQueryMintingSpeed(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "minting-speed",
		Short: "Query the current minting speed value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", mint.QuerierRoute, mint.QueryMintingSpeed)
			res, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var mintingSpeed sdk.Dec
			if err := cdc.UnmarshalJSON(res, &mintingSpeed); err != nil {
				return err
			}

			return cliCtx.PrintOutput(mintingSpeed)
		},
	}
}
