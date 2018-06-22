package connsec

import (
	"context"
	"net"

	inet "gx/ipfs/QmYj8wdn5sZEHX2XMDWGBvcXJNdzVbaVpHmXvhHBVZepen/go-libp2p-net"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// A Transport turns inbound and outbound unauthenticated,
// plain-text connections into authenticated, encrypted connections.
type Transport interface {
	// SecureInbound secures an inbound connection.
	SecureInbound(ctx context.Context, insecure net.Conn) (Conn, error)

	// SecureOutbound secures an outbound connection.
	SecureOutbound(ctx context.Context, insecure net.Conn, p peer.ID) (Conn, error)
}

// Conn is an authenticated, encrypted connection.
type Conn interface {
	net.Conn
	inet.ConnSecurity
}
