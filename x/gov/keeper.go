package gov

import (
	"fmt"
	"time"

	codec "github.com/ColorPlatform/color-sdk/codec"
	sdk "github.com/ColorPlatform/color-sdk/types"
	distr "github.com/ColorPlatform/color-sdk/x/distribution"
	"github.com/ColorPlatform/color-sdk/x/mint"
	"github.com/ColorPlatform/color-sdk/x/params"

	"github.com/ColorPlatform/prism/crypto"
)

const (
	// ModuleName is the name of the module
	ModuleName = "gov"

	// StoreKey is the store key string for gov
	StoreKey = ModuleName

	// RouterKey is the message route for gov
	RouterKey = ModuleName

	// QuerierRoute is the querier route for gov
	QuerierRoute = ModuleName

	// DefaultParamspace Parameter store default namestore
	DefaultParamspace = ModuleName
)

// Parameter store key
var (
	ParamStoreKeyDepositParams      = []byte("depositparams")
	ParamStoreKeyVotingParams       = []byte("votingparams")
	ParamStoreKeyTallyParams        = []byte("tallyparams")
	ParamStoreKeyFundingCycleParams = []byte("fundingcycleprams")

	// TODO: Find another way to implement this without using accounts, or find a cleaner way to implement it using accounts.
	DepositedCoinsAccAddr     = sdk.AccAddress(crypto.AddressHash([]byte("govDepositedCoins")))
	BurnedDepositCoinsAccAddr = sdk.AccAddress(crypto.AddressHash([]byte("govBurnedDepositCoins")))
	FourWeeksProvission       = sdk.NewDec(4)
)

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		ParamStoreKeyDepositParams, DepositParams{},
		ParamStoreKeyVotingParams, VotingParams{},
		ParamStoreKeyTallyParams, TallyParams{},
	)
}

// Keeper Governance Keeper
type Keeper struct {
	// The reference to the Param Keeper to get and set Global Params
	paramsKeeper params.Keeper
	// The reference to the distribution keeper
	distrKeeper distr.Keeper

	minKeeper mint.Keeper

	// The reference to the Paramstore to get and set gov specific params
	paramSpace params.Subspace

	// The reference to the CoinKeeper to modify balances
	ck BankKeeper

	// The reference to the StakingKeeper
	stk StakingKeeper

	// The ValidatorSet to get information about validators
	vs sdk.ValidatorSet

	// The reference to the DelegationSet to get information about delegators
	ds sdk.DelegationSet

	// The (unexposed) keys used to access the stores from the Context.
	storeKey sdk.StoreKey

	// The codec codec for binary encoding/decoding.
	cdc *codec.Codec

	// Reserved codespace
	codespace sdk.CodespaceType
}

// NewKeeper returns a governance keeper. It handles:
// - submitting governance proposals
// - depositing funds into proposals, and activating upon sufficient funds being deposited
// - users voting on proposals, with weight proportional to stake in the system
// - and tallying the result of the vote.
func NewKeeper(cdc *codec.Codec, dk distr.Keeper, mk mint.Keeper, key sdk.StoreKey, paramsKeeper params.Keeper,
	paramSpace params.Subspace, ck BankKeeper, sk StakingKeeper, ds sdk.DelegationSet, codespace sdk.CodespaceType) Keeper {

	return Keeper{
		storeKey:     key,
		distrKeeper:  dk,
		minKeeper:    mk,
		paramsKeeper: paramsKeeper,
		paramSpace:   paramSpace.WithKeyTable(ParamKeyTable()),
		ck:           ck,
		stk:          sk,
		ds:           ds,
		vs:           ds.GetValidatorSet(),
		cdc:          cdc,
		codespace:    codespace,
	}
}

