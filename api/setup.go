package api

import (
	"fmt"
	"github.com/evilsocket/pwngrid/crypto"
	"github.com/go-chi/chi"
	"github.com/go-chi/docgen"
	"net/http"

	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/evilsocket/islazy/log"
)

type API struct {
	Router *chi.Mux
	Keys   *crypto.KeyPair
	Client *Client
}

func Setup(keys *crypto.KeyPair, routes bool) (err error, api *API) {
	api = &API{
		Router: chi.NewRouter(),
		Keys:   keys,
		Client: NewClient(keys),
	}

	api.Router.Use(CORS)
	api.Router.Route("/api", func(r chi.Router) {
		r.Options("/", corsRoute)

		r.Route("/v1", func(r chi.Router) {
			if api.Keys == nil {
				log.Debug("registering server api ...")

				r.Route("/units", func(r chi.Router) {
					// /api/v1/units/
					r.Get("/", api.ListUnits)
					// /api/v1/units/by_country
					r.Get("/by_country", api.UnitsByCountry)
				})

				r.Route("/unit", func(r chi.Router) {
					// /api/v1/unit/deadbeefdeadbeef
					r.Get("/{fingerprint:[a-fA-F0-9]+}", api.ShowUnit)

					// /api/v1/unit/inbox
					r.Get("/inbox", api.GetInbox)

					// PUT /api/v1/unit/deadbeefdeadbeef/inbox
					r.Put("/{fingerprint:[a-fA-F0-9]+}/inbox", api.SendMessageTo)

					// POST /api/v1/unit/enroll
					r.Post("/enroll", api.UnitEnroll)
					r.Route("/report", func(r chi.Router) {
						// POST /api/v1/unit/report/ap
						r.Post("/ap", api.UnitReportAP)
					})
				})
			} else {
				log.Debug("registering peer api ...")

				r.Get("/inbox", api.PeerGetInbox)

				r.Route("/units", func(r chi.Router) {
					r.Get("/", api.PeerListUnits)
					// r.Put("/{fingerprint:[a-fA-F0-9]+}/inbox", api.PeerSendMessageTo)
				})
			}
		})
	})

	if routes {
		fmt.Println(docgen.MarkdownRoutesDoc(api.Router, docgen.MarkdownOpts{
			ProjectPath: "github.com/evilsocket/pwngrid",
			Intro:       "Welcome to the pwngrid API generated docs.",
		}))
	}

	return
}

func (api *API) Run(addr string) {
	log.Info("pwngrid api starting on %s ...", addr)
	log.Fatal("%v", http.ListenAndServe(addr, api.Router))
}
