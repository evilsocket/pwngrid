package mesh

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/crypto"
	"github.com/evilsocket/pwngrid/version"
	"github.com/evilsocket/pwngrid/wifi"
	"github.com/google/gopacket/layers"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	SignalingPeriod = 300
)

type SessionID []byte

type Peer struct {
	sync.Mutex

	DetectedAt   time.Time
	SeenAt       time.Time
	Channel      int
	RSSI         int
	SessionID    SessionID
	SessionIDStr string
	Keys         *crypto.KeyPair
	AdvData      sync.Map
	AdvPeriod    int

	advEnabled bool
	mux        *PacketMuxer
	stop       chan struct{}
}

type jsonPeer struct {
	DetectedAt    time.Time              `json:"detected_at"`
	SeenAt        time.Time              `json:"seen_at"`
	Channel       int                    `json:"channel"`
	RSSI          int                    `json:"rssi"`
	SessionID     string                 `json:"session_id"`
	Advertisement map[string]interface{} `json:"advertisement"`
}

func MakeLocalPeer(name string, keys *crypto.KeyPair) *Peer {
	now := time.Now()
	peer := &Peer{
		DetectedAt: now,
		SeenAt:     now,
		SessionID:  make([]byte, 6),
		Keys:       keys,
		AdvData:    sync.Map{},
		AdvPeriod:  SignalingPeriod,
		stop:       make(chan struct{}),
		advEnabled: false,
	}

	if _, err := rand.Read(peer.SessionID); err != nil {
		panic(err)
	}

	parts := make([]string, 6)
	for idx, byte := range peer.SessionID {
		parts[idx] = fmt.Sprintf("%02x", byte)
	}
	peer.SessionIDStr = strings.Join(parts, ":")

	peer.AdvData.Store("name", name)
	peer.AdvData.Store("identity", keys.FingerprintHex)
	peer.AdvData.Store("session_id", peer.SessionIDStr)
	peer.AdvData.Store("grid_version", version.Version)

	peer.AdvData.Range(func(key, value interface{}) bool {
		log.Debug("local.adv.%s = %s", key, value)
		return true
	})

	return peer
}

func (peer *Peer) Advertise(enabled bool) {
	peer.Lock()
	defer peer.Unlock()
	diff := peer.advEnabled != enabled
	peer.advEnabled = enabled
	if diff {
		if enabled {
			log.Info("peer advertisement enabled")
		} else {
			log.Info("peer advertisement disabled")
		}
	}
}

