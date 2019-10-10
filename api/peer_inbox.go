package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/models"
	"github.com/go-chi/chi"
	"io/ioutil"
	"net/http"
)

var (
	ErrEmptyMessage = errors.New("empty message body")
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
	messageBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("error reading request body: %v", err)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	messageSize := len(messageBody)
	if messageSize == 0 {
		ERROR(w, http.StatusUnprocessableEntity, ErrEmptyMessage)
		return
	} else if messageSize > models.MessageDataMaxSize {
		err := fmt.Errorf("max message signature size is %d", models.MessageSignatureMaxSize)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	fingerprint := chi.URLParam(r, "fingerprint")

	log.Info("signing new message of %d bytes for %s ...", len(messageBody), fingerprint)

	signature, err := api.Keys.SignMessage(messageBody)
	if err != nil {
		log.Error("%v", err)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	msg := Message{
		Signature: base64.StdEncoding.EncodeToString(signature),
		Data:      base64.StdEncoding.EncodeToString(messageBody),
	}

	log.Debug("%v", msg)

	if err := api.Client.SendMessageTo(fingerprint, msg); err != nil {
		log.Error("%v", err)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
