# pinger
ICMP/UDP ping library for Go
## Example
```
package main

import (
	"github.com/arvinkulagin/pinger"
	"log"
	"os"
	"fmt"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("You must specify remote host address")
		os.Exit(1)
	}
	raddr := os.Args[1]
	p, err := pinger.NewPinger("icmp", "0.0.0.0")
	if err != nil {
		log.Fatal(err)
	}
	for {
		time.Sleep(1 * time.Second)
		pong, err := p.Ping(raddr)
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Printf("%d Bytes from %s: icmp_seq=%d time=%f\n", pong.Size, pong.Peer.String(), pong.Seq, pong.RTT.Seconds())
	}
}
```
## API
```
import "github.com/arvinkulagin/pinger"

type Pinger interface {
    Ping(raddr string) (Pong, error)
    ResetCounter()
    SetTimeout(time.Duration)
}

func NewPinger(network, laddr string) (Pinger, error)
    Returns new Pinger interface. network argument must be "icmp" or "udp"
    (ICMP6 and UDP6 support will be added later).

type ICMP4Pinger struct {
    // contains filtered or unexported fields
}

func NewICMP4Pinger(laddr string) (*ICMP4Pinger, error)
    Returns new ICMP4Pinger. laddr is local ip address for listening.

func (i *ICMP4Pinger) Ping(raddr string) (Pong, error)
    Send Echo-Request to remote host and wait Echo-Reply. raddr is an
    address of remote host.

func (i *ICMP4Pinger) ResetCounter()
    Sets ICMP4Pinger counter to 0. Counter increments with each Ping() call.
    Counter value is set to Seq field in Echo-Request.

func (i *ICMP4Pinger) SetTimeout(d time.Duration)
    Sets ICMP4Pinger timeout. Timeout is a waiting time for a Echo-Reply
    from a remote host.

type UDP4Pinger struct {
    // contains filtered or unexported fields
}

func NewUDP4Pinger(laddr string) (*UDP4Pinger, error)
    Returns new UDP4Pinger. laddr is local ip address for listening.

func (i *UDP4Pinger) Ping(raddr string) (Pong, error)
    Send Echo-Request to remote host and wait Echo-Reply. raddr is an
    address of remote host.

func (i *UDP4Pinger) ResetCounter()
    Sets UDP4Pinger counter to 0. Counter increments with each Ping() call.
    Counter value is set to Seq field in Echo-Request.

func (i *UDP4Pinger) SetTimeout(d time.Duration)
    Sets UDP4Pinger timeout. Timeout is a waiting time for a Echo-Reply from
    a remote host.

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
```
