package mesh

import (
	"encoding/json"
	"time"
)

type jsonPeer struct {
	Fingerprint   string                 `json:"fingerprint"`
	MetAt         time.Time              `json:"met_at"`
	Encounters    int                    `json:"encounters"`
	Bond          float64                `json:"bond"`
	DetectedAt    time.Time              `json:"detected_at"`
	SeenAt        time.Time              `json:"seen_at"`
	Channel       int                    `json:"channel"`
	RSSI          int                    `json:"rssi"`
	SessionID     string                 `json:"session_id"`
	Advertisement map[string]interface{} `json:"advertisement"`
}

func (peer *Peer) json() *jsonPeer {
	fingerprint := ""
	if v, found := peer.AdvData.Load("identity"); found {
		fingerprint = v.(string)
	}

	// see https://www.patreon.com/posts/bonding-equation-30954153
	var bond float64
	t := float64(time.Since(peer.MetAt).Hours() + 1e-50) // avoid division by 0
	e := float64(peer.Encounters)
	bond = e / (t * 10.0)

	doc := jsonPeer{
		Fingerprint:   fingerprint,
		MetAt:         peer.MetAt,
		Encounters:    peer.Encounters,
		Bond:          bond,
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
