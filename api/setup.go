package api

import (
	"fmt"
	"github.com/evilsocket/pwngrid/crypto"
	"github.com/evilsocket/pwngrid/mesh"
	"github.com/go-chi/chi"
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

func Setup(keys *crypto.KeyPair, peer *mesh.Peer, routes bool) (err error, api *API) {
	api = &API{
		Router: chi.NewRouter(),
		Keys:   keys,
		Peer:   peer,
		Client: NewClient(keys),
	}

	api.Router.Use(CORS)
	if api.Keys == nil {
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
