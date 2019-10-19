package api

import (
	"fmt"
	"github.com/evilsocket/pwngrid/crypto"
	"github.com/evilsocket/pwngrid/mesh"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/docgen"
	"net/http"

	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/evilsocket/islazy/log"
)

type API struct {
	Router *chi.Mux
	Keys   *crypto.KeyPair
	Peer   *mesh.Peer
	Client *Client
}

func (api *API) setupServerRoutes() {
	log.Debug("registering server api ...")

	api.Router.Route("/api", func(r chi.Router) {
		r.Options("/", CORSOptionHandler)
		r.Route("/v1", func(r chi.Router) {
			r.Route("/units", func(r chi.Router) {
				// GET /api/v1/units/
				r.Get("/", api.ListUnits)
				// GET /api/v1/units/by_country
				r.Get("/by_country", api.UnitsByCountry)
			})
			r.Route("/unit", func(r chi.Router) {
				// GET /api/v1/unit/<fingerprint>
				r.Get("/{fingerprint:[a-fA-F0-9]+}", api.ShowUnit)
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
				})
			})
		})
	})
}

func (api *API) setupPeerRoutes() {
	log.Debug("registering peer api ...")

	api.Router.Route("/api", func(r chi.Router) {
		r.Options("/", CORSOptionHandler)
		r.Route("/v1", func(r chi.Router) {
			r.Route("/mesh", func(r chi.Router) {
				// GET /api/v1/mesh/peers
				r.Get("/peers", api.PeerGetPeers)

				// GET /api/v1/mesh/<status>
				r.Get("/{status:[a-z]+}", api.PeerSetSignaling)

				// GET /api/v1/mesh/data
				r.Get("/data", api.PeerGetMeshData)
				// POST /api/v1/mesh/data
				r.Post("/data", api.PeerSetMeshData)
			})

			// GET /api/v1/data
			r.Post("/data", api.PeerGetData)
			// POST /api/v1/data
			r.Post("/data", api.PeerSetData)

			r.Route("/report", func(r chi.Router) {
				// POST /api/v1/report/ap
				r.Post("/ap", api.PeerReportAP)
			})
			r.Route("/inbox", func(r chi.Router) {
				// GET /api/v1/inbox/
				r.Get("/", api.PeerGetInbox)
				r.Route("/{msg_id:[0-9]+}", func(r chi.Router) {
					// GET /api/v1/inbox/<msg_id>
					r.Get("/", api.PeerGetInboxMessage)
					// GET /api/v1/inbox/<msg_id>/<mark>
					r.Get("/{mark:[a-z]+}", api.PeerMarkInboxMessage)
				})
			})
			r.Route("/unit", func(r chi.Router) {
				// POST /api/v1/unit/<fingerprint>/inbox
				r.Post("/{fingerprint:[a-fA-F0-9]+}/inbox", api.PeerSendMessageTo)
			})
			r.Route("/units", func(r chi.Router) {
				// GET /api/v1/units/
				r.Get("/", api.PeerListUnits)
			})
		})
	})
}

func Setup(keys *crypto.KeyPair, peer *mesh.Peer, routes bool) (err error, api *API) {
	api = &API{
		Router: chi.NewRouter(),
		Keys:   keys,
		Peer:   peer,
		Client: NewClient(keys),
	}

	// api.Router.Use(CORS)
	if api.Keys == nil {
		api.Router.Use(middleware.DefaultCompress)
		api.setupServerRoutes()
	} else {
		api.setupPeerRoutes()
	}

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
