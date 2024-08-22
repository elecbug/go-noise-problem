package main

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
)

func main() {

	ctx := context.Background()

	CreateHostAndExchangeInfo(ctx)

	select {}
}

func CreateHostAndExchangeInfo(ctx context.Context) {
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Muxer(yamux.ID, yamux.DefaultTransport),
	)

	if err != nil {
		log.Fatalln(err)
	} else {
		addrs := make([]string, len(host.Addrs()))

		for i, v := range host.Addrs() {
			addrs[i] = v.String()
		}

		log.Println("Self:", host.Addrs(), host.ID())
	}

	if err != nil {
		log.Fatalln(err)
	}

	peerChan := InitMDNS(host, "mDNS")

	go func() {
		for {
			found := <-peerChan

			order := (bytes.Compare([]byte(host.ID()), []byte(found.ID)))
			if order == -1 {
				time.Sleep(100*time.Millisecond)
			}
			// existing code below
			err = host.Connect(ctx, found)

			// host.Peerstore().AddAddrs(found.ID, found.Addrs, 1*time.Hour)

			if err != nil {
				log.Println(err)
				continue
			}

			log.Println("Found:", found.String())
		}
	}()
}

type DiscoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

func (n *DiscoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

func InitMDNS(peerhost host.Host, rendezvous string) chan peer.AddrInfo {
	n := &DiscoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)

	ser := mdns.NewMdnsService(peerhost, rendezvous, n)

	err := ser.Start()

	if err != nil {
		log.Fatalln(err)
	}

	return n.PeerChan
}
