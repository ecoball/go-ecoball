package routinghelpers

import (
	"context"

	routing "gx/ipfs/QmXijJ3T9MjB2v8xpFDoEX6FqR9u8PkJkzu49TgwJ8Ndr5/go-libp2p-routing"
	ropts "gx/ipfs/QmXijJ3T9MjB2v8xpFDoEX6FqR9u8PkJkzu49TgwJ8Ndr5/go-libp2p-routing/options"

	pstore "gx/ipfs/QmZb7hAgQEhW9dBbzBudU39gCeD4zbe6xafD52LUuF4cUN/go-libp2p-peerstore"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	cid "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"
)

// Null is a router that doesn't do anything.
type Null struct{}

// PutValue always returns ErrNotSupported
func (nr Null) PutValue(context.Context, string, []byte, ...ropts.Option) error {
	return routing.ErrNotSupported
}

// GetValue always returns ErrNotFound
func (nr Null) GetValue(context.Context, string, ...ropts.Option) ([]byte, error) {
	return nil, routing.ErrNotFound
}

// Provide always returns ErrNotSupported
func (nr Null) Provide(context.Context, *cid.Cid, bool) error {
	return routing.ErrNotSupported
}

// FindProvidersAsync always returns a closed channel
func (nr Null) FindProvidersAsync(context.Context, *cid.Cid, int) <-chan pstore.PeerInfo {
	ch := make(chan pstore.PeerInfo)
	close(ch)
	return ch
}

// FindPeer always returns ErrNotFound
func (nr Null) FindPeer(context.Context, peer.ID) (pstore.PeerInfo, error) {
	return pstore.PeerInfo{}, routing.ErrNotFound
}

// Bootstrap always succeeds instantly
func (nr Null) Bootstrap(context.Context) error {
	return nil
}

var _ routing.IpfsRouting = Null{}
