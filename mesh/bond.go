package mesh

import (
	"github.com/evilsocket/islazy/log"
	"time"
)

func (peer *Peer) Bond() float64 {
	// see https://www.patreon.com/posts/bonding-equation-30954153
	daysSinceMet := time.Since(peer.MetAt).Hours() / 24.0
	// assuming an average of at least 100 encounters per day
	// with a 10% statistical loss
	maxEncounters := (daysSinceMet * 100) * .9
	bond := float64(peer.Encounters) / (maxEncounters + 1e-50) // avoid division by 0

	log.Debug("bond with %s: days_since_met=%f max_enc=%f encounters=%d bond=%f",
		peer.ID(),
		daysSinceMet,
		maxEncounters,
		peer.Encounters,
		bond)

	return bond
}
