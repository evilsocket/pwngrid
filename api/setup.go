package api

import (
	"github.com/evilsocket/pwngrid/crypto"
	"github.com/evilsocket/pwngrid/mesh"
	"github.com/go-chi/chi"
	"net/http"

	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/evilsocket/islazy/log"
)

type API struct {
	Router *chi.Mux
	Keys   *crypto.KeyPair
	Peer   *mesh.Peer
	Mesh   *mesh.Router
	Client *Client
}

func Setup(keys *crypto.KeyPair, peer *mesh.Peer, router *mesh.Router) (err error, api *API) {
	api = &API{
		Router: chi.NewRouter(),
		Keys:   keys,
		Peer:   peer,
		Mesh:   router,
		Client: NewClient(keys),
	}

	api.Router.Use(CORS)
	if api.Keys == nil {
		api.setupServerRoutes()
	} else {
		api.setupPeerRoutes()
	}

	return
}

func (api *API) Run(addr string) {
	log.Info("pwngrid api starting on %s ...", addr)
	log.Fatal("%v", http.ListenAndServe(addr, api.Router))
}
