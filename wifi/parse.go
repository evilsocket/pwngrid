package wifi

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func Parse(packet gopacket.Packet) (ok bool, radio *layers.RadioTap, dot11 *layers.Dot11) {
	ok = false
	radio = nil
	dot11 = nil

	radioLayer := packet.Layer(layers.LayerTypeRadioTap)
	if radioLayer == nil {
		return
	}
	radio, ok = radioLayer.(*layers.RadioTap)
	if !ok || radio == nil {
		return
	}

	dot11Layer := packet.Layer(layers.LayerTypeDot11)
	if dot11Layer == nil {
		ok = false
		return
	}

	dot11, ok = dot11Layer.(*layers.Dot11)
	return
}
