package gov

import (
	"fmt"

	abci "github.com/ColorPlatform/prism/abci/types"

	"github.com/ColorPlatform/color-sdk/codec"
	sdk "github.com/ColorPlatform/color-sdk/types"
)

// query endpoints supported by the governance Querier
const (
	QueryParams    = "params"
	QueryProposals = "proposals"
	QueryProposal  = "proposal"
	QueryDeposits  = "deposits"
	QueryDeposit   = "deposit"
	QueryVotes     = "votes"
	QueryVote      = "vote"
	QueryTally     = "tally"
	QueryCycle     = "fundingcycle"
	QueryCycles    = "fundingcycles"
	ParamDeposit   = "deposit"
	ParamVoting    = "voting"
	ParamTallying  = "tallying"
)

func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryParams:
			return queryParams(ctx, path[1:], req, keeper)
		case QueryProposals:
			return queryProposals(ctx, path[1:], req, keeper)
		case QueryProposal:
			return queryProposal(ctx, path[1:], req, keeper)
		case QueryDeposits:
			return queryDeposits(ctx, path[1:], req, keeper)
		case QueryDeposit:
			return queryDeposit(ctx, path[1:], req, keeper)
		case QueryVotes:
			return queryVotes(ctx, path[1:], req, keeper)
		case QueryVote:
			return queryVote(ctx, path[1:], req, keeper)
		case QueryTally:
			return queryTally(ctx, path[1:], req, keeper)
		case QueryCycle:
			return queryFuncingCycle(ctx, path[1:], req, keeper)
		case QueryCycles:
			return queryFuncingCycles(ctx, path[1:], req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown gov query endpoint")
		}
	}
}

func queryParams(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	switch path[0] {
	case ParamDeposit:
		bz, err := codec.MarshalJSONIndent(keeper.cdc, keeper.GetDepositParams(ctx))
		if err != nil {
			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
		}
		return bz, nil
	case ParamVoting:
		bz, err := codec.MarshalJSONIndent(keeper.cdc, keeper.GetVotingParams(ctx))
		if err != nil {
			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
		}
		return bz, nil
	case ParamTallying:
		bz, err := codec.MarshalJSONIndent(keeper.cdc, keeper.GetTallyParams(ctx))
		if err != nil {
			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
		}
		return bz, nil
	default:
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("%s is not a valid query request path", req.Path))
	}
}

// Params for queries:
// - 'custom/gov/proposal'
// - 'custom/gov/deposits'
// - 'custom/gov/tally'
// - 'custom/gov/votes'
type QueryProposalParams struct {
	ProposalID uint64
}

// creates a new instance of QueryProposalParams
func NewQueryProposalParams(proposalID uint64) QueryProposalParams {
	return QueryProposalParams{
		ProposalID: proposalID,
	}
}

// nolint: unparam
func queryProposal(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params QueryProposalParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	proposal, ok := keeper.GetProposal(ctx, params.ProposalID)
	if !ok {
		return nil, ErrUnknownProposal(DefaultCodespace, params.ProposalID)
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, proposal)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// Params for query 'custom/gov/deposit'
type QueryDepositParams struct {
	ProposalID uint64
	Depositor  sdk.AccAddress
}

// creates a new instance of QueryDepositParams
func NewQueryDepositParams(proposalID uint64, depositor sdk.AccAddress) QueryDepositParams {
	return QueryDepositParams{
		ProposalID: proposalID,
		Depositor:  depositor,
	}
}

// nolint: unparam
func queryDeposit(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params QueryDepositParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	deposit, _ := keeper.GetDeposit(ctx, params.ProposalID, params.Depositor)
	bz, err := codec.MarshalJSONIndent(keeper.cdc, deposit)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// Params for query 'custom/gov/vote'
type QueryVoteParams struct {
	ProposalID uint64
	Voter      sdk.AccAddress
}

// creates a new instance of QueryVoteParams
func NewQueryVoteParams(proposalID uint64, voter sdk.AccAddress) QueryVoteParams {
	return QueryVoteParams{
		ProposalID: proposalID,
		Voter:      voter,
	}
}

// nolint: unparam
func queryVote(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params QueryVoteParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	vote, _ := keeper.GetVote(ctx, params.ProposalID, params.Voter)
	bz, err := codec.MarshalJSONIndent(keeper.cdc, vote)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// nolint: unparam
func queryDeposits(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params QueryProposalParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	var deposits []Deposit
	depositsIterator := keeper.GetDeposits(ctx, params.ProposalID)
	defer depositsIterator.Close()
	for ; depositsIterator.Valid(); depositsIterator.Next() {
		deposit := Deposit{}
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(depositsIterator.Value(), &deposit)
		deposits = append(deposits, deposit)
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, deposits)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// nolint: unparam
func queryTally(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params QueryProposalParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	proposalID := params.ProposalID

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		return nil, ErrUnknownProposal(DefaultCodespace, proposalID)
	}

	var tallyResult TallyResult

	if proposal.Status == StatusDepositPeriod {
		tallyResult = EmptyTallyResult()
	} else if proposal.Status == StatusPassed || proposal.Status == StatusRejected {
		tallyResult = proposal.FinalTallyResult
	} else {
		// proposal is in voting period
		_, tallyResult, _ = tally(ctx, keeper, proposal)
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, tallyResult)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// nolint: unparam
func queryVotes(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params QueryProposalParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)

	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	var votes []Vote
	votesIterator := keeper.GetVotes(ctx, params.ProposalID)
	defer votesIterator.Close()
	for ; votesIterator.Valid(); votesIterator.Next() {
		vote := Vote{}
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(votesIterator.Value(), &vote)
		votes = append(votes, vote)
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, votes)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// Params for query 'custom/gov/proposals'
type QueryProposalsParams struct {
	Voter          sdk.AccAddress
	Depositor      sdk.AccAddress
	ProposalStatus ProposalStatus
	Limit          uint64
}

// creates a new instance of QueryProposalsParams
func NewQueryProposalsParams(status ProposalStatus, limit uint64, voter, depositor sdk.AccAddress) QueryProposalsParams {
	return QueryProposalsParams{
		Voter:          voter,
		Depositor:      depositor,
		ProposalStatus: status,
		Limit:          limit,
	}
}

// nolint: unparam
func queryProposals(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params QueryProposalsParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	proposals := keeper.GetProposalsFiltered(ctx, params.Voter, params.Depositor, params.ProposalStatus, params.Limit)

	bz, err := codec.MarshalJSONIndent(keeper.cdc, proposals)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// creates a new instance of QueryProposalParams
func NewQueryFuncingCycleParams(fndingCycleID uint64) QueryFuncingCycleParams {
	return QueryFuncingCycleParams{
		CycleID: fndingCycleID,
	}
}

// Params for query 'custom/gov/proposals'
type QueryFuncingCycleParams struct {
	CycleID uint64
}

func queryFuncingCycle(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params QueryFuncingCycleParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	fundingCycle, found := keeper.GetFundingCycle(ctx, params.CycleID)
	if !found {
		return nil, sdk.ErrInternal("funding cycle not found")
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, fundingCycle)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryFuncingCycles(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {

	fundingCycles := keeper.GetAllFundingCycle(ctx)

	bz, err := codec.MarshalJSONIndent(keeper.cdc, fundingCycles)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}
