package dependency

import (
	"fmt"

	"github.com/hashicorp/hcat/dep"
	"github.com/pkg/errors"
)

var (
	// Ensure implements
	_ isDependency = (*KVExistsGetQuery)(nil)
)

// KVExistsGetQuery queries the KV store for a single key.
type KVExistsGetQuery struct {
	KVExistsQuery
}

// NewKVExistsGetQueryV1 processes options in the format of "key key=value"
// e.g. "my/key dc=dc1"
func NewKVExistsGetQueryV1(key string, opts []string) (*KVExistsGetQuery, error) {
	if key == "" || key == "/" {
		return nil, fmt.Errorf("kv.exists.get: key required")
	}

	q, err := NewKVExistsQueryV1(key, opts)
	if err != nil {
		return nil, err
	}
	return &KVExistsGetQuery{KVExistsQuery: *q}, nil
}

// CanShare returns a boolean if this dependency is shareable.
func (d *KVExistsGetQuery) CanShare() bool {
	return true
}

// String returns the human-friendly version of this dependency.
func (d *KVExistsGetQuery) String() string {
	key := d.key
	if d.dc != "" {
		key = key + "dc=" + d.dc
	}
	if d.ns != "" {
		key = key + "ns=" + d.ns
	}
	return fmt.Sprintf("kv.exists.get(%s)", key)
}

// Stop halts the dependency's fetch function.
func (d *KVExistsGetQuery) Stop() {
	close(d.stopCh)
}

func (d *KVExistsGetQuery) SetOptions(opts QueryOptions) {
	d.opts = opts
}

// Fetch queries the Consul API defined by the given client.
func (d *KVExistsGetQuery) Fetch(clients dep.Clients) (interface{}, *dep.ResponseMetadata, error) {
	select {
	case <-d.stopCh:
		return nil, nil, ErrStopped
	default:
	}

	opts := d.opts.Merge(&QueryOptions{
		Datacenter: d.dc,
		Namespace:  d.ns,
	})

	pair, qm, err := clients.Consul().KV().Get(d.key, opts.ToConsulOpts())
	if err != nil {
		return nil, nil, errors.Wrap(err, d.String())
	}

	rm := &dep.ResponseMetadata{
		LastIndex:   qm.LastIndex,
		LastContact: qm.LastContact,
	}

	if pair == nil {
		return &dep.KeyPair{Exists: false}, rm, nil
	}

	return &dep.KeyPair{
		Path:        pair.Key,
		Key:         pair.Key,
		Value:       string(pair.Value),
		Exists:      true,
		CreateIndex: pair.CreateIndex,
		ModifyIndex: pair.ModifyIndex,
		LockIndex:   pair.LockIndex,
		Flags:       pair.Flags,
		Session:     pair.Session,
	}, rm, nil
}
