package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	store2 "github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Keeper of the slashing store
type Keeper struct {
	storeKey    storetypes.StoreKey
	cdc         codec.BinaryCodec
	legacyAmino *codec.LegacyAmino
	sk          types.StakingKeeper

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper creates a slashing keeper
func NewKeeper(cdc codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key storetypes.StoreKey, sk types.StakingKeeper, authority string) Keeper {
	return Keeper{
		storeKey:    key,
		cdc:         cdc,
		legacyAmino: legacyAmino,
		sk:          sk,
		authority:   authority,
	}
}

// GetAuthority returns the x/slashing module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) getStore(ctx sdk.Context) store2.StoreAPI {
	return store2.NewStoreAPI(ctx.KVStore(k.storeKey))
}

// AddPubkey sets a address-pubkey relation
func (k Keeper) AddPubkey(ctx sdk.Context, pubkey cryptotypes.PubKey) error {
	bz, err := k.cdc.MarshalInterface(pubkey)
	if err != nil {
		return err
	}
	store := k.getStore(ctx)
	key := types.AddrPubkeyRelationKey(pubkey.Address())
	store.Set(key, bz)
	return nil
}

func (k Keeper) decodePubKey(bz []byte) (cryptotypes.PubKey, error) {
	var pk cryptotypes.PubKey
	err := k.cdc.UnmarshalInterface(bz, &pk)
	if err != nil {
		return nil, err
	}
	return pk, nil
}

// GetPubkey returns the pubkey from the adddress-pubkey relation
func (k Keeper) GetPubkey(ctx sdk.Context, a cryptotypes.Address) (cryptotypes.PubKey, error) {
	store := ctx.KVStore(k.storeKey)
	pubkey, err := store2.GetAndDecode(store, k.decodePubKey, types.AddrPubkeyRelationKey(a))
	if pubkey == nil {
		return pubkey, fmt.Errorf("address %s not found", sdk.ConsAddress(a))
	}
	if err != nil {
		return pubkey, err
	}
	return pubkey, nil
}

// Slash attempts to slash a validator. The slash is delegated to the staking
// module to make the necessary validator changes.
func (k Keeper) Slash(ctx sdk.Context, consAddr sdk.ConsAddress, fraction sdk.Dec, power, distributionHeight int64, infraction stakingtypes.Infraction) {
	coinsBurned := k.sk.Slash(ctx, consAddr, distributionHeight, power, fraction, infraction)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlash,
			sdk.NewAttribute(types.AttributeKeyAddress, consAddr.String()),
			sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", power)),
			sdk.NewAttribute(types.AttributeKeyReason, types.AttributeValueDoubleSign),
			sdk.NewAttribute(types.AttributeKeyBurnedCoins, coinsBurned.String()),
		),
	)
}

// Jail attempts to jail a validator. The slash is delegated to the staking module
// to make the necessary validator changes.
func (k Keeper) Jail(ctx sdk.Context, consAddr sdk.ConsAddress) {
	k.sk.Jail(ctx, consAddr)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlash,
			sdk.NewAttribute(types.AttributeKeyJailed, consAddr.String()),
		),
	)
}

func (k Keeper) deleteAddrPubkeyRelation(ctx sdk.Context, addr cryptotypes.Address) {
	store := k.getStore(ctx)
	store.Delete(types.AddrPubkeyRelationKey(addr))
}
