package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/crypto"
	"github.com/evilsocket/pwngrid/models"
	"github.com/go-chi/chi"
	"io/ioutil"
	"net/http"
)

func (api *API) GetInbox(w http.ResponseWriter, r *http.Request) {
	unit := Authenticate(w, r)
	if unit == nil {
		ERROR(w, http.StatusForbidden, ErrEmpty)
		return
	}

	page, err := pageNum(r)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	messages, total, pages := unit.GetPagedInbox(page)
	JSON(w, http.StatusOK, map[string]interface{}{
		"records":  total,
		"pages":    pages,
		"messages": messages,
	})
}

func (api *API) SendMessageTo(w http.ResponseWriter, r *http.Request) {
	// authenticate source unit
	srcUnit := Authenticate(w, r)
	if srcUnit == nil {
		ERROR(w, http.StatusForbidden, ErrEmpty)
		return
	}

	// get dest unit by fingerprint
	dstUnitFingerprint := chi.URLParam(r, "fingerprint")
	dstUnit := models.FindUnitByFingerprint(dstUnitFingerprint)
	if dstUnit == nil {
		ERROR(w, http.StatusNotFound, ErrEmpty)
		return
	}

	// read the message and signature from the source unit
	client := clientIP(r)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	var message Message
	if err = json.Unmarshal(body, &message); err != nil {
		log.Warning("error while decoding message from %s: %v", srcUnit.Identity(), err)
		log.Debug("%s", body)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// validate sizes
	if dataSize := len(message.Data); dataSize > models.MessageDataMaxSize {
		log.Warning("client %s sent a message of size %d", srcUnit.Identity(), dataSize)
		ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("max message data size is %d", models.MessageDataMaxSize))
		return
	} else if sigSize := len(message.Signature); sigSize > models.MessageSignatureMaxSize {
		log.Warning("client %s sent a message signature of size %d", srcUnit.Identity(), sigSize)
		ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("max message signature size is %d", models.MessageSignatureMaxSize))
		return
	}

	// parse source unit key
	srcKeys, err := crypto.FromPublicPEM(srcUnit.PublicKey)
	if err != nil {
		log.Warning("error decoding key from %s: %v", srcUnit.Identity(), err)
		log.Debug("%s", srcUnit.PublicKey)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// decode data, signature and verify SIGN(SHA256(data))
	data, err := base64.StdEncoding.DecodeString(message.Data)
	if err != nil {
		log.Warning("error decoding message from %s: %v", srcUnit.Identity(), err)
		log.Debug("%s", message.Data)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	signature, err := base64.StdEncoding.DecodeString(message.Signature)
	if err != nil {
		log.Warning("error decoding signature from %s: %v", srcUnit.Identity(), err)
		log.Debug("%s", message.Signature)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := srcKeys.VerifyMessage(data, signature); err != nil {
		log.Warning("error verifying signature from %s: %v", srcUnit.Identity(), err)
		log.Debug("%s", message.Signature)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	msg := models.Message{
		SenderID:   srcUnit.ID,
		ReceiverID: dstUnit.ID,
		Data:       message.Data,
		Signature:  message.Signature,
	}

	if err := models.Create(&msg).Error; err != nil {
		log.Warning("error creating msg %v from %s: %v", msg, client, err)
		ERROR(w, http.StatusInternalServerError, err)
		return
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
