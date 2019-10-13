package wifi

import (
	"bytes"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func Serialize(layers ...gopacket.SerializableLayer) (error, []byte) {
	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buf, SerializationOptions, layers...); err != nil {
		return err, nil
	}
	return nil, buf.Bytes()
}

func IsBroadcast(dot11 *layers.Dot11) bool {
	return bytes.Equal(dot11.Address1, BroadcastAddr)
}

func Freq2Chan(freq int) int {
	if freq <= 2472 {
		return ((freq - 2412) / 5) + 1
	} else if freq == 2484 {
		return 14
	} else if freq >= 5035 && freq <= 5865 {
		return ((freq - 5035) / 5) + 7
	}
	return 0
}

func Chan2Freq(channel int) int {
	if channel <= 13 {
		return ((channel - 1) * 5) + 2412
	} else if channel == 14 {
		return 2484
	} else if channel <= 173 {
		return ((channel - 7) * 5) + 5035
	}

	return 0
}