// Proposals
func (keeper Keeper) SubmitProposal(ctx sdk.Context, content ProposalContent) (proposal Proposal, err sdk.Error) {
	proposalID, err := keeper.getNewProposalID(ctx)
	if err != nil {
		return
	}

	submitTime := ctx.BlockHeader().Time
	//depositPeriod := keeper.GetDepositParams(ctx).MaxDepositPeriod

	proposal = Proposal{
		ProposalContent:       content,
		ProposalID:            proposalID,
		Status:                StatusDepositPeriod,
		FinalTallyResult:      EmptyTallyResult(),
		TotalDeposit:          sdk.NewCoins(),
		RemainingFundingCycle: content.GetFundingCycle(),
		FundingCycleCount:     0,
		SubmitTime:            submitTime,
		DepositEndTime:        submitTime,
	}
	keeper.SetProposal(ctx, proposal)
	_, err = keeper.GetCurrentCycle(ctx)
	if err != nil {
		keeper.InsertInactiveProposalQueue(ctx, proposal.DepositEndTime, proposalID)
	} else {
		keeper.activateVotingPeriod(ctx, proposal)
	}

	return
}

// GetProposal from store by ProposalID
func (keeper Keeper) GetProposal(ctx sdk.Context, proposalID uint64) (proposal Proposal, ok bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyProposal(proposalID))
	if bz == nil {
		return
	}
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &proposal)
	return proposal, true
}

// SetProposal Implements sdk.AccountKeeper.
func (keeper Keeper) SetProposal(ctx sdk.Context, proposal Proposal) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(proposal)
	store.Set(KeyProposal(proposal.ProposalID), bz)

}

// DeleteProposal Implements sdk.AccountKeeper.
func (keeper Keeper) DeleteProposal(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		panic("DeleteProposal cannot fail to GetProposal")
	}
	keeper.RemoveFromInactiveProposalQueue(ctx, proposal.DepositEndTime, proposalID)
	keeper.RemoveFromActiveProposalQueue(ctx, proposal.VotingEndTime, proposalID)
	store.Delete(KeyProposal(proposalID))
}

// GetProposalsFiltered Get Proposal from store by ProposalID
// voterAddr will filter proposals by whether or not that address has voted on them
// depositorAddr will filter proposals by whether or not that address has deposited to them
// status will filter proposals by status
// numLatest will fetch a specified number of the most recent proposals, or 0 for all proposals
func (keeper Keeper) GetProposalsFiltered(ctx sdk.Context, voterAddr sdk.AccAddress, depositorAddr sdk.AccAddress, status ProposalStatus, numLatest uint64) []Proposal {

	maxProposalID, err := keeper.peekCurrentProposalID(ctx)
	if err != nil {
		return nil
	}

	matchingProposals := []Proposal{}

	if numLatest == 0 {
		numLatest = maxProposalID
	}

	for proposalID := maxProposalID - numLatest; proposalID < maxProposalID; proposalID++ {
		if voterAddr != nil && len(voterAddr) != 0 {
			_, found := keeper.GetVote(ctx, proposalID, voterAddr)
			if !found {
				continue
			}
		}

		if depositorAddr != nil && len(depositorAddr) != 0 {
			_, found := keeper.GetDeposit(ctx, proposalID, depositorAddr)
			if !found {
				continue
			}
		}

		proposal, ok := keeper.GetProposal(ctx, proposalID)
		if !ok {
			continue
		}

		if validProposalStatus(status) {
			if proposal.Status != status {
				continue
			}
		}

		matchingProposals = append(matchingProposals, proposal)
	}
	return matchingProposals
}

// Set the initial proposal ID
func (keeper Keeper) setInitialProposalID(ctx sdk.Context, proposalID uint64) sdk.Error {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyNextProposalID)
	if bz != nil {
		return ErrInvalidGenesis(keeper.codespace, "Initial ProposalID already set")
	}
	bz = keeper.cdc.MustMarshalBinaryLengthPrefixed(proposalID)
	store.Set(KeyNextProposalID, bz)
	return nil
}

// GetLastProposalID Get the last used proposal ID
func (keeper Keeper) GetLastProposalID(ctx sdk.Context) (proposalID uint64) {
	proposalID, err := keeper.peekCurrentProposalID(ctx)
	if err != nil {
		return 0
	}
	proposalID--
	return
}

