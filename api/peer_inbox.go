package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/crypto"
	"github.com/evilsocket/pwngrid/models"
	"github.com/go-chi/chi"
	"io/ioutil"
	"net/http"
	"strconv"
)

var (
	ErrEmptyMessage   = errors.New("empty message body")
	ErrSenderNotFound = errors.New("sender not found")
)

// /api/v1/inbox/
func (api *API) PeerGetInbox(w http.ResponseWriter, r *http.Request) {
	page, err := pageNum(r)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := api.Client.Inbox(page)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	JSON(w, http.StatusOK, obj)
}

// /api/v1/inbox/<msg_id>
func (api *API) PeerGetInboxMessage(w http.ResponseWriter, r *http.Request) {
	msgIDParam := chi.URLParam(r, "msg_id")
	msgID, err := strconv.Atoi(msgIDParam)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	message, err := api.Client.InboxMessage(msgID)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	sender, found := message["sender"]
	if !found {
		ERROR(w, http.StatusNotFound, ErrSenderNotFound)
		return
	}

	fingerprint, ok := sender.(string)
	if !ok {
		ERROR(w, http.StatusUnprocessableEntity, ErrSenderNotFound)
		return
	}

	unit, err := api.Client.Unit(fingerprint)
	if err != nil {
		ERROR(w, http.StatusNotFound, err)
		return
	}

	srcKeys,err := crypto.FromPublicPEM(unit["public_key"].(string))
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	data, err := base64.StdEncoding.DecodeString(message["data"].(string))
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	signature, err := base64.StdEncoding.DecodeString(message["signature"].(string))
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	log.Info("verifying message from %s ...", fingerprint)

	if err := srcKeys.VerifyMessage(data, signature); err !=  nil{
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	log.Info("decrypting message from %s ...", fingerprint)

	clearText, err := api.Keys.Decrypt(data)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	message["data"] = clearText

	JSON(w, http.StatusOK, message)
}

// POST /api/v1/unit/<fingerprint>/inbox
func (api *API) PeerSendMessageTo(w http.ResponseWriter, r *http.Request) {
	cleartextMessage, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("error reading request body: %v", err)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// get the peer public signature
	fingerprint := chi.URLParam(r, "fingerprint")
	unit, err := api.Client.Unit(fingerprint)
	if err != nil {
		ERROR(w, http.StatusNotFound, err)
		return
	}

	unitKeys, err := crypto.FromPublicPEM(unit["public_key"].(string))
	if err != nil {
		log.Error("error parsing public key of %s: %v", fingerprint, err)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	cleartextSize := len(cleartextMessage)

	log.Info("encrypting message of %d bytes for %s ...", cleartextSize, fingerprint)

	messageBody, err := api.Keys.EncryptFor(cleartextMessage, unitKeys.Public)
	if err != nil {
		log.Error("error encrypting message for %s: %v", fingerprint, err)
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

	log.Info("signing encrypted message of %d bytes for %s ...", messageSize, fingerprint)

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
