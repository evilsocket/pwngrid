package mesh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/wifi"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"sync"
	"time"
)

var (
	Workers = 0
	PeerTTL = 1800
	Peers   = sync.Map{}
)

func dummyPeerActivityCallback(ident string, peer *Peer) {}

type PeerActivityCallback func(ident string, peer *Peer)

type Router struct {
	local      *Peer
	mux        *PacketMuxer
	onNewPeer  PeerActivityCallback
	onPeerLost PeerActivityCallback
	memory     *Storage
}

func StartRouting(iface string, peersPath string, local *Peer) (*Router, error) {
	err, memory := StorageFromPath(peersPath)
	if err != nil {
		return nil, err
	}

	filter := fmt.Sprintf("type mgt subtype beacon and ether src %s", wifi.SignatureAddrStr)
	mux, err := NewPacketMuxer(iface, filter, Workers)
	if err != nil {
		return nil, err
	}

	router := &Router{
		mux:        mux,
		local:      local,
		memory:     memory,
		onNewPeer:  dummyPeerActivityCallback,
		onPeerLost: dummyPeerActivityCallback,
	}
	mux.OnPacket(router.onPacket)
	mux.Start()

	log.Info("started beacon discovery and message routing (%d known peers)", router.memory.Size())

	go router.peersPruner()

	return router, nil
}

func (router *Router) Memory() []*jsonPeer {
	return router.memory.List()
}

func (router *Router) OnNewPeer(cb PeerActivityCallback) {
	router.onNewPeer = cb
}

func (router *Router) OnPeerLost(cb PeerActivityCallback) {
	router.onPeerLost = cb
}

func (router *Router) peersPruner() {
	period := time.Duration(500) * time.Millisecond
	tick := time.NewTicker(period)

	log.Debug("peers pruner started with a %s period", period)

	for _ = range tick.C {
		stale := map[string]*Peer{}

		Peers.Range(func(key, value interface{}) bool {
			ident := key.(string)
			peer := value.(*Peer)
			inactive := peer.InactiveFor()
			if int(inactive) > PeerTTL {
				stale[ident] = peer
			}
			return true
		})

		for ident, peer := range stale {
			Peers.Delete(ident)
			router.onPeerLost(ident, peer)
		}
	}
}

func (router *Router) newPeer(ident string, peer *Peer) {
	Peers.Store(ident, peer)
	if err := router.memory.Track(ident, peer); err != nil {
		log.Error("error saving peer encounter for %s: %v", ident, err)
	}
	router.onNewPeer(ident, peer)
}

func (router *Router) onPeerAdvertisement(pkt gopacket.Packet, radio *layers.RadioTap, dot11 *layers.Dot11) {
	err, payload := wifi.Unpack(pkt, radio, dot11)
	if err != nil {
		log.Debug("%v", err)
		return
	}

	advData := make(map[string]interface{})
	if err := json.Unmarshal(payload, &advData); err != nil {
		log.Debug("error decoding payload '%s': %v", payload, err)
		return
	}

	ident, ok := advData["identity"]
	if !ok {
		log.Debug("error parsing identity from payload '%s'", payload)
		return
	}

	if _peer, existing := Peers.Load(ident); existing {
		peer := _peer.(*Peer)
		if err := peer.Update(radio, dot11, advData); err != nil {
			log.Warning("error updating peer %s: %v", peer.ID(), err)
		}
	} else if peer, err := NewPeer(radio, dot11, advData); err != nil {
		log.Debug("error creating peer: %v", err)
		return
	} else {
		router.newPeer(ident.(string), peer)
	}
}

func (router *Router) onPacket(pkt gopacket.Packet) {
	if ok, radio, dot11 := wifi.Parse(pkt); ok && dot11.ChecksumValid() {
		src := dot11.Address3
		dst := dot11.Address1
		if !bytes.Equal(src, router.local.SessionID) {
			if bytes.Equal(dst, wifi.BroadcastAddr) {
				router.onPeerAdvertisement(pkt, radio, dot11)
			} else {
				// log.Debug("ignoring message %x > %x", src, dst)
			}
		}
	}
}
