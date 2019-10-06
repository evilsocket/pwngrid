package api

import (
	"encoding/json"
	"errors"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/models"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	ErrEmpty = errors.New("")
)

func (api *API) UnitEnroll(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	clientAddress := strings.Split(r.RemoteAddr, ":")[0]
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		clientAddress = forwardedFor
	}

	// https://support.cloudflare.com/hc/en-us/articles/206776727-What-is-True-Client-IP-
	if trueClient := r.Header.Get("True-Client-IP"); trueClient != "" {
		clientAddress = trueClient
	}

	// handle ip, ip
	clientAddress = strings.Trim(strings.Split(clientAddress, ",")[0], " ")

	var enroll UnitEnrollmentRequest
	if err = json.Unmarshal(body, &enroll); err != nil {
		log.Warning("error while reading enrollment request from %s: %v", clientAddress, err)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err = enroll.Validate(); err != nil {
		log.Warning("error while validating enrollment request from %s: %v", clientAddress, err)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	unit := models.FindUnitByFingerprint(api.DB, enroll.Fingerprint)
	if unit == nil {
		log.Info("enrolling new unit for %s: %s", clientAddress, enroll.Identity)

		unit = &models.Unit{
			Address:   clientAddress,
			Name:  enroll.Name,
			Fingerprint: enroll.Fingerprint,
			PublicKey: string(enroll.KeyPair.PublicPEM),
			Token:     "",
			CreatedAt: time.Now(),
		}

		if err = api.DB.Model(&models.Unit{}).Create(unit).Error; err != nil {
			log.Error("error enrolling %s: %v", unit.Identity(), err)
			ERROR(w, http.StatusInternalServerError, ErrEmpty)
			return
		}
	}

	if unit.Name != enroll.Name {
		log.Info("unit %s changed name: %s -> %s", unit.Identity(), unit.Name, enroll.Name)
		unit.Name = enroll.Name
	}

	unit.Address = clientAddress
	if unit.Token, err = CreateTokenFor(unit); err != nil {
		log.Error("error creating token for %s: %v", unit.Identity(), err)
		ERROR(w, http.StatusInternalServerError, ErrEmpty)
		return
	}

	err = api.DB.Model(unit).UpdateColumns(map[string]interface{}{
		"token":      unit.Token,
		"name":   unit.Name,
		"address":    unit.Address,
		"updated_at": time.Now(),
	}).Error
	if err != nil {
		log.Error("error setting token for %s: %v", unit.Identity(), err)
		ERROR(w, http.StatusInternalServerError, ErrEmpty)
		return
	}

	log.Debug("unit %s enrolled: id:%d address:%s", unit.Identity(), unit.ID, unit.Address)

	JSON(w, http.StatusOK, map[string]string{
		"token": unit.Token,
	})
}