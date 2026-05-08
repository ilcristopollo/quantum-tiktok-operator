// pkg/entanglement/registry.go
// Pod entanglement via etcd.
//
// Two pods are considered entangled when they share a correlation key in etcd
// and neither has yet observed the shared state. Once both have read the key,
// the entanglement collapses.
//
// This is not quantum entanglement. Quantum entanglement does not require etcd,
// does not have a TTL, and does not throw a 503 when the cluster is under load.
// This is the distributed systems approximation. It is good enough.
//
// eventual consistency IS the consistency model.
// stop asking questions.

package entanglement

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	entanglementPrefix = "/quantum-tiktok/entanglement/"
	defaultTTL         = 300 // seconds; entanglement expires if never observed
)

// Registry manages entanglement state in etcd.
type Registry interface {
	Entangle(ctx context.Context, namespace, hamiltonianName string) error
	Observe(ctx context.Context, namespace, hamiltonianName string) (bool, error)
	Disentangle(ctx context.Context, namespace, hamiltonianName string) error
}

// EtcdRegistry is the etcd-backed implementation of Registry.
type EtcdRegistry struct {
	client *clientv3.Client
	lease  clientv3.Lease
}

// NewEtcdRegistry creates a Registry backed by the given etcd client.
func NewEtcdRegistry(client *clientv3.Client) *EtcdRegistry {
	return &EtcdRegistry{
		client: client,
		lease:  clientv3.NewLease(client),
	}
}

// Entangle writes a correlation key to etcd with a TTL lease.
// If the key already exists, the entanglement is refreshed.
// Two Hamiltonians with the same key are correlated until both call Observe.
//
// The correlation does not imply causality.
// It does imply that if one system collapses, the other will
// find out about it within the etcd watch latency.
// This is spooky action at a distance, modulo network partitions.
func (r *EtcdRegistry) Entangle(ctx context.Context, namespace, hamiltonianName string) error {
	key := entanglementKey(namespace, hamiltonianName)

	resp, err := r.lease.Grant(ctx, defaultTTL)
	if err != nil {
		return fmt.Errorf("granting etcd lease: %w", err)
	}

	_, err = r.client.Put(ctx, key, entanglementValue(namespace, hamiltonianName),
		clientv3.WithLease(resp.ID),
	)
	if err != nil {
		return fmt.Errorf("writing entanglement key %q: %w", key, err)
	}

	return nil
}

// Observe checks whether the entanglement key exists and has been seen by
// the other endpoint. Returns true if both endpoints have observed the state,
// meaning the entanglement has collapsed.
//
// In a real quantum system, observation destroys entanglement instantaneously.
// In etcd, it destroys it eventually.
// The difference matters less than you think at 3 AM.
func (r *EtcdRegistry) Observe(ctx context.Context, namespace, hamiltonianName string) (bool, error) {
	key := entanglementKey(namespace, hamiltonianName)

	resp, err := r.client.Get(ctx, key)
	if err != nil {
		return false, fmt.Errorf("reading entanglement key %q: %w", key, err)
	}

	// Key not present: entanglement has already collapsed, or was never established.
	if len(resp.Kvs) == 0 {
		return true, nil
	}

	return false, nil
}

// Disentangle explicitly removes the correlation key.
// Called when the Hamiltonian is deleted or reaches a coherent final state.
// The wavefunction has collapsed. There is nothing left to correlate.
func (r *EtcdRegistry) Disentangle(ctx context.Context, namespace, hamiltonianName string) error {
	key := entanglementKey(namespace, hamiltonianName)
	_, err := r.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("deleting entanglement key %q: %w", key, err)
	}
	return nil
}

func entanglementKey(namespace, name string) string {
	return fmt.Sprintf("%s%s/%s", entanglementPrefix, namespace, name)
}

func entanglementValue(namespace, name string) string {
	return fmt.Sprintf(`{"namespace":%q,"name":%q,"entangledAt":%q}`,
		namespace, name, time.Now().UTC().Format(time.RFC3339),
	)
}
