package config

import (
	"fmt"
	"reflect"

	security "gx/ipfs/QmSieFUauuYnroStqmRAEgu9BMXDNY5LbtNgzXcFitBKXQ/go-conn-security"
	pnet "gx/ipfs/QmW7Ump7YyBMr712Ta3iEVh3ZYcfVvJaPryfbCnyE826b4/go-libp2p-interface-pnet"
	mux "gx/ipfs/QmY9JXR3FupnYAYJWK9aMr9bCpqWKcToQ1tz8DVGTrHpHw/go-stream-muxer"
	filter "gx/ipfs/QmYF6XMumtqvXtZfQKrCokXEdFvfy2i7tbpvSRzRrHyY8c/go-maddr-filter"
	inet "gx/ipfs/QmYj8wdn5sZEHX2XMDWGBvcXJNdzVbaVpHmXvhHBVZepen/go-libp2p-net"
	transport "gx/ipfs/QmYnjSGtvn7LhrxCvwrU9uDWxKyg28uBYeXvgzTDDDzVy4/go-libp2p-transport"
	pstore "gx/ipfs/QmZb7hAgQEhW9dBbzBudU39gCeD4zbe6xafD52LUuF4cUN/go-libp2p-peerstore"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	host "gx/ipfs/QmdHyfNVTZ5VtUx4Xz23z8wtnioSrFQ28XSfpVkdhQBkGA/go-libp2p-host"
	crypto "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	tptu "gx/ipfs/Qmf3ejfGWR8Bd3wKFBvwYGFMJ9TeKJwYJUc2WchXjMxzg7/go-libp2p-transport-upgrader"
)

var (
	// interfaces
	hostType      = reflect.TypeOf((*host.Host)(nil)).Elem()
	networkType   = reflect.TypeOf((*inet.Network)(nil)).Elem()
	transportType = reflect.TypeOf((*transport.Transport)(nil)).Elem()
	muxType       = reflect.TypeOf((*mux.Transport)(nil)).Elem()
	securityType  = reflect.TypeOf((*security.Transport)(nil)).Elem()
	protectorType = reflect.TypeOf((*pnet.Protector)(nil)).Elem()
	privKeyType   = reflect.TypeOf((*crypto.PrivKey)(nil)).Elem()
	pubKeyType    = reflect.TypeOf((*crypto.PubKey)(nil)).Elem()
	pstoreType    = reflect.TypeOf((*pstore.Peerstore)(nil)).Elem()

	// concrete types
	peerIDType   = reflect.TypeOf((peer.ID)(""))
	filtersType  = reflect.TypeOf((*filter.Filters)(nil))
	upgraderType = reflect.TypeOf((*tptu.Upgrader)(nil))
)

var argTypes = map[reflect.Type]constructor{
	upgraderType:  func(h host.Host, u *tptu.Upgrader) interface{} { return u },
	hostType:      func(h host.Host, u *tptu.Upgrader) interface{} { return h },
	networkType:   func(h host.Host, u *tptu.Upgrader) interface{} { return h.Network() },
	muxType:       func(h host.Host, u *tptu.Upgrader) interface{} { return u.Muxer },
	securityType:  func(h host.Host, u *tptu.Upgrader) interface{} { return u.Secure },
	protectorType: func(h host.Host, u *tptu.Upgrader) interface{} { return u.Protector },
	filtersType:   func(h host.Host, u *tptu.Upgrader) interface{} { return u.Filters },
	peerIDType:    func(h host.Host, u *tptu.Upgrader) interface{} { return h.ID() },
	privKeyType:   func(h host.Host, u *tptu.Upgrader) interface{} { return h.Peerstore().PrivKey(h.ID()) },
	pubKeyType:    func(h host.Host, u *tptu.Upgrader) interface{} { return h.Peerstore().PubKey(h.ID()) },
	pstoreType:    func(h host.Host, u *tptu.Upgrader) interface{} { return h.Peerstore() },
}

func newArgTypeSet(types ...reflect.Type) map[reflect.Type]constructor {
	result := make(map[reflect.Type]constructor, len(types))
	for _, ty := range types {
		c, ok := argTypes[ty]
		if !ok {
			panic(fmt.Sprintf("missing constructor for type %s", ty))
		}
		result[ty] = c
	}
	return result
}
