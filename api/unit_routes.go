package api

import (
	"encoding/json"
	"errors"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/models"
	"io/ioutil"
	"net/http"
)

var (
	ErrEmpty = errors.New("")
)

func (api *API) readEnrollment(w http.ResponseWriter, r *http.Request) (error, models.EnrollmentRequest) {
	var enroll models.EnrollmentRequest

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return err, enroll
	}

	enroll.Address = clientIP(r)
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
