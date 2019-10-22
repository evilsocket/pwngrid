package mesh

import (
	"encoding/json"
	"github.com/evilsocket/islazy/log"
	"time"
)

type jsonPeer struct {
	Fingerprint   string                 `json:"fingerprint"`
	MetAt         time.Time              `json:"met_at"`
	DetectedAt    time.Time              `json:"detected_at"`
	SeenAt        time.Time              `json:"seen_at"`
	PrevSeenAt    time.Time              `json:"prev_seen_at"`
	Encounters    int                    `json:"encounters"`
	Bond          float64                `json:"bond"`
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
	e := float64(peer.Encounters / 50.0)
	bond = e / (t * 10.0)

	log.Debug("bond for %s: hours_since_met=%f encounters=%d bond=%f",
		fingerprint,
		time.Since(peer.MetAt).Hours(),
		peer.Encounters,
		bond)

	doc := jsonPeer{
		Fingerprint:   fingerprint,
		MetAt:         peer.MetAt,
		Encounters:    peer.Encounters,
		Bond:          bond,
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
