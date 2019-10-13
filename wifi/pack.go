package wifi

import (
	"bytes"
	"encoding/binary"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
)

func Info(id layers.Dot11InformationElementID, info []byte) *layers.Dot11InformationElement {
	return &layers.Dot11InformationElement{
		ID:     id,
		Length: uint8(len(info) & 0xff),
		Info:   info,
	}
}

func PackOneOf(from, to net.HardwareAddr, peerID []byte, signature []byte, streamID uint64, seqNum uint64, seqTot uint64, payload []byte, compress bool) (error, []byte) {
	stack := []gopacket.SerializableLayer{
		&layers.RadioTap{},
		&layers.Dot11{
			Address1: to,
			Address2: SignatureAddr,
			Address3: from,
			Type:     layers.Dot11TypeMgmtBeacon,
		},
		&layers.Dot11MgmtBeacon{
			Flags:    uint16(wpaFlags),
			Interval: 100,
		},
	}

	if peerID != nil {
		stack = append(stack, Info(IDWhisperIdentity, peerID))
	}

	if signature != nil {
		stack = append(stack, Info(IDWhisperSignature, signature))
	}

	if streamID > 0 {
		streamBuf := new(bytes.Buffer)
		if err := binary.Write(streamBuf, binary.LittleEndian, streamID); err != nil {
			return err, nil
		} else if err = binary.Write(streamBuf, binary.LittleEndian, seqNum); err != nil {
			return err, nil
		} else if err = binary.Write(streamBuf, binary.LittleEndian, seqTot); err != nil {
			return err, nil
		}
		stack = append(stack, Info(IDWhisperStreamHeader, streamBuf.Bytes()))
	}

	if compress {
		if didCompress, compressed, err := Compress(payload); err != nil {
			return err, nil
		} else if didCompress {
			stack = append(stack, Info(IDWhisperCompression, []byte{1}))
			payload = compressed
		}
	}

	dataSize := len(payload)
	dataLeft := dataSize
	dataOff := 0
	chunkSize := 0xff

	for dataLeft > 0 {
		sz := chunkSize
		if dataLeft < chunkSize {
			sz = dataLeft
		}

		chunk := payload[dataOff : dataOff+sz]
		stack = append(stack, Info(IDWhisperPayload, chunk))

		dataOff += sz
		dataLeft -= sz
	}

	return Serialize(stack...)
}

func Pack(from, to net.HardwareAddr, payload []byte, compress bool) (error, []byte) {
	return PackOneOf(from, to, nil, nil, 0, 0, 0, payload, compress)
}
