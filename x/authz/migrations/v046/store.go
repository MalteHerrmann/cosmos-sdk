package v046

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/internal/conv"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// MigrateStore performs in-place store migrations from v0.45 to v0.46. The
// migration includes:
//
// - pruning expired authorizations
// - create secondary index for pruning expired authorizations
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	err := addExpiredGrantsIndex(ctx, store, cdc)
	if err != nil {
		return err
	}

	return nil
}

func addExpiredGrantsIndex(ctx sdk.Context, st storetypes.KVStore, cdc codec.BinaryCodec) error {
	grantsStore := prefix.NewStore(st, GrantPrefix)
	newStore := store.NewStoreAPI(grantsStore)

	grantsIter := grantsStore.Iterator(nil, nil)
	defer grantsIter.Close()

	queueItems := make(map[string][]string)
	now := ctx.BlockTime()
	for ; grantsIter.Valid(); grantsIter.Next() {
		var grant authz.Grant
		bz := grantsIter.Value()
		if err := cdc.Unmarshal(bz, &grant); err != nil {
			return err
		}

		// delete expired authorization
		// before 0.46 Expiration was required so it's safe to dereference
		if grant.Expiration.Before(now) {
			newStore.Delete(grantsIter.Key())
		} else {
			granter, grantee, msgType := ParseGrantKey(grantsIter.Key())
			// before 0.46 expiration was not a pointer, so now it's safe to dereference
			key := GrantQueueKey(*grant.Expiration, granter, grantee)

			queueItem, ok := queueItems[conv.UnsafeBytesToStr(key)]
			if !ok {
				queueItems[string(key)] = []string{msgType}
			} else {
				queueItem = append(queueItem, msgType)
				queueItems[string(key)] = queueItem
			}
		}
	}

	for key, v := range queueItems {
		bz, err := cdc.Marshal(&authz.GrantQueueItem{
			MsgTypeUrls: v,
		})
		if err != nil {
			return err
		}
		newStore.Set(conv.UnsafeStrToBytes(key), bz)
	}

	return nil
}
