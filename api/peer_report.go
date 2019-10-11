package api

import (
	"encoding/json"
	"github.com/evilsocket/islazy/log"
	"io/ioutil"
	"net/http"
)

// POST /api/v1/report/ap
func (api *API) PeerReportAP(w http.ResponseWriter, r *http.Request) {
	var report apReport

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	log.Debug("%s", body)

	if err = json.Unmarshal(body, &report); err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := api.Client.ReportAP(report)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	JSON(w, http.StatusOK, obj)
}
