package api

import (
	"encoding/json"
	"errors"
	"github.com/biezhi/gorm-paginator/pagination"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/models"
	"github.com/go-chi/chi"
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

type apReport struct {
	ESSID string `json:"essid"`
	BSSID string `json:"bssid"`
}

func (api *API) UnitReportAP(w http.ResponseWriter, r *http.Request) {
	unit := Authenticate(api.DB, w, r)
	if unit == nil {
		return
	}

	client := clientIP(r)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	var ap apReport
	if err = json.Unmarshal(body, &ap); err != nil {
		log.Warning("error while reading wifi ap from %s: %v", client, err)
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	if existing := unit.FindAccessPoint(api.DB, ap.ESSID, ap.BSSID); existing == nil {
		log.Info("unit %s (%s %s) reporting new wifi access point %v", unit.Identity(), unit.Address,
			unit.Country, ap)

		newAP := models.AccessPoint{
			Name:   ap.ESSID,
			Mac:    ap.BSSID,
			UnitID: unit.ID,
		}

		if err := api.DB.Create(&newAP).Error; err != nil {
			log.Warning("%v", err)
			ERROR(w, http.StatusInternalServerError, err)
			return
		}
	} else if err := api.DB.Save(existing).Error; err != nil {
		log.Warning("%v", err)
		ERROR(w, http.StatusInternalServerError, err)
		return
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func (api *API) ShowUnit(w http.ResponseWriter, r *http.Request) {
	unitFingerprint := chi.URLParam(r, "fingerprint")
	if unit := models.FindUnitByFingerprint(api.DB, unitFingerprint); unit == nil {
		ERROR(w, http.StatusNotFound, ErrEmpty)
		return
	} else {
		JSON(w, http.StatusOK, unit)
	}
}

func (api *API) ListUnits(w http.ResponseWriter, r *http.Request) {
	page, err := pageNum(r)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	var units []models.Unit

	paginator := pagination.Paging(&pagination.Param{
		DB:      api.DB,
		Page:    page,
		Limit:   512,
		OrderBy: []string{"id desc"},
	}, &units)

	JSON(w, http.StatusOK, map[string]interface{}{
		"records": paginator.TotalRecord,
		"pages":   paginator.TotalPage,
		"units":   units,
	})
}

type byCountry struct {
	Country string `json:"country"`
	Count   int    `json:"units"`
}

func (api *API) UnitsByCountry(w http.ResponseWriter, r *http.Request) {
	results := make([]byCountry, 0)
	if err := api.DB.Raw("SELECT country,COUNT(id) AS count FROM units GROUP BY country ORDER BY count DESC").Scan(&results).Error; err != nil {
		log.Warning("%v", err)
		ERROR(w, http.StatusInternalServerError, err)
		return
	}

	JSON(w, http.StatusOK, results)
}
