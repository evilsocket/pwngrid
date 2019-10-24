package api

import (
	"github.com/evilsocket/islazy/log"
	"github.com/go-chi/chi"
)

func (api *API) setupPeerRoutes() {
	log.Debug("registering peer api ...")

	api.Router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Route("/mesh", func(r chi.Router) {
				// GET /api/v1/mesh/peers
				r.Get("/peers", api.PeerGetPeers)

				r.Route("/memory", func(r chi.Router) {
					// GET /api/v1/mesh/memory
					r.Get("/", api.PeerGetMemory)
					// GET /api/v1/mesh/memory/<fingerprint>
					r.Get("/{fingerprint:[a-fA-F0-9]+}", api.PeerGetMemoryOf)
				})

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
