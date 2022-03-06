// Package dosbox implements a client for connecting to a DOSbox IPX server
// over UDP.
package dosbox

import (
	"context"
	"errors"
	"io"

	udpclient "github.com/fragglet/ipxbox/client"
	"github.com/fragglet/ipxbox/ipx"
	"github.com/fragglet/ipxbox/network"
	"github.com/fragglet/ipxbox/network/pipe"
)

var (
	_ = (network.Node)(&client{})
)

type client struct {
	inner  ipx.ReadWriteCloser
	rxpipe ipx.ReadWriteCloser
	addr   ipx.Addr
}

func (c *client) ReadPacket(ctx context.Context) (*ipx.Packet, error) {
	return c.rxpipe.ReadPacket(ctx)
}

func (c *client) WritePacket(packet *ipx.Packet) error {
	return c.inner.WritePacket(packet)
}

func (c *client) Close() error {
	c.rxpipe.Close()
	return c.inner.Close()
}

func (c *client) GetProperty(x interface{}) bool {
	switch x.(type) {
	case *ipx.Addr:
		*x.(*ipx.Addr) = c.addr
		return true
	default:
		return false
	}
}

func (c *client) recvLoop(ctx context.Context) {
	for {
		packet, err := c.inner.ReadPacket(ctx)
		if errors.Is(err, io.ErrClosedPipe) {
			break
		} else if err != nil {
			// TODO: Log error?
			continue
		}

		// TODO: Send ping responses
		c.rxpipe.WritePacket(packet)
	}
}

func Dial(addr string) (network.Node, error) {
	udp, err := udpclient.Dial(addr)
	if err != nil {
		return nil, err
	}
	// TODO: Do connection handshake, obtain address.
	c := &client{
		inner:  udp,
		rxpipe: pipe.New(1),
	}
	go c.recvLoop(context.Background())
	return c, nil
}
