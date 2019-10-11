package api

import (
	"encoding/json"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/models"
	"io/ioutil"
	"net"
	"net/http"
)

type apReport struct {
	ESSID string `json:"essid"`
	BSSID string `json:"bssid"`
}

func (api *API) UnitReportAP(w http.ResponseWriter, r *http.Request) {
	unit := Authenticate(w, r)
	if unit == nil {
		return
	}

	client := clientIP(r)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, ErrEmpty)
		return
	}

	var ap apReport
	if err = json.Unmarshal(body, &ap); err != nil {
		log.Warning("error while reading wifi ap from %s: %v", client, err)
		ERROR(w, http.StatusUnprocessableEntity, ErrEmpty)
		return
	}

	if parsed, err := net.ParseMAC(ap.BSSID); err != nil {
		log.Warning("error while parsing wifi ap bssid %s from %s: %v", ap.BSSID, client, err)
		ERROR(w, http.StatusUnprocessableEntity, ErrEmpty)
		return
	} else {
		// normalize
		ap.BSSID = parsed.String()
	}

	if existing := unit.FindAccessPoint(ap.ESSID, ap.BSSID); existing == nil {
		log.Debug("unit %s (%s %s) reporting new wifi access point %v", unit.Identity(), unit.Address,
			unit.Country, ap)

		newAP := models.AccessPoint{
			Name:   ap.ESSID,
			Mac:    ap.BSSID,
			UnitID: unit.ID,
		}

		if err := models.Create(&newAP).Error; err != nil {
			log.Warning("error creating ap %v: %v", newAP, err)
			ERROR(w, http.StatusUnprocessableEntity, ErrEmpty)
			return
		}
	} else if err := models.Update(existing).Error; err != nil {
		log.Warning("error updating ap %v: %v", existing, err)
		ERROR(w, http.StatusUnprocessableEntity, ErrEmpty)
		return
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
