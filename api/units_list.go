package api

import (
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/models"
	"net/http"
)

func (api *API) ListUnits(w http.ResponseWriter, r *http.Request) {
	page, err := pageNum(r)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	units, total, pages := models.GetPagedUnits(page)

	w.Header().Set("Cache-Control", "max-age:600, public")
	w.Header().Set("Expires", "600")

	JSON(w, http.StatusOK, map[string]interface{}{
		"records": total,
		"pages":   pages,
		"units":   units,
	})
}

func (api *API) UnitsByCountry(w http.ResponseWriter, r *http.Request) {
	if results, err := models.GetUnitsByCountry(); err != nil {
		log.Warning("error getting units by country: %v", err)
		ERROR(w, http.StatusInternalServerError, err)
		return
	} else {
		w.Header().Set("Cache-Control", "max-age:600, public")
		w.Header().Set("Expires", "600")

		JSON(w, http.StatusOK, results)
	}
}
