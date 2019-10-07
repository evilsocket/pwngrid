package api

import (
	"encoding/json"
	"errors"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/models"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	ErrEmpty = errors.New("")
)

func (api *API) readEnrollment(w http.ResponseWriter, r *http.Request) (error, UnitEnrollmentRequest) {
	var enroll UnitEnrollmentRequest

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return err, enroll
	}

	enroll.Address = strings.Split(r.RemoteAddr, ":")[0]
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		enroll.Address = forwardedFor
	}

	// https://support.cloudflare.com/hc/en-us/articles/206776727-What-is-True-Client-IP-
	if trueClient := r.Header.Get("True-Client-IP"); trueClient != "" {
		enroll.Address = trueClient
	}

	// handle "ip, ip"
	enroll.Address = strings.Trim(strings.Split(enroll.Address, ",")[0], " ")
	enroll.Country = r.Header.Get("CF-IPCountry")

	if err = json.Unmarshal(body, &enroll); err != nil {
		log.Warning("error while reading enrollment request from %s: %v", enroll.Address, err)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return err, enroll
	}

	if err = enroll.Validate(); err != nil {
		log.Warning("error while validating enrollment request from %s: %v", enroll.Address, err)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return err, enroll
	}

	return nil, enroll
}

func (api *API) UnitEnroll(w http.ResponseWriter, r *http.Request) {
	err, enroll := api.readEnrollment(w, r)
	if err != nil {
		return
	}

	err, unit := models.EnrollUnit(api.DB, enroll)
	if err != nil {
		log.Error("%v", err)
		ERROR(w, http.StatusInternalServerError, ErrEmpty)
		return
	}

	log.Debug("unit %s enrolled: id:%d address:%s", unit.Identity(), unit.ID, unit.Address)

	JSON(w, http.StatusOK, map[string]string{
		"token": unit.Token,
	})
}
