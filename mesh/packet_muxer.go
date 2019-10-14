package mesh

import (
	"fmt"
	"github.com/evilsocket/islazy/async"
	"github.com/evilsocket/islazy/log"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"strings"
	"time"
)

const (
	// Ugly, but gopacket folks are not exporting pcap errors, so ...
	// ref. https://github.com/google/gopacket/blob/96986c90e3e5c7e01deed713ff8058e357c0c047/pcap/pcap.go#L281
	ErrIfaceNotUp = "Interface Not Up"
)

var (
	SnapLength  = 65536
	ReadTimeout = 100
)

type PacketCallback func(pkt gopacket.Packet)

type PacketMuxer struct {
	iface   string
	filter  string
	handle  *pcap.Handle
	source  *gopacket.PacketSource
	channel chan gopacket.Packet
	queue   *async.WorkQueue
	stop    chan struct{}

	onPacket PacketCallback
}

func dummyPacketCallback(pkt gopacket.Packet) {

}

func NewPacketMuxer(iface, filter string, workers int) (mux *PacketMuxer, err error) {
	mux = &PacketMuxer{
		iface:    iface,
		filter:   filter,
		stop:     make(chan struct{}),
		onPacket: dummyPacketCallback,
	}

	for retry := 0; ; retry++ {
		inactiveHandle, err := pcap.NewInactiveHandle(iface)
		if err != nil {
			return nil, fmt.Errorf("error while opening interface %s: %s", iface, err)
		}
		defer inactiveHandle.CleanUp()

		if err = inactiveHandle.SetRFMon(true); err != nil {
			log.Warning("error while setting interface %s in monitor mode: %s", iface, err)
		}

		if err = inactiveHandle.SetSnapLen(SnapLength); err != nil {
			return nil, fmt.Errorf("error while settng span len: %s", err)
		}
		/*
		 * We don't want to pcap.BlockForever otherwise pcap_close(handle)
		 * could hang waiting for a timeout to expire ...
		 */
		readTimeout := time.Duration(ReadTimeout) * time.Millisecond
		if err = inactiveHandle.SetTimeout(readTimeout); err != nil {
			return nil, fmt.Errorf("error while setting timeout: %s", err)
		} else if mux.handle, err = inactiveHandle.Activate(); err != nil {
			if retry == 0 && err.Error() == ErrIfaceNotUp {
				log.Info("interface %s is down, bringing it up ...", iface)
				if err := ActivateInterface(iface); err != nil {
					return nil, err
				}
				continue
			}
			return nil, fmt.Errorf("error while activating handle: %s", err)
		}

		if filter != "" {
			if err := mux.handle.SetBPFFilter(filter); err != nil {
				return nil, fmt.Errorf("error setting BPF filter '%s': %v", filter, err)
			}
		}

		break
	}

	mux.source = gopacket.NewPacketSource(mux.handle, mux.handle.LinkType())
	mux.channel = mux.source.Packets()
	mux.queue = async.NewQueue(workers, func(arg async.Job) {
		mux.onPacket(arg.(gopacket.Packet))
	})

	return mux, nil
}

func (mux *PacketMuxer) OnPacket(cb PacketCallback) {
	mux.onPacket = cb
}

func (mux *PacketMuxer) Write(data []byte) error {
	var err error
	for attempt := 0; attempt < 5; attempt++ {
		if err = mux.handle.WritePacketData(data); err == nil {
			return nil
		} else if strings.Contains(err.Error(), "temporarily unavailable") {
			log.Debug("resource temporarily unavailable when sending data")
			// if it's the last attempt this will set err to nil as we can't really
			// do a lot about this case, otherwise it'll wait 200ms before the next
			// attempt is made.
			err = nil
			if attempt < 5 {
				time.Sleep(200 * time.Millisecond)
			}
		} else {
			return nil
		}
	}
	return err
}

func (mux *PacketMuxer) Start() {
	go func() {
		log.Debug("packet muxer started (iface:%s filter:%s)", mux.iface, mux.filter)
		for {
			select {
			case packet := <-mux.channel:
				mux.queue.Add(async.Job(packet))
			case <-mux.stop:
				return
			}
		}
	}()
}

func (mux *PacketMuxer) Stop() {
	log.Debug("stopping packet muxer ...")
	mux.stop <- struct{}{}
	mux.queue.WaitDone()
	log.Debug("packet muxer stopped")
}
