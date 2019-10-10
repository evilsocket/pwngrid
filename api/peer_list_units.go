package api

import (
	"fmt"
	"net/http"
)

func (api *API) PeerListUnits(w http.ResponseWriter, r *http.Request) {
	page, err := pageNum(r)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	url := fmt.Sprintf("%s/units/?page=%d", Endpoint, page)

	obj, err := GetJSON(url)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	JSON(w, http.StatusOK, obj)
}
