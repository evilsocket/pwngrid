package wifi

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func Unpack(pkt gopacket.Packet, radio *layers.RadioTap, dot11 *layers.Dot11) (error, []byte) {
	compressed := false
	payload := make([]byte, 0)

	for _, layer := range pkt.Layers() {
		if layer.LayerType() == layers.LayerTypeDot11InformationElement {
			if info, ok := layer.(*layers.Dot11InformationElement); ok {
				if info.ID == IDWhisperPayload {
					payload = append(payload, info.Info...)
				} else if info.ID == IDWhisperCompression {
					compressed = true
				}
			}
		}
	}

	if compressed {
		if decompressed, err := Decompress(payload); err != nil {
			return fmt.Errorf("error decompressing payload: %v", err), nil
		} else {
			payload = decompressed
		}
	}

	return nil, payload
}
