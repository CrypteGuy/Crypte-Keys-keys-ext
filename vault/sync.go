package vault

import (
	"context"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/docs"
	"github.com/keys-pub/keys/docs/events"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// SyncStatus is status of sync.
type SyncStatus struct {
	KID      keys.ID
	Salt     []byte
	SyncedAt time.Time
}

// Sync vault.
func (v *Vault) Sync(ctx context.Context) error {
	v.mtx.Lock()
	defer v.mtx.Unlock()
	logger.Infof("Syncing...")

	if err := v.push(ctx); err != nil {
		return errors.Wrapf(err, "failed to push vault")
	}
	if err := v.pull(ctx); err != nil {
		return errors.Wrapf(err, "failed to pull vault")
	}

	if err := v.setLastSync(time.Now()); err != nil {
		return err
	}

	return nil
}

// SyncStatus returns status for sync, or nil, if no sync has been performed.
func (v *Vault) SyncStatus() (*SyncStatus, error) {
	lastSync, err := v.lastSync()
	if err != nil {
		return nil, err
	}
	if lastSync.IsZero() {
		return nil, nil
	}
	remote := v.Remote()
	if remote == nil {
		return nil, nil
	}
	return &SyncStatus{
		KID:      remote.Key.ID(),
		Salt:     remote.Salt,
		SyncedAt: lastSync,
	}, nil
}

// Unsync removes vault from the remote and resets the vault log.
//
// The steps for "unsyncing" are:
// - Delete the vault from the server
// - Reset log (move pull into push)
// - Clear status (last synced, push, pull, nonces, rsalt)
// - Clear remote
func (v *Vault) Unsync(ctx context.Context) error {
	if v.remote == nil {
		return errors.Errorf("no remote set")
	}
	if v.mk == nil {
		return errors.Errorf("vault is locked")
	}

	// Delete vault from the server
	if err := v.client.VaultDelete(ctx, v.remote.Key); err != nil {
		return err
	}

	// Reset log (move pull into push)
	if err := v.resetLog(); err != nil {
		return err
	}

	// Clear status (last synced,index, nonces)
	if err := v.setLastSync(time.Time{}); err != nil {
		return err
	}
	if err := v.setPullIndex(0); err != nil {
		return err
	}
	if err := v.clearNonces(); err != nil {
		return err
	}

	// Clear remote
	if err := v.clearRemote(); err != nil {
		return err
	}

	return nil
}

func (v *Vault) resetLog() error {
	push, err := v.store.Documents(docs.Prefix(docs.Path("push")))
	if err != nil {
		return err
	}

	pull, err := v.store.Documents(docs.Prefix(docs.Path("pull")))
	if err != nil {
		return err
	}
	if len(pull) == 0 {
		return nil
	}

	if err := v.setPushIndex(int64(len(pull) + len(push))); err != nil {
		return err
	}

	// Move push to the end
	index := int64(len(pull))
	for _, doc := range push {
		index++
		path := docs.PathFrom(doc.Path, 2)
		push := docs.Path("push", pad(index), path)
		if err := v.store.Set(push, doc.Data); err != nil {
			return err
		}
	}

	// Move pull back to push
	index = int64(0)
	for _, doc := range pull {
		index++
		var event events.Event
		if err := msgpack.Unmarshal(doc.Data, &event); err != nil {
			return err
		}
		path := docs.PathFrom(doc.Path, 2)
		push := docs.Path("push", pad(index), path)
		if err := v.store.Set(push, event.Data); err != nil {
			return err
		}
		if _, err := v.store.Delete(doc.Path); err != nil {
			return err
		}
	}

	return nil
}

func (v *Vault) pullIndex() (int64, error) {
	return v.getInt64("/sync/pull")
}

func (v *Vault) setPullIndex(n int64) error {
	return v.setInt64("/sync/pull", n)
}

func (v *Vault) pushIndex() (int64, error) {
	return v.getInt64("/sync/push")
}

func (v *Vault) setPushIndex(n int64) error {
	return v.setInt64("/sync/push", n)
}

func (v *Vault) pushIndexNext() (int64, error) {
	n, err := v.pushIndex()
	if err != nil {
		return 0, err
	}
	n++
	if err := v.setPushIndex(n); err != nil {
		return 0, err
	}
	return n, nil
}

func (v *Vault) autoSyncDisabled() (bool, error) {
	return v.getBool("/sync/autoDisabled")
}

// func (v *Vault) setAutoSyncDisabled(b bool) error {
// 	return v.setBool("/sync/autoDisabled", b)
// }

func (v *Vault) lastSync() (time.Time, error) {
	return v.getTime("/sync/lastSync")
}

func (v *Vault) setLastSync(tm time.Time) error {
	return v.setTime("/sync/lastSync", tm)
}

func (v *Vault) setRemoteSalt(b []byte) error {
	return v.setValue("/sync/rsalt", b)
}

func (v *Vault) getRemoteSalt(init bool) ([]byte, error) {
	salt, err := v.getValue("/sync/rsalt")
	if err != nil {
		return nil, err
	}
	if salt == nil && init {
		salt = keys.RandBytes(32)
		if err := v.setRemoteSalt(salt); err != nil {
			return nil, err
		}
	}
	return salt, nil
}

func (v *Vault) checkNonce(n []byte) error {
	nb := encoding.MustEncode(n, encoding.Base62)
	b, err := v.store.Get(docs.Path("sync", "nonces", nb))
	if err != nil {
		return err
	}
	if b != nil {
		return errors.Errorf("nonce collision %s", nb)
	}
	return nil
}

func (v *Vault) commitNonce(n []byte) error {
	nb := encoding.MustEncode(n, encoding.Base62)
	if err := v.store.Set(docs.Path("sync", "nonces", nb), []byte{0x01}); err != nil {
		return err
	}
	return nil
}

func (v *Vault) clearNonces() error {
	docs, err := v.store.Documents(docs.Prefix(docs.Path("sync", "nonces")), docs.NoData())
	if err != nil {
		return err
	}
	for _, doc := range docs {
		if _, err := v.store.Delete(doc.Path); err != nil {
			return err
		}
	}
	return nil
}
