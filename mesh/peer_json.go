package mesh

import (
	"encoding/json"
	"github.com/evilsocket/islazy/log"
	"net"
	"sync"
	"time"
)

type jsonPeer struct {
	Fingerprint   string                 `json:"fingerprint"`
	MetAt         time.Time              `json:"met_at"`
	DetectedAt    time.Time              `json:"detected_at"`
	SeenAt        time.Time              `json:"seen_at"`
	PrevSeenAt    time.Time              `json:"prev_seen_at"`
	Encounters    uint64                 `json:"encounters"`
	Channel       int                    `json:"channel"`
	RSSI          int                    `json:"rssi"`
	SessionID     string                 `json:"session_id"`
	Advertisement map[string]interface{} `json:"advertisement"`
}

// creates a Peer object filled with the fields of the JSON representation
func peerFromJSON(j jsonPeer) *Peer {
	peer := &Peer{
		DetectedAt:   j.DetectedAt,
		SeenAt:       j.SeenAt,
		PrevSeenAt:   j.PrevSeenAt,
		SessionIDStr: j.SessionID,
		Encounters:   j.Encounters,
		Channel:      j.Channel,
		RSSI:         j.RSSI,
		AdvData:      sync.Map{},
	}

	if hw, err := net.ParseMAC(j.SessionID); err == nil {
		copy(peer.SessionID, hw)
	} else {
		log.Warning("error parsing peer session id %s: %v", j.SessionID, err)
	}

	for key, val := range j.Advertisement {
		peer.AdvData.Store(key, val)
	}

	return peer
}

// converts a peer into a JSON friendly representation
func (peer *Peer) json() *jsonPeer {
	fingerprint := ""
	if v, found := peer.AdvData.Load("identity"); found {
		fingerprint = v.(string)
	}

	doc := jsonPeer{
		Fingerprint:   fingerprint,
		MetAt:         peer.MetAt,
		Encounters:    peer.Encounters,
		PrevSeenAt:    peer.PrevSeenAt,
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

	return &doc
}

func (peer *Peer) MarshalJSON() ([]byte, error) {
	peer.Lock()
	defer peer.Unlock()
	return json.Marshal(peer.json())
}
