package libp2p

// This file contains all the default configuration options.

import (
	"crypto/rand"

	secio "gx/ipfs/QmQS7P1VV4JuqbXEEaPQZ5ERHDQuGp8qk26Gfdg9ZtP1Eb/go-libp2p-secio"
	ws "gx/ipfs/QmTeRSFgnXRCh13sxsZkLTVCc1diUbZiT5mkGUgkR1J1on/go-ws-transport"
	tcp "gx/ipfs/QmU7NiV3ZRJoNk8XZEKazs4AW7XQT1rQCrhh8cmhSjZgrC/go-tcp-transport"
	pstore "gx/ipfs/QmZb7hAgQEhW9dBbzBudU39gCeD4zbe6xafD52LUuF4cUN/go-libp2p-peerstore"
	mplex "gx/ipfs/QmZeGmoJ3bEwEe6Huz6GKcHENWZCx7DReuAS5va4zP24PB/go-smux-multiplex"
	yamux "gx/ipfs/QmcsgrV3nCAKjiHKZhKVXWc4oY3WBECJCqahXEMpHeMrev/go-smux-yamux"
	crypto "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

// DefaultSecurity is the default security option.
//
// Useful when you want to extend, but not replace, the supported transport
// security protocols.
var DefaultSecurity = Security(secio.ID, secio.New)

// DefaultMuxer configures libp2p to use the stream connection multiplexers.
//
// Use this option when you want to *extend* the set of multiplexers used by
// libp2p instead of replacing them.
var DefaultMuxers = ChainOptions(
	Muxer("/yamux/1.0.0", yamux.DefaultTransport),
	Muxer("/mplex/6.3.0", mplex.DefaultTransport),
)

// DefaultTransports are the default libp2p transports.
//
// Use this option when you want to *extend* the set of multiplexers used by
// libp2p instead of replacing them.
var DefaultTransports = ChainOptions(
	Transport(tcp.NewTCPTransport),
	Transport(ws.New),
)

// DefaultPeerstore configures libp2p to use the default peerstore.
var DefaultPeerstore Option = func(cfg *Config) error {
	return cfg.Apply(Peerstore(pstore.NewPeerstore()))
}

// RandomIdentity generates a random identity (default behaviour)
var RandomIdentity = func(cfg *Config) error {
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return err
	}
	return cfg.Apply(Identity(priv))
}

// Complete list of default options and when to fallback on them.
//
// Please *DON'T* specify default options any other way. Putting this all here
// makes tracking defaults *much* easier.
var defaults = []struct {
	fallback func(cfg *Config) bool
	opt      Option
}{
	{
		fallback: func(cfg *Config) bool { return cfg.Transports == nil },
		opt:      DefaultTransports,
	},
	{
		fallback: func(cfg *Config) bool { return cfg.Muxers == nil },
		opt:      DefaultMuxers,
	},
	{
		fallback: func(cfg *Config) bool { return !cfg.Insecure && cfg.SecurityTransports == nil },
		opt:      DefaultSecurity,
	},
	{
		fallback: func(cfg *Config) bool { return cfg.PeerKey == nil },
		opt:      RandomIdentity,
	},
	{
		fallback: func(cfg *Config) bool { return cfg.Peerstore == nil },
		opt:      DefaultPeerstore,
	},
}

// Defaults configures libp2p to use the default options. Can be combined with
// other options to *extend* the default options.
var Defaults Option = func(cfg *Config) error {
	for _, def := range defaults {
		if err := cfg.Apply(def.opt); err != nil {
			return err
		}
	}
	return nil
}

// FallbackDefaults applies default options to the libp2p node if and only if no
// other relevent options have been applied. will be appended to the options
// passed into New.
var FallbackDefaults Option = func(cfg *Config) error {
	for _, def := range defaults {
		if !def.fallback(cfg) {
			continue
		}
		if err := cfg.Apply(def.opt); err != nil {
			return err
		}
	}
	return nil
}
