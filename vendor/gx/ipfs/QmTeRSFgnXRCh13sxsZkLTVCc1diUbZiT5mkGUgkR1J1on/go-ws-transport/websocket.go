// Package websocket implements a websocket based transport for go-libp2p.
package websocket

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"

	mafmt "gx/ipfs/QmTy17Jm1foTnvUS9JXRhLbRQ3XuC64jPTjUfpB4mHz2QM/mafmt"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	tpt "gx/ipfs/QmYnjSGtvn7LhrxCvwrU9uDWxKyg28uBYeXvgzTDDDzVy4/go-libp2p-transport"
	ws "gx/ipfs/QmZH5VXfAJouGMyCCHTRPGCT3e5MG9Lu78Ln3YAYW1XTts/websocket"
	manet "gx/ipfs/QmcGXGdw9BWDysPJQHxJinjGHha3eEg4vzFETre4woNwcX/go-multiaddr-net"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	tptu "gx/ipfs/Qmf3ejfGWR8Bd3wKFBvwYGFMJ9TeKJwYJUc2WchXjMxzg7/go-libp2p-transport-upgrader"
)

// WsProtocol is the multiaddr protocol definition for this transport.
var WsProtocol = ma.Protocol{
	Code:  477,
	Name:  "ws",
	VCode: ma.CodeToVarint(477),
}

// WsFmt is multiaddr formatter for WsProtocol
var WsFmt = mafmt.And(mafmt.TCP, mafmt.Base(WsProtocol.Code))

// WsCodec is the multiaddr-net codec definition for the websocket transport
var WsCodec = &manet.NetCodec{
	NetAddrNetworks:  []string{"websocket"},
	ProtocolName:     "ws",
	ConvertMultiaddr: ConvertWebsocketMultiaddrToNetAddr,
	ParseNetAddr:     ParseWebsocketNetAddr,
}

// Default gorilla upgrader
var upgrader = ws.Upgrader{
	// Allow requests from *all* origins.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func init() {
	err := ma.AddProtocol(WsProtocol)
	if err != nil {
		panic(fmt.Errorf("error registering websocket protocol: %s", err))
	}

	manet.RegisterNetCodec(WsCodec)
}

// WebsocketTransport is the actual go-libp2p transport
type WebsocketTransport struct {
	Upgrader *tptu.Upgrader
}

func New(u *tptu.Upgrader) *WebsocketTransport {
	return &WebsocketTransport{u}
}

var _ tpt.Transport = (*WebsocketTransport)(nil)

func (t *WebsocketTransport) CanDial(a ma.Multiaddr) bool {
	return WsFmt.Matches(a)
}

func (t *WebsocketTransport) Protocols() []int {
	return []int{WsProtocol.Code}
}

func (t *WebsocketTransport) Proxy() bool {
	return false
}

func (t *WebsocketTransport) maDial(ctx context.Context, raddr ma.Multiaddr) (manet.Conn, error) {
	wsurl, err := parseMultiaddr(raddr)
	if err != nil {
		return nil, err
	}

	wscon, _, err := ws.DefaultDialer.Dial(wsurl, nil)
	if err != nil {
		return nil, err
	}

	mnc, err := manet.WrapNetConn(NewConn(wscon, nil))
	if err != nil {
		wscon.Close()
		return nil, err
	}
	return mnc, nil
}

func (t *WebsocketTransport) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (tpt.Conn, error) {
	macon, err := t.maDial(ctx, raddr)
	if err != nil {
		return nil, err
	}
	return t.Upgrader.UpgradeOutbound(ctx, t, macon, p)
}

func (t *WebsocketTransport) maListen(a ma.Multiaddr) (manet.Listener, error) {
	lnet, lnaddr, err := manet.DialArgs(a)
	if err != nil {
		return nil, err
	}

	nl, err := net.Listen(lnet, lnaddr)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse("http://" + nl.Addr().String())
	if err != nil {
		nl.Close()
		return nil, err
	}

	malist, err := t.wrapListener(nl, u)
	if err != nil {
		nl.Close()
		return nil, err
	}

	go malist.serve()

	return malist, nil
}

func (t *WebsocketTransport) Listen(a ma.Multiaddr) (tpt.Listener, error) {
	malist, err := t.maListen(a)
	if err != nil {
		return nil, err
	}
	return t.Upgrader.UpgradeListener(t, malist), nil
}

func (t *WebsocketTransport) wrapListener(l net.Listener, origin *url.URL) (*listener, error) {
	laddr, err := manet.FromNetAddr(l.Addr())
	if err != nil {
		return nil, err
	}
	wsma, err := ma.NewMultiaddr("/ws")
	if err != nil {
		return nil, err
	}
	laddr = laddr.Encapsulate(wsma)

	return &listener{
		laddr:    laddr,
		Listener: l,
		incoming: make(chan *Conn),
		closed:   make(chan struct{}),
	}, nil
}