func NewPeer(radiotap *layers.RadioTap, dot11 *layers.Dot11, adv map[string]interface{}) (peer *Peer, err error) {
	now := time.Now()
	peer = &Peer{
		DetectedAt: now,
		SeenAt:     now,
		Channel:    wifi.Freq2Chan(int(radiotap.ChannelFrequency)),
		RSSI:       int(radiotap.DBMAntennaSignal),
		SessionID:  SessionID(dot11.Address3),
		AdvData:    sync.Map{},
	}

	parts := make([]string, 6)
	for idx, byte := range peer.SessionID {
		parts[idx] = fmt.Sprintf("%02x", byte)
	}
	peer.SessionIDStr = strings.Join(parts, ":")

	// parse the fingerprint, the signature and the public key
	fingerprint, found := adv["identity"].(string)
	if !found {
		return nil, fmt.Errorf("peer %x is not advertising any identity", peer.SessionID)
	}

	if pubKey64, found := adv["public_key"]; found {
		pubKey, err := base64.StdEncoding.DecodeString(pubKey64.(string))
		if err != nil {
			return nil, fmt.Errorf("error decoding peer %s public key: %s", fingerprint, err)
		}

		peer.Keys, err = crypto.FromPublicPEM(string(pubKey))
		if err != nil {
			return nil, fmt.Errorf("error parsing peer %s public key: %s", fingerprint, err)
		}

		// basic consistency check
		if peer.Keys.FingerprintHex != fingerprint {
			return nil, fmt.Errorf("peer %x is advertising fingerprint %s, but it should be %s", peer.SessionID, fingerprint, peer.Keys.FingerprintHex)
		}
	} else if !found {
		log.Debug("peer %s is not advertising any public key", fingerprint)
	}

	/*
		No need for signature in the advertisement protocol, however:

		signature64, found := adv["signature"].(string)
		if !found {
			return nil, fmt.Errorf("peer %s is advertising unsigned data", fingerprint)
		}

		signature, err := base64.StdEncoding.DecodeString(signature64)
		if err != nil {
			return nil, fmt.Errorf("error decoding peer %s signature: %s", fingerprint, err)
		}

		// the signature is SIGN(advertisement), so we need to remove the signature field and convert back to json.
		// NOTE: fortunately, keys will be always sorted, so we don't have to do anything in order to guarantee signature
		// consistency (https://stackoverflow.com/questions/18668652/how-to-produce-json-with-sorted-keys-in-go)
		signedMap := adv
		delete(signedMap, "signature")

		signedData, err := json.Marshal(signedMap)
		if err != nil {
			return nil, fmt.Errorf("error packing data for signature verification: %v", err)
		}

		// verify the signature
		if err = peer.Keys.VerifyMessage(signedData, signature); err != nil {
			return nil, fmt.Errorf("peer %x signature is invalid", peer.SessionID)
		}
	 */

	for key, value := range adv {
		peer.AdvData.Store(key, value)
	}

	return peer, nil
}

func (peer *Peer) MarshalJSON() ([]byte, error) {
	peer.Lock()
	defer peer.Unlock()

	doc := jsonPeer{
		DetectedAt:    peer.DetectedAt,
		SeenAt:        peer.SeenAt,
		Channel:       peer.Channel,
		RSSI:          peer.RSSI,
		SessionID:     peer.SessionIDStr,
		Advertisement: make(map[string]interface{}),
	}
	peer.AdvData.Range(func(key, value interface{}) bool {
		doc.Advertisement[key.(string)] = value
		return true
	})
	return json.Marshal(doc)
}

func (peer *Peer) Update(radio *layers.RadioTap, dot11 *layers.Dot11, adv map[string]interface{}) (err error) {
	peer.Lock()
	defer peer.Unlock()

	// parse the fingerprint, the signature and the public key
	fingerprint, found := adv["identity"].(string)
	if !found {
		return fmt.Errorf("peer %x is not advertising any identity", peer.SessionID)
	}

	// basic consistency check
	if peer.Keys != nil && peer.Keys.FingerprintHex != fingerprint {
		return fmt.Errorf("peer %x is advertising fingerprint %s, but it should be %s", peer.SessionID, fingerprint, peer.Keys.FingerprintHex)
	}

	/*
		No need for signature in the advertisement protocol, however:

		signature64, found := adv["signature"].(string)
		if !found {
			return fmt.Errorf("peer %x is not advertising any signature", peer.SessionID)
		}

		signature, err := base64.StdEncoding.DecodeString(signature64)
		if err != nil {
			return fmt.Errorf("error decoding peer %d signature: %s", peer.SessionID, err)
		}

		// the signature is SIGN(advertisement), so we need to remove the signature field and convert back to json.
		// NOTE: fortunately, keys will be always sorted, so we don't have to do anything in order to guarantee signature
		// consistency (https://stackoverflow.com/questions/18668652/how-to-produce-json-with-sorted-keys-in-go)
		signedMap := adv
		delete(signedMap, "signature")

		signedData, err := json.Marshal(signedMap)
		if err != nil {
			return fmt.Errorf("error packing data for signature verification: %v", err)
		}

		// verify the signature
		if err = peer.Keys.VerifyMessage(signedData, signature); err != nil {
			return fmt.Errorf("peer %x signature is invalid", peer.SessionID)
		}
	 */

	peer.Channel = wifi.Freq2Chan(int(radio.ChannelFrequency))
	peer.RSSI = int(radio.DBMAntennaSignal)

	if !bytes.Equal(peer.SessionID, dot11.Address3) {
		log.Info("peer %s changed session id: %x -> %x", peer.ID(), peer.SessionIDStr, dot11.Address3)
		copy(peer.SessionID, dot11.Address3)
		parts := make([]string, 6)
		for idx, byte := range peer.SessionID {
			parts[idx] = fmt.Sprintf("%02x", byte)
		}
		peer.SessionIDStr = strings.Join(parts, ":")
	}

	for key, value := range adv {
		peer.AdvData.Store(key, value)
	}

	return nil
}

