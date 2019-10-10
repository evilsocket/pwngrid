package api

import (
	"fmt"
	"net/http"
)

func (api *API) PeerGetInbox(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%s/units/inbox", Endpoint)
	obj, err := GetJSON(url)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	JSON(w, http.StatusOK, obj)
}
