package outbound

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/metacubex/mihomo/common/pool"
	"github.com/metacubex/mihomo/component/dialer"
	"github.com/metacubex/mihomo/component/resolver"
	C "github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/log"
)

type Dns struct {
	*Base
}

type DnsOption struct {
	BasicOption
	Name string `proxy:"name"`
}

// DialContext implements C.ProxyAdapter
func (d *Dns) DialContext(ctx context.Context, metadata *C.Metadata, opts ...dialer.Option) (C.Conn, error) {
	return nil, fmt.Errorf("dns outbound does not support tcp")
}

// ListenPacketContext implements C.ProxyAdapter
func (d *Dns) ListenPacketContext(ctx context.Context, metadata *C.Metadata, opts ...dialer.Option) (C.PacketConn, error) {
	log.Debugln("[DNS] hijack udp:%s from %s", metadata.RemoteAddress(), metadata.SourceAddrPort())

	ctx, cancel := context.WithCancel(context.Background())

	return newPacketConn(&dnsPacketConn{
		response: make(chan dnsPacket, 1),
		ctx:      ctx,
		cancel:   cancel,
	}, d), nil
}

type dnsPacket struct {
	data []byte
	put  func()
	addr net.Addr
}

// dnsPacketConn implements net.PacketConn
type dnsPacketConn struct {
	response chan dnsPacket
	ctx      context.Context
	cancel   context.CancelFunc
}

func (d *dnsPacketConn) WaitReadFrom() (data []byte, put func(), addr net.Addr, err error) {
	select {
	case packet := <-d.response:
		return packet.data, packet.put, packet.addr, nil
	case <-d.ctx.Done():
		return nil, nil, nil, net.ErrClosed
	}
}

func (d *dnsPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	select {
	case packet := <-d.response:
		n = copy(p, packet.data)
		if packet.put != nil {
			packet.put()
		}
		return n, packet.addr, nil
	case <-d.ctx.Done():
		return 0, nil, net.ErrClosed
	}
}

func (d *dnsPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	ctx, cancel := context.WithTimeout(d.ctx, resolver.DefaultDnsRelayTimeout)
	defer cancel()

	buf := pool.Get(resolver.SafeDnsPacketSize)
	put := func() { _ = pool.Put(buf) }
	buf, err = resolver.RelayDnsPacket(ctx, p, buf)
	if err != nil {
		put()
		return 0, err
	}

	packet := dnsPacket{
		data: buf,
		put:  put,
		addr: addr,
	}
	select {
	case d.response <- packet:
		return len(p), nil
	case <-d.ctx.Done():
		put()
		return 0, net.ErrClosed
	}
}

func (d *dnsPacketConn) Close() error {
	d.cancel()
	return nil
}

func (*dnsPacketConn) LocalAddr() net.Addr {
	return &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 53,
		Zone: "",
	}
}

func (*dnsPacketConn) SetDeadline(t time.Time) error {
	return nil
}

func (*dnsPacketConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (*dnsPacketConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func NewDnsWithOption(option DnsOption) *Dns {
	return &Dns{
		Base: &Base{
			name:   option.Name,
			tp:     C.Dns,
			udp:    true,
			tfo:    option.TFO,
			mpTcp:  option.MPTCP,
			iface:  option.Interface,
			rmark:  option.RoutingMark,
			prefer: C.NewDNSPrefer(option.IPVersion),
		},
	}
}