func (peer *Peer) ID() string {
	name, _ := peer.AdvData.Load("name")
	ident := "???"

	if peer.Keys != nil {
		ident = peer.Keys.FingerprintHex
	} else if _ident, found := peer.AdvData.Load("identity"); found {
		ident = _ident.(string)
	}

	return fmt.Sprintf("%s@%s", name, ident)
}

func (peer *Peer) InactiveFor() float64 {
	peer.Lock()
	defer peer.Unlock()
	return time.Since(peer.DetectedAt).Seconds()
}

func (peer *Peer) SetData(adv map[string]interface{}) {
	peer.Lock()
	defer peer.Unlock()

	for key, val := range adv {
		if val == nil {
			peer.AdvData.Delete(key)
		} else {
			peer.AdvData.Store(key, val)
		}
	}
}

func (peer *Peer) Data() map[string]interface{} {
	peer.Lock()
	defer peer.Unlock()
	return peer.dataFrame()
}

func (peer *Peer) dataFrame() map[string]interface{} {
	data := map[string]interface{}{}
	peer.AdvData.Range(func(key, value interface{}) bool {
		data[key.(string)] = value
		return true
	})
	return data
}

func (peer *Peer) advertise() {
	peer.Lock()
	defer peer.Unlock()

	if peer.advEnabled {
		data := peer.dataFrame()

		data["timestamp"] = time.Now().Unix()
		adv, err := json.Marshal(data)
		if err != nil {
			log.Error("could not serialize advertisement data: %v", err)
			return
		}

		/*
			No need for signature in the advertisement protocol, however:

			// sign the advertisement
			signature, err := peer.Keys.SignMessage(adv)
			if err != nil {
				log.Error("error signing advertisement: %v", err)
				return
			}

			// add the signature to the advertisement itself and encode again
			data["signature"] = base64.StdEncoding.EncodeToString(signature)
			adv, err = json.Marshal(data)
			if err != nil {
				log.Error("could not serialize signed advertisement data: %v", err)
				return
			}

			log.Debug("advertising:\n%+v", data)
		*/

		err, raw := wifi.Pack(
			net.HardwareAddr(peer.SessionID),
			wifi.BroadcastAddr,
			adv,
			false) // set compression to true if using signature
		if err != nil {
			log.Error("could not encapsulate %d bytes of advertisement data: %v", len(adv), err)
			return
		}

		if err = peer.mux.Write(raw); err != nil {
			log.Error("error sending %d bytes of advertisement frame: %v", len(raw), err)
		}
	}
}

func (peer *Peer) StartAdvertising(iface string) (err error) {
	if peer.mux == nil {
		if peer.mux, err = NewPacketMuxer(iface, "", Workers); err != nil {
			return
		}
	}

	go func() {
		period := time.Duration(peer.AdvPeriod) * time.Millisecond
		ticker := time.NewTicker(period)

		log.Debug("advertiser started with a %s period", period)

		for {
			select {
			case _ = <-ticker.C:
				peer.advertise()
			case <-peer.stop:
				log.Info("advertiser stopped")
				return
			}
		}
	}()

	return nil
}

func (peer *Peer) StopAdvertising() {
	log.Debug("stopping advertiser ...")
	peer.stop <- struct{}{}
}
