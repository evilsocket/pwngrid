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

func (api *API) InboxMessage(id int)(map[string]interface{}, int, error) {
	message, err := api.Client.InboxMessage(id)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}

	log.Info("%+v", message)

	sender, found := message["sender"]
	if !found {
		return nil, http.StatusNotFound, ErrSenderNotFound
	}

	fingerprint, ok := sender.(string)
	if !ok {
		return nil, http.StatusUnprocessableEntity, ErrSenderNotFound
	}

	unit, err := api.Client.Unit(fingerprint)
	if err != nil {
		return nil, http.StatusNotFound, err
	}

	srcKeys, err := crypto.FromPublicPEM(unit["public_key"].(string))
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}

	data, err := base64.StdEncoding.DecodeString(message["data"].(string))
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}

	signature, err := base64.StdEncoding.DecodeString(message["signature"].(string))
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}

	log.Info("verifying message from %s ...", fingerprint)

	if err := srcKeys.VerifyMessage(data, signature); err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}

	log.Info("decrypting message from %s ...", fingerprint)

	clearText, err := api.Keys.Decrypt(data)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}

	message["data"] = clearText

	return message, 0, nil
}

// /api/v1/inbox/<msg_id>
func (api *API) PeerGetInboxMessage(w http.ResponseWriter, r *http.Request) {
	msgIDParam := chi.URLParam(r, "msg_id")
	msgID, err := strconv.Atoi(msgIDParam)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	message, status, err := api.InboxMessage(msgID)
	if err != nil {
		ERROR(w, status, err)
		return
	}

	JSON(w, http.StatusOK, message)
}

// /api/v1/inbox/<msg_id>/<mark>
func (api *API) PeerMarkInboxMessage(w http.ResponseWriter, r *http.Request) {
	markAs := chi.URLParam(r, "mark")
	msgIDParam := chi.URLParam(r, "msg_id")
	msgID, err := strconv.Atoi(msgIDParam)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := api.Client.MarkInboxMessage(msgID, markAs)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	JSON(w, http.StatusOK, obj)
}

func (api *API) SendMessage(fingerprint string, cleartext []byte) (int, error) {
	unit, err := api.Client.Unit(fingerprint)
	if err != nil {
		return http.StatusNotFound, err
	}

	unitKeys, err := crypto.FromPublicPEM(unit["public_key"].(string))
	if err != nil {
		log.Error("error parsing public key of %s: %v", fingerprint, err)
		return http.StatusUnprocessableEntity, err
	}

	messageBody, err := api.Keys.EncryptFor(cleartext, unitKeys.Public)
	if err != nil {
		log.Error("error encrypting message for %s: %v", fingerprint, err)
		return http.StatusUnprocessableEntity, err
	}

	messageSize := len(messageBody)
	if messageSize == 0 {
		return http.StatusUnprocessableEntity, ErrEmptyMessage
	} else if messageSize > models.MessageDataMaxSize {
		err := fmt.Errorf("max message signature size is %d", models.MessageSignatureMaxSize)
		return http.StatusUnprocessableEntity, err
	}

	log.Info("signing encrypted message of %d bytes for %s ...", messageSize, fingerprint)

	signature, err := api.Keys.SignMessage(messageBody)
	if err != nil {
		log.Error("%v", err)
		return http.StatusUnprocessableEntity, err
	}

	msg := Message{
		Signature: base64.StdEncoding.EncodeToString(signature),
		Data:      base64.StdEncoding.EncodeToString(messageBody),
	}

	if err := api.Client.SendMessageTo(fingerprint, msg); err != nil {
		log.Error("%v", err)
		return http.StatusUnprocessableEntity, err
	}

	return 0, nil
}

// POST /api/v1/unit/<fingerprint>/inbox
func (api *API) PeerSendMessageTo(w http.ResponseWriter, r *http.Request) {
	cleartextMessage, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("error reading request body: %v", err)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	fingerprint := chi.URLParam(r, "fingerprint")
	status, err := api.SendMessage(fingerprint, cleartextMessage)
	if err != nil {
		ERROR(w, status, err)
		return
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