// Gets the next available ProposalID and increments it
func (keeper Keeper) getNewProposalID(ctx sdk.Context) (proposalID uint64, err sdk.Error) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyNextProposalID)
	if bz == nil {
		return 0, ErrInvalidGenesis(keeper.codespace, "InitialProposalID never set")
	}
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &proposalID)
	bz = keeper.cdc.MustMarshalBinaryLengthPrefixed(proposalID + 1)
	store.Set(KeyNextProposalID, bz)
	return proposalID, nil
}

// Peeks the next available ProposalID without incrementing it
func (keeper Keeper) peekCurrentProposalID(ctx sdk.Context) (proposalID uint64, err sdk.Error) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyNextProposalID)
	if bz == nil {
		return 0, ErrInvalidGenesis(keeper.codespace, "InitialProposalID never set")
	}
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &proposalID)
	return proposalID, nil
}

func (keeper Keeper) activateVotingPeriod(ctx sdk.Context, proposal Proposal) {
	proposal.VotingStartTime = ctx.BlockHeader().Time
	proposal.VotingEndTime = proposal.VotingStartTime
	proposal.Status = StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	keeper.RemoveFromInactiveProposalQueue(ctx, proposal.DepositEndTime, proposal.ProposalID)
	keeper.InsertActiveProposalQueue(ctx, proposal.VotingEndTime, proposal.ProposalID)
}

// Params

// GetDepositParams Returns the current DepositParams from the global param store
func (keeper Keeper) GetDepositParams(ctx sdk.Context) DepositParams {
	var depositParams DepositParams
	keeper.paramSpace.Get(ctx, ParamStoreKeyDepositParams, &depositParams)
	return depositParams
}

// GetVotingParams Returns the current VotingParams from the global param store
func (keeper Keeper) GetVotingParams(ctx sdk.Context) VotingParams {
	var votingParams VotingParams
	keeper.paramSpace.Get(ctx, ParamStoreKeyVotingParams, &votingParams)
	return votingParams
}

// GetTallyParams Returns the current TallyParam from the global param store
func (keeper Keeper) GetTallyParams(ctx sdk.Context) TallyParams {
	var tallyParams TallyParams
	keeper.paramSpace.Get(ctx, ParamStoreKeyTallyParams, &tallyParams)
	return tallyParams
}

func (keeper Keeper) setDepositParams(ctx sdk.Context, depositParams DepositParams) {
	keeper.paramSpace.Set(ctx, ParamStoreKeyDepositParams, &depositParams)
}

func (keeper Keeper) setVotingParams(ctx sdk.Context, votingParams VotingParams) {
	keeper.paramSpace.Set(ctx, ParamStoreKeyVotingParams, &votingParams)
}

func (keeper Keeper) setTallyParams(ctx sdk.Context, tallyParams TallyParams) {
	keeper.paramSpace.Set(ctx, ParamStoreKeyTallyParams, &tallyParams)
}

// Votes

// AddVote Adds a vote on a specific proposal
func (keeper Keeper) AddVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress, option VoteOption) sdk.Error {
	_, chk := keeper.stk.GetCouncilMemberShares(ctx, voterAddr)
	if chk == false {
		return ErrInvalidCouncilMember(keeper.codespace, voterAddr)
	}
	activeCycle := keeper.CheckCycleActive(ctx)
	if activeCycle == false {

		return ErrInvalidCycle(keeper.codespace, "No Active Cycle Found.")
	}
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		return ErrUnknownProposal(keeper.codespace, proposalID)
	}
	if proposal.Status != StatusVotingPeriod {
		return ErrInactiveProposal(keeper.codespace, proposalID)
	}

	if !validVoteOption(option) {
		return ErrInvalidVote(keeper.codespace, option)
	}

	vote := Vote{
		ProposalID: proposalID,
		Voter:      voterAddr,
		Option:     option,
	}
	keeper.setVote(ctx, proposalID, voterAddr, vote)

	return nil
}

// GetVote Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) GetVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) (Vote, bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyVote(proposalID, voterAddr))
	if bz == nil {
		return Vote{}, false
	}
	var vote Vote
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &vote)
	return vote, true
}

