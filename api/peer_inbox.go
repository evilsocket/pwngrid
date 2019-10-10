package api

import (
	"net/http"
)

// /api/v1/inbox
func (api *API) PeerGetInbox(w http.ResponseWriter, r *http.Request) {
	obj, err := api.Client.Inbox()
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	JSON(w, http.StatusOK, obj)
}

// POST /api/v1/unit/<fingerprint>/inbox
func (api *API) PeerSendMessageTo(w http.ResponseWriter, r *http.Request) {
	// fingerprint := chi.URLParam(r, "fingerprint")


	/*
	type Message struct {
		Data      string `json:"data"`
		Signature string `json:"signature"`
	}
	 */

	/*
	obj, err := api.Client.Inbox()
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	JSON(w, http.StatusOK, obj)
	 */
}
