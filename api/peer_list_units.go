package api

import (
	"net/http"
)

func (api *API) PeerListUnits(w http.ResponseWriter, r *http.Request) {
	page, err := pageNum(r)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := api.Client.PagedUnits(page)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	JSON(w, http.StatusOK, obj)
}
