package api

import (
	"net/http"
)

func (api *API) PeerGetInbox(w http.ResponseWriter, r *http.Request) {
	obj, err := api.Client.Inbox()
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	JSON(w, http.StatusOK, obj)
}
