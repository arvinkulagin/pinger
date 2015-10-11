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