func (keeper Keeper) setVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress, vote Vote) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(vote)
	store.Set(KeyVote(proposalID, voterAddr), bz)
}

// GetVotes Gets all the votes on a specific proposal
func (keeper Keeper) GetVotes(ctx sdk.Context, proposalID uint64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, KeyVotesSubspace(proposalID))
}

func (keeper Keeper) deleteVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(KeyVote(proposalID, voterAddr))
}

// Deposits

// GetDeposit Gets the deposit of a specific depositor on a specific proposal
func (keeper Keeper) GetDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) (Deposit, bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyDeposit(proposalID, depositorAddr))
	if bz == nil {
		return Deposit{}, false
	}
	var deposit Deposit
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &deposit)
	return deposit, true
}

func (keeper Keeper) setDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress, deposit Deposit) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(deposit)
	store.Set(KeyDeposit(proposalID, depositorAddr), bz)
}

// AddDeposit Adds or updates a deposit of a specific depositor on a specific proposal
// Activates voting period when appropriate
func (keeper Keeper) AddDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress, depositAmount sdk.Coins) (sdk.Error, bool) {
	// Checks to see if proposal exists
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		return ErrUnknownProposal(keeper.codespace, proposalID), false
	}

	// Check if proposal is still depositable
	if (proposal.Status != StatusDepositPeriod) && (proposal.Status != StatusVotingPeriod) {
		return ErrAlreadyFinishedProposal(keeper.codespace, proposalID), false
	}
	// Send coins from depositor's account to DepositedCoinsAccAddr account
	// TODO: Don't use an account for this purpose; it's clumsy and prone to misuse.
	_, err := keeper.ck.SendCoins(ctx, depositorAddr, DepositedCoinsAccAddr, depositAmount)
	if err != nil {
		return err, false
	}

	// Update proposal
	proposal.TotalDeposit = proposal.TotalDeposit.Add(depositAmount)
	keeper.SetProposal(ctx, proposal)

	// Check if deposit has provided sufficient total funds to transition the proposal into the voting period
	activatedVotingPeriod := false
	if proposal.Status == StatusDepositPeriod && proposal.TotalDeposit.IsAllGTE(keeper.GetDepositParams(ctx).MinDeposit) {
		keeper.activateVotingPeriod(ctx, proposal)
		activatedVotingPeriod = true
	}

	// Add or update deposit object
	currDeposit, found := keeper.GetDeposit(ctx, proposalID, depositorAddr)
	if !found {
		newDeposit := Deposit{depositorAddr, proposalID, depositAmount}
		keeper.setDeposit(ctx, proposalID, depositorAddr, newDeposit)
	} else {
		currDeposit.Amount = currDeposit.Amount.Add(depositAmount)
		keeper.setDeposit(ctx, proposalID, depositorAddr, currDeposit)
	}

	return nil, activatedVotingPeriod
}

// GetDeposits Gets all the deposits on a specific proposal as an sdk.Iterator
func (keeper Keeper) GetDeposits(ctx sdk.Context, proposalID uint64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, KeyDepositsSubspace(proposalID))
}

// RefundDeposits Refunds and deletes all the deposits on a specific proposal
func (keeper Keeper) RefundDeposits(ctx sdk.Context, proposalID uint64) {

	// ===TOD0 add logic to transfer requted funding from treasury to this prosal
	store := ctx.KVStore(keeper.storeKey)
	depositsIterator := keeper.GetDeposits(ctx, proposalID)
	defer depositsIterator.Close()
	for ; depositsIterator.Valid(); depositsIterator.Next() {
		deposit := &Deposit{}
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(depositsIterator.Value(), deposit)
		_, err := keeper.ck.SendCoins(ctx, DepositedCoinsAccAddr, deposit.Depositor, deposit.Amount)
		if err != nil {
			panic("should not happen")
		}

		store.Delete(depositsIterator.Key())
	}
}

// TransferFunds Transfer funds from treasury to depositor
func (keeper Keeper) TransferFunds(ctx sdk.Context, proposals []Proposal) {
	totalFundCount := sdk.NewCoins()
	weeklyIncome := keeper.GetTreasuryWeeklyIncome(ctx)
	limit := sdk.NewInt(GetPercentageAmount(weeklyIncome, 0.5))

	fundingcycle, err := keeper.GetCurrentCycle(ctx)
	if err != nil {
		return
	}

	for _, proposal := range proposals {

		totalFundCount = totalFundCount.Add(proposal.GetRequestedFund())
		if VerifyAmount(totalFundCount, limit) {

			proposal = proposal.ReduceCycleCount()
			if proposal.IsZeroRemainingCycle() {
				proposal.Status = StatusPassed
				keeper.RefundDeposits(ctx, proposal.ProposalID)
				keeper.DeleteProposalEligibility(ctx, proposal)
				keeper.RemoveFromInactiveProposalQueue(ctx, proposal.DepositEndTime, proposal.ProposalID)
				keeper.RemoveFromActiveProposalQueue(ctx, proposal.VotingEndTime, proposal.ProposalID)
			}

			err := keeper.distrKeeper.DistributeFeePool(ctx, proposal.GetRequestedFund(), proposal.GetProposer())
			if err != nil {
				panic("should not happen")
			}
			fundingcycle.FundedProposals = append(fundingcycle.FundedProposals, proposal.ProposalID)
			keeper.SetProposal(ctx, proposal)
		}
	}
	keeper.SetFundingCycle(ctx, fundingcycle)

}

// DeleteDeposits Deletes all the deposits on a specific proposal without refunding them
func (keeper Keeper) DeleteDeposits(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)
	depositsIterator := keeper.GetDeposits(ctx, proposalID)
	defer depositsIterator.Close()
	for ; depositsIterator.Valid(); depositsIterator.Next() {
		deposit := &Deposit{}
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(depositsIterator.Value(), deposit)

		// TODO: Find a way to do this without using accounts.
		_, err := keeper.ck.SendCoins(ctx, DepositedCoinsAccAddr, BurnedDepositCoinsAccAddr, deposit.Amount)
		if err != nil {
			panic("should not happen")
		}

		store.Delete(depositsIterator.Key())
	}
}

// ProposalQueues

// ActiveProposalQueueIterator Returns an iterator for all the proposals in the Active Queue that expire by endTime
func (keeper Keeper) ActiveProposalQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return store.Iterator(PrefixActiveProposalQueue, sdk.PrefixEndBytes(PrefixActiveProposalQueueTime(endTime)))
}

// InsertActiveProposalQueue Inserts a ProposalID into the active proposal queue at endTime
func (keeper Keeper) InsertActiveProposalQueue(ctx sdk.Context, endTime time.Time, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(proposalID)
	store.Set(KeyActiveProposalQueueProposal(endTime, proposalID), bz)
}

// RemoveFromActiveProposalQueue removes a proposalID from the Active Proposal Queue
func (keeper Keeper) RemoveFromActiveProposalQueue(ctx sdk.Context, endTime time.Time, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(KeyActiveProposalQueueProposal(endTime, proposalID))
}

// InactiveProposalQueueIterator Returns an iterator for all the proposals in the Inactive Queue that expire by endTime
func (keeper Keeper) InactiveProposalQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return store.Iterator(PrefixInactiveProposalQueue, sdk.PrefixEndBytes(PrefixInactiveProposalQueueTime(endTime)))
}

// InsertInactiveProposalQueue Inserts a ProposalID into the inactive proposal queue at endTime
func (keeper Keeper) InsertInactiveProposalQueue(ctx sdk.Context, endTime time.Time, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(proposalID)
	store.Set(KeyInactiveProposalQueueProposal(endTime, proposalID), bz)
}

// RemoveFromInactiveProposalQueue removes a proposalID from the Inactive Proposal Queue
func (keeper Keeper) RemoveFromInactiveProposalQueue(ctx sdk.Context, endTime time.Time, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(KeyInactiveProposalQueueProposal(endTime, proposalID))
}

