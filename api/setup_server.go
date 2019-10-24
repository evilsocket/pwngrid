package api

import (
	"fmt"
	"github.com/evilsocket/islazy/log"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"net/http"
)

func cached(seconds int, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", fmt.Sprintf("public, max-age=%d", seconds))
		w.Header().Add("Expires", fmt.Sprintf("%d", seconds))
		next.ServeHTTP(w, r)
	}
}

func (api *API) setupServerRoutes() {
	log.Debug("registering server api ...")

	api.Router.Use(middleware.DefaultCompress)

	api.Router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Route("/units", func(r chi.Router) {
				// GET /api/v1/units/
				r.Get("/", cached(600, api.ListUnits))
				// GET /api/v1/units/by_country
				r.Get("/by_country", cached(600, api.UnitsByCountry))
			})
			r.Route("/unit", func(r chi.Router) {
				// GET /api/v1/unit/<fingerprint>
				r.Get("/{fingerprint:[a-fA-F0-9]+}", cached(600, api.ShowUnit))
				r.Route("/inbox", func(r chi.Router) {
					// GET /api/v1/unit/inbox/
					r.Get("/", api.GetInbox)
					r.Route("/{msg_id:[0-9]+}", func(r chi.Router) {
						// GET /api/v1/unit/inbox/<msg_id>
						r.Get("/", api.GetInboxMessage)
						// GET /api/v1/unit/inbox/<msg_id>/<mark>
						r.Get("/{mark:[a-z]+}", api.MarkInboxMessage)
					})
				})
				// POST /api/v1/unit/<fingerprint>/inbox
				r.Post("/{fingerprint:[a-fA-F0-9]+}/inbox", api.SendMessageTo)
				// POST /api/v1/unit/enroll
				r.Post("/enroll", api.UnitEnroll)
				r.Route("/report", func(r chi.Router) {
					// POST /api/v1/unit/report/ap
					r.Post("/ap", api.UnitReportAP)
					// POST /api/v1/unit/report/aps
					r.Post("/aps", api.UnitReportMultipleAP)
				})
			})
		})
	})
}
