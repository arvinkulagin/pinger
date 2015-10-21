package pinger

import (
	"math/rand"
	"time"
	"net"
	"errors"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	protocolICMP = 1
)

type Pinger interface {
	Ping(raddr string) (Pong, error)
	ResetCounter()
	SetTimeout(time.Duration)
}

// Returns new Pinger interface. network argument must be "icmp" or "udp" (ICMP6 and UDP6 support will be added later).
func NewPinger(network, laddr string) (Pinger, error) {
	switch network {
		case "icmp": return NewICMP4Pinger(laddr)
		case "udp": return NewUDP4Pinger(laddr)
		default: return &ICMP4Pinger{}, errors.New("Unknown network " + network)
	}
}

type ICMP4Pinger struct {
	laddr net.Addr
	id int
	counter int
	timeout time.Duration
}

// Returns new ICMP4Pinger. laddr is local ip address for listening.
func NewICMP4Pinger(laddr string) (*ICMP4Pinger, error) {
	addr, err := net.ResolveIPAddr("ip4", laddr)
	if err != nil {
		return nil, err
	}
	pinger := ICMP4Pinger{
		laddr: addr,
		id: rand.Int() & 0xffff,
		counter: 0,
		timeout: 2 * time.Second,
	}
	return &pinger, nil
}

// Sets ICMP4Pinger counter to 0. Counter increments with each Ping() call.
// Counter value is set to Seq field in Echo-Request.
func (i *ICMP4Pinger) ResetCounter() {
	i.counter = 0
}

// Sets ICMP4Pinger timeout. Timeout is a waiting time for a Echo-Reply from a remote host.
func (i *ICMP4Pinger) SetTimeout(d time.Duration) {
	i.timeout = d
}

// Send Echo-Request to remote host and wait Echo-Reply.
// raddr is an address of remote host.
func (i *ICMP4Pinger) Ping(raddr string) (Pong, error) {
	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   i.id,
			Seq:  i.counter,
		},
	}
	i.counter++
	listener, err := icmp.ListenPacket("ip4:icmp", i.laddr.String())
	if err != nil {
		return Pong{}, err
	}
	defer listener.Close()
	addr, err := net.ResolveIPAddr("ip4", raddr)
	if err != nil {
		return Pong{}, err
	}
	return ping(listener, message, addr, i.timeout)
}

type UDP4Pinger struct {
	laddr net.Addr
	id int
	counter int
	timeout time.Duration
}

// Returns new UDP4Pinger. laddr is local ip address for listening.
func NewUDP4Pinger(laddr string) (*UDP4Pinger, error) {
	addr, err := net.ResolveIPAddr("ip4", laddr)
	if err != nil {
		return nil, err
	}
	pinger := UDP4Pinger{
		laddr: addr,
		id: rand.Int() & 0xffff,
		counter: 0,
		timeout: 2 * time.Second,
	}
	return &pinger, nil
}

// Sets UDP4Pinger counter to 0. Counter increments with each Ping() call.
// Counter value is set to Seq field in Echo-Request.
func (i *UDP4Pinger) ResetCounter() {
	i.counter = 0
}

// Sets UDP4Pinger timeout. Timeout is a waiting time for a Echo-Reply from a remote host.
func (i *UDP4Pinger) SetTimeout(d time.Duration) {
	i.timeout = d
}

// Send Echo-Request to remote host and wait Echo-Reply.
// raddr is an address of remote host.
func (i *UDP4Pinger) Ping(raddr string) (Pong, error) {
	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   i.id,
			Seq:  i.counter,
		},
	}
	i.counter++
	listener, err := icmp.ListenPacket("udp4", i.laddr.String())
	if err != nil {
		return Pong{}, err
	}
	defer listener.Close()
	addr, err := net.ResolveIPAddr("ip4", raddr)
	if err != nil {
		return Pong{}, err
	}
	return ping(listener, message, &net.UDPAddr{IP: net.ParseIP(addr.String())}, i.timeout)
}

func ping(listener *icmp.PacketConn, message icmp.Message, raddr net.Addr, timeout time.Duration) (Pong, error) {
	data, err := message.Marshal(nil)
	if err != nil {
		return Pong{}, err
	}
	_, err = listener.WriteTo(data, raddr)
	if err != nil {
		return Pong{}, err
	}
	now := time.Now()
	done := make(chan Pong)
	go func() {
		for {
			buf := make([]byte, 10000)
			// bufio
			n, peer, err := listener.ReadFrom(buf)
			if err != nil {
				return
			}
			since := time.Since(now)
			input, err := icmp.ParseMessage(protocolICMP, buf[:n])
			if err != nil {
				return
			}
			if input.Type != ipv4.ICMPTypeEchoReply {
				continue
			}
			echo := input.Body.(*icmp.Echo)
			pong := Pong{
				Peer: peer,
				ID: echo.ID,
				Seq: echo.Seq,
				Data: echo.Data,
				Size: n,
				RTT: since,
			}
			done <- pong
			return
		}
	}()
	select {
	case pong := <-done:
		return pong, nil
	case <-time.After(timeout):
		return Pong{}, errors.New("Timeout")
	}
}

type Pong struct {
	// IP address of pinged host.
	Peer net.Addr
	// ICMP ID.
	ID int
	// ICMP sequence number.
	Seq int
	// Content of ICMP data field.
	Data []byte
	// Size of ICMP Echo-Reply.
	Size int
	// Round-trip time.
	RTT time.Duration
}