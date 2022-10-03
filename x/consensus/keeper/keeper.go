package keeper

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	store2 "github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/consensus/exported"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"
)

var _ exported.ConsensusParamSetter = (*Keeper)(nil)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	authority string
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, authority string) Keeper {
	return Keeper{
		storeKey:  storeKey,
		cdc:       cdc,
		authority: authority,
	}
}

func (k *Keeper) GetAuthority() string {
	return k.authority
}

func (k *Keeper) decodeConsensusParams(bz []byte) (*tmproto.ConsensusParams, error) {
	if bz == nil {
		return nil, nil
	}
	cp := &tmproto.ConsensusParams{}
	if err := k.cdc.Unmarshal(bz, cp); err != nil {
		return nil, err
	}
	return cp, nil
}

// Get gets the consensus parameters
func (k *Keeper) Get(ctx sdk.Context) (*tmproto.ConsensusParams, error) {
	store := ctx.KVStore(k.storeKey)

	cp, err := store2.GetAndDecode(store, k.decodeConsensusParams, types.ParamStoreKeyConsensusParams)
	if err != nil {
		panic(err)
	}

	return cp, nil
}

func (k *Keeper) Has(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.ParamStoreKeyConsensusParams)
}

// Set sets the consensus parameters
func (k *Keeper) Set(ctx sdk.Context, cp *tmproto.ConsensusParams) {
	store := ctx.KVStore(k.storeKey)
	store2.Set(store, types.ParamStoreKeyConsensusParams, k.cdc.MustMarshal(cp))
}
