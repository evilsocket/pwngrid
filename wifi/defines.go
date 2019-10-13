package wifi

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
)

const (
	IDWhisperPayload      layers.Dot11InformationElementID = 222
	IDWhisperCompression  layers.Dot11InformationElementID = 223
	IDWhisperIdentity     layers.Dot11InformationElementID = 224
	IDWhisperSignature    layers.Dot11InformationElementID = 225
	IDWhisperStreamHeader layers.Dot11InformationElementID = 226
)

var SerializationOptions = gopacket.SerializeOptions{
	FixLengths:       true,
	ComputeChecksums: true,
}

var (
	SignatureAddr    = net.HardwareAddr{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad}
	SignatureAddrStr = "de:ad:be:ef:de:ad"
	BroadcastAddr    = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	wpaFlags         = 1041
)
