package mockrouting

import (
	"context"
	"errors"
	"time"

	u "gx/ipfs/QmNiJuT8Ja3hMVpBHXv3Q6dwmperaQ6JjLtpMQgMCD7xvx/go-ipfs-util"
	logging "gx/ipfs/QmTG23dvpBCBjqQwyDxV8CQT6jmS4PSftNr1VqHhE3MLy7/go-log"
	dhtpb "gx/ipfs/QmTUyK82BVPA6LmSzEJpfEunk9uBaQzWtMsNP917tVj4sT/go-libp2p-record/pb"
	"gx/ipfs/QmUJzxQQ2kzwQubsMqBTr1NGDpLfh7pGA2E1oaJULcKDPq/go-testutil"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	routing "gx/ipfs/QmXijJ3T9MjB2v8xpFDoEX6FqR9u8PkJkzu49TgwJ8Ndr5/go-libp2p-routing"
	ropts "gx/ipfs/QmXijJ3T9MjB2v8xpFDoEX6FqR9u8PkJkzu49TgwJ8Ndr5/go-libp2p-routing/options"
	dshelp "gx/ipfs/QmYJgz1Z5PbBGP7n2XA8uv5sF1EKLfYUjL7kFemVAjMNqC/go-ipfs-ds-help"
	proto "gx/ipfs/QmZ4Qi3GaRbjcx28Sme5eMH7RQjGkt8wHxt2a65oLaeFEV/gogo-protobuf/proto"
	pstore "gx/ipfs/QmZb7hAgQEhW9dBbzBudU39gCeD4zbe6xafD52LUuF4cUN/go-libp2p-peerstore"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	cid "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"
	ds "gx/ipfs/QmeiCcJfDW1GJnWUArudsv5rQsihpi4oyddPhdqo3CfX6i/go-datastore"
)

var log = logging.Logger("mockrouter")

type client struct {
	datastore ds.Datastore
	server    server
	peer      testutil.Identity
}

// FIXME(brian): is this method meant to simulate putting a value into the network?
func (c *client) PutValue(ctx context.Context, key string, val []byte, opts ...ropts.Option) error {
	log.Debugf("PutValue: %s", key)
	rec := new(dhtpb.Record)
	rec.Value = val
	rec.Key = proto.String(string(key))
	rec.TimeReceived = proto.String(u.FormatRFC3339(time.Now()))
	data, err := proto.Marshal(rec)
	if err != nil {
		return err
	}

	return c.datastore.Put(dshelp.NewKeyFromBinary([]byte(key)), data)
}

// FIXME(brian): is this method meant to simulate getting a value from the network?
func (c *client) GetValue(ctx context.Context, key string, opts ...ropts.Option) ([]byte, error) {
	log.Debugf("GetValue: %s", key)
	v, err := c.datastore.Get(dshelp.NewKeyFromBinary([]byte(key)))
	if err != nil {
		return nil, err
	}

	data, ok := v.([]byte)
	if !ok {
		return nil, errors.New("could not cast value from datastore")
	}

	rec := new(dhtpb.Record)
	err = proto.Unmarshal(data, rec)
	if err != nil {
		return nil, err
	}

	return rec.GetValue(), nil
}

func (c *client) FindProviders(ctx context.Context, key *cid.Cid) ([]pstore.PeerInfo, error) {
	return c.server.Providers(key), nil
}

func (c *client) FindPeer(ctx context.Context, pid peer.ID) (pstore.PeerInfo, error) {
	log.Debugf("FindPeer: %s", pid)
	return pstore.PeerInfo{}, nil
}

func (c *client) FindProvidersAsync(ctx context.Context, k *cid.Cid, max int) <-chan pstore.PeerInfo {
	out := make(chan pstore.PeerInfo)
	go func() {
		defer close(out)
		for i, p := range c.server.Providers(k) {
			if max <= i {
				return
			}
			select {
			case out <- p:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Provide returns once the message is on the network. Value is not necessarily
// visible yet.
func (c *client) Provide(_ context.Context, key *cid.Cid, brd bool) error {
	if !brd {
		return nil
	}
	info := pstore.PeerInfo{
		ID:    c.peer.ID(),
		Addrs: []ma.Multiaddr{c.peer.Address()},
	}
	return c.server.Announce(info, key)
}

func (c *client) Ping(ctx context.Context, p peer.ID) (time.Duration, error) {
	return 0, nil
}

func (c *client) Bootstrap(context.Context) error {
	return nil
}

var _ routing.IpfsRouting = &client{}
