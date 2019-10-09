package api

import (
	"github.com/go-chi/chi"
	"net/http"

	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/evilsocket/islazy/log"
)

type API struct {
	Router *chi.Mux
}

func Setup() (err error, api *API) {
	api = &API{}

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