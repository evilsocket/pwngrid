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
	"strconv"
	"time"
)

func (api *API) GetInbox(w http.ResponseWriter, r *http.Request) {
	unit := Authenticate(w, r)
	if unit == nil {
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

// we need this because the models.Message structure doesn't not export data and signature for fast listing.
type fullMessage struct {
	ID         uint       `json:"id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at"`
	SeenAt     *time.Time `json:"seen_at"`
	Sender     string     `json:"sender"`
	SenderName string     `json:"sender_name"`
	Data       string     `json:"data"`
	Signature  string     `json:"signature"`
}

func (api *API) GetInboxMessage(w http.ResponseWriter, r *http.Request) {
	unit := Authenticate(w, r)
	if unit == nil {
		return
	}

	msgIDParam := chi.URLParam(r, "msg_id")
	msgID, err := strconv.Atoi(msgIDParam)

	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	} else if message := unit.GetInboxMessage(msgID); message == nil {
		ERROR(w, http.StatusNotFound, ErrEmpty)
		return
	} else {
		JSON(w, http.StatusOK, fullMessage{
			ID:         message.ID,
			CreatedAt:  message.CreatedAt,
			UpdatedAt:  message.UpdatedAt,
			DeletedAt:  message.DeletedAt,
			SeenAt:     message.SeenAt,
			Sender:     message.Sender,
			SenderName: message.SenderName,
			Data:       message.Data,
			Signature:  message.Signature,
		})
	}
}

func (api *API) MarkInboxMessage(w http.ResponseWriter, r *http.Request) {
	unit := Authenticate(w, r)
	if unit == nil {
		return
	}

	now := time.Now()
	markAs := chi.URLParam(r, "mark")
	msgIDParam := chi.URLParam(r, "msg_id")
	msgID, err := strconv.Atoi(msgIDParam)

	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	} else if message := unit.GetInboxMessage(msgID); message == nil {
		ERROR(w, http.StatusNotFound, ErrEmpty)
		return
	} else if markAs == "seen" {
		if err := models.UpdateFields(message, map[string]interface{}{"seen_at": &now}).Error; err != nil {
			ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
	} else if markAs == "unseen" {
		if err := models.UpdateFields(message, map[string]interface{}{"seen_at": nil}).Error; err != nil {
			ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
	} else if markAs == "deleted" {
		if err := models.UpdateFields(message, map[string]interface{}{"deleted_at": &now}).Error; err != nil {
			ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
	} else if markAs == "restored" {
		if err := models.UpdateFields(message, map[string]interface{}{"deleted_at": nil}).Error; err != nil {
			ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
	} else {
		ERROR(w, http.StatusNotFound, ErrEmpty)
		return
	}

	JSON(w, http.StatusOK, map[string]bool{
		"success": true,
	})
}

func (api *API) SendMessageTo(w http.ResponseWriter, r *http.Request) {
	// authenticate source unit
	srcUnit := Authenticate(w, r)
	if srcUnit == nil {
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
		Sender:     srcUnit.Fingerprint,
		SenderName: srcUnit.Name,
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