func (keeper Keeper) RemoveFromInactiveProposalQueueIterator(ctx sdk.Context) {

	inactiveIterator := keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	defer inactiveIterator.Close()
	for ; inactiveIterator.Valid(); inactiveIterator.Next() {
		var proposalID uint64

		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(inactiveIterator.Value(), &proposalID)
		inactiveProposal, ok := keeper.GetProposal(ctx, proposalID)
		if !ok {
			panic(fmt.Sprintf("proposal %d does not exist", proposalID))
		}
		keeper.activateVotingPeriod(ctx, inactiveProposal)
	}
}

// GetCurrentCycle return the active funding cycle
func (keeper Keeper) GetCurrentCycle(ctx sdk.Context) (FundingCycle, sdk.Error) {

	store := ctx.KVStore(keeper.storeKey)
	lastFundingCycleID := KeyFundingCycle(keeper.GetLastFundingCycleID(ctx))
	bz := store.Get(lastFundingCycleID)
	if bz == nil {
		return FundingCycle{}, ErrInvalidGenesis(keeper.codespace, "Currently no active funding cycle")
	}
	var fundingCycle FundingCycle
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &fundingCycle)
	return fundingCycle, nil
}

// GetDaysPassed calculate total days passed since blockchain start
func (keeper Keeper) GetDaysPassed(ctx sdk.Context) (int, sdk.Error) {
	firstBlockTime, err := keeper.GetBlockTime(ctx)
	if err != nil {
		return 0, err
	}
	currentBlock := ctx.BlockHeader().Time
	return int(currentBlock.Sub(firstBlockTime).Hours() / 24), nil
}

func (keeper Keeper) SetBlockTime(ctx sdk.Context) sdk.Error {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyFirstBlock)
	if bz != nil {
		return ErrInvalidGenesis(keeper.codespace, "First Block already set")
	}
	bz = keeper.cdc.MustMarshalBinaryLengthPrefixed(ctx.BlockHeader().Time)
	store.Set(KeyFirstBlock, bz)
	return nil

}

func (keeper Keeper) GetBlockTime(ctx sdk.Context) (firstblocktime time.Time, err sdk.Error) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyFirstBlock)
	if bz == nil {
		return ctx.BlockHeader().Time, ErrInvalidGenesis(keeper.codespace, "Fist block time  never set")
	}
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &firstblocktime)
	return firstblocktime, nil

}

// AddFundingCycle start a new funding cycle
func (keeper Keeper) AddFundingCycle(ctx sdk.Context) {
	fundingCycleID, err := keeper.GetNewFundingCycleID(ctx)
	if err != nil {
		panic("AddFundingCycle fail to get New Funding Cycle ID")
	}
	startTime := ctx.BlockHeader().Time
	endTime := ctx.BlockHeader().Time.Add(FourWeeksHours)

	fundingCycle := FundingCycle{
		CycleID:        fundingCycleID,
		CycleStartTime: startTime,
		CycleEndTime:   endTime,
	}
	keeper.SetFundingCycle(ctx, fundingCycle)
}

// GetNewFundingCycleID Gets the next available FundingCycleID and increments it
func (keeper Keeper) GetNewFundingCycleID(ctx sdk.Context) (fundingCycleID uint64, err sdk.Error) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyNextFundingCycleID)
	if bz == nil {
		return 0, ErrInvalidGenesis(keeper.codespace, "InitialFundingCycleID never set")
	}
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &fundingCycleID)
	bz = keeper.cdc.MustMarshalBinaryLengthPrefixed(fundingCycleID + 1)
	store.Set(KeyNextFundingCycleID, bz)
	return fundingCycleID, nil
}

// SetFundingCycle store the funding cycle in KV store
func (keeper Keeper) SetFundingCycle(ctx sdk.Context, fundingCycle FundingCycle) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(fundingCycle)
	store.Set(KeyFundingCycle(fundingCycle.CycleID), bz)

}

