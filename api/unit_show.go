package api

import (
	"github.com/evilsocket/pwngrid/models"
	"github.com/go-chi/chi"
	"net/http"
)

func (api *API) ShowUnit(w http.ResponseWriter, r *http.Request) {
	unitFingerprint := chi.URLParam(r, "fingerprint")
	if unit := models.FindUnitByFingerprint(unitFingerprint); unit == nil {
		ERROR(w, http.StatusNotFound, ErrEmpty)
		return
	} else {
		w.Header().Set("Cache-Control", "max-age:600, public")
		w.Header().Set("Expires", "600")
		JSON(w, http.StatusOK, unit)
	}
}
