package api

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
	"net/http"

	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/models"
)

type API struct {
	DB     *gorm.DB
	Router *chi.Mux
}

func Setup(DbUser, DbPassword, DbPort, DbHost, DbName string) (err error, api *API) {
	log.Info("connecting to %s:%s ...", DbHost, DbPort)
	api = &API{}
	dbURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", DbUser, DbPassword, DbHost, DbPort, DbName)
	if api.DB, err = gorm.Open("mysql", dbURL); err != nil {
		return
	}
	api.DB.Debug().AutoMigrate(&models.Unit{}, &models.AccessPoint{})

	api.Router = chi.NewRouter()

	api.Router.Use(CORS)

	api.Router.Route("/api", func(r chi.Router) {
		r.Options("/", corsRoute)

		r.Route("/v1", func(r chi.Router) {
			r.Route("/units", func(r chi.Router) {
				r.Get("/", api.ListUnits)
				r.Get("/by_country", api.UnitsByCountry)
			})

			r.Route("/unit", func(r chi.Router) {
				r.Get("/{fingerprint:[a-fA-F0-9]+}", api.ShowUnit)

				r.Get("/unit/", api.ListUnits)

				r.Post("/enroll", api.UnitEnroll)

				r.Route("/report", func(r chi.Router) {
					r.Post("/ap", api.UnitReportAP)
				})
			})
		})
	})

	return
}

func (api *API) Run(addr string) {
	log.Info("pwngrid api starting on %s ...", addr)
	log.Fatal("%v", http.ListenAndServe(addr, api.Router))
}