// setInitialFundingCycleID Set the initial funding cycle  ID
func (keeper Keeper) setInitialFundingCycleID(ctx sdk.Context, fundingCycleID uint64) sdk.Error {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyNextFundingCycleID)
	if bz != nil {
		return ErrInvalidGenesis(keeper.codespace, "Initial fundingCycleID already set")
	}
	bz = keeper.cdc.MustMarshalBinaryLengthPrefixed(fundingCycleID)
	store.Set(KeyNextFundingCycleID, bz)
	return nil
}

// GetLastFundingCycleID Get the last used funding cycle ID
func (keeper Keeper) GetLastFundingCycleID(ctx sdk.Context) (fundingCycle uint64) {
	fundingCycle, err := keeper.peekCurrentFundingCycleID(ctx)
	if err != nil {
		return 0
	}
	fundingCycle--
	return fundingCycle
}

// peekCurrentFundingCycleID Peeks the next available ProposalID without incrementing it
func (keeper Keeper) peekCurrentFundingCycleID(ctx sdk.Context) (fundingCycleID uint64, err sdk.Error) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyNextFundingCycleID)
	if bz == nil {
		return 0, ErrInvalidGenesis(keeper.codespace, "InitialFundingCycleID never set")
	}
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &fundingCycleID)
	return fundingCycleID, nil
}

// GetDeposits Gets all the deposits on a specific proposal as an sdk.Iterator
func (keeper Keeper) GetFundingCyclesIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(PrefixFudingCycleQueue))
}

func (keeper Keeper) GetAllFundingCycle(ctx sdk.Context) FundingCycles {
	fundingCycles := FundingCycles{}
	fundingCycleIterator := keeper.GetFundingCyclesIterator(ctx)
	defer fundingCycleIterator.Close()
	for ; fundingCycleIterator.Valid(); fundingCycleIterator.Next() {
		fundingCycle := &FundingCycle{}
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(fundingCycleIterator.Value(), fundingCycle)
		fundingCycles = append(fundingCycles, *fundingCycle)

	}
	return fundingCycles
}

//GetFundingCycle
func (keeper Keeper) GetFundingCycle(ctx sdk.Context, CycleID uint64) (FundingCycle, bool) {

	store := ctx.KVStore(keeper.storeKey)
	fundingCycleID := KeyFundingCycle(CycleID)
	bz := store.Get(fundingCycleID)
	if bz == nil {
		return FundingCycle{}, false
	}
	var fundingCycle FundingCycle
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &fundingCycle)
	return fundingCycle, true

}

func (keeper Keeper) SetProposalEligibility(ctx sdk.Context, proposalEligibility ProposalEligibility) sdk.Error {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(proposalEligibility)
	if bz == nil {
		return ErrInvalidEligibility(keeper.codespace, "Invalid proposal eligibility")
	}
	store.Set(KeyEligibility(proposalEligibility.ProposalID), bz)

	return nil

}
func (keeper Keeper) DeleteProposalEligibility(ctx sdk.Context, proposal Proposal) {
	proposal.Ranking = sdk.ZeroInt()
	keeper.SetProposal(ctx, proposal)
}

func (keeper Keeper) GetEligibilityIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(PrefixEligibilityQueue))
}

func (keeper Keeper) SetEligibilityDetails(ctx sdk.Context, proposals []Proposal) {
	for index, proposal := range proposals {
		updatedProposal, ok := keeper.GetProposal(ctx, proposal.ProposalID)
		if !ok {
			panic(fmt.Sprintf("proposal %d does not exist", proposal.ProposalID))
		}
		updatedProposal.Ranking = sdk.NewInt(int64(index + 1))
		keeper.SetProposal(ctx, updatedProposal)

	}

}

func (keeper Keeper) GetTreasuryWeeklyIncome(ctx sdk.Context) sdk.Dec {
	communityTx := keeper.distrKeeper.GetCommunityTax(ctx)
	weeklyProivssion := keeper.minKeeper.GetMinter(ctx).WeeklyProvisions
	treasuryIncome := weeklyProivssion.Mul(FourWeeksProvission)
	treasuryIncome = treasuryIncome.Mul(communityTx)
	return treasuryIncome

}
