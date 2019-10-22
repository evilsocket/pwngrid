package api

import (
	"encoding/json"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/mesh"
	"github.com/go-chi/chi"
	"io/ioutil"
	"net/http"
	"sort"
)

// GET /api/v1/mesh/peers
func (api *API) PeerGetPeers(w http.ResponseWriter, r *http.Request) {
	peers := make([]*mesh.Peer, 0)
	mesh.Peers.Range(func(key, value interface{}) bool {
		peers = append(peers, value.(*mesh.Peer))
		return true
	})

	// closer first
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].RSSI > peers[j].RSSI
	})

	JSON(w, http.StatusOK, peers)
}

// GET /api/v1/mesh/memory
func (api *API) PeerGetMemory(w http.ResponseWriter, r *http.Request) {
	peers := api.Mesh.Memory()
	// more interactions first
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].Encounters > peers[j].Encounters
	})
	JSON(w, http.StatusOK, peers)
}

// GET /api/v1/mesh/<status>
func (api *API) PeerSetSignaling(w http.ResponseWriter, r *http.Request) {
	status := chi.URLParam(r, "status")

	if status == "enabled" || status == "true" {
		api.Peer.Advertise(true)
	} else if status == "disabled" || status == "false" {
		api.Peer.Advertise(false)
	} else {
		ERROR(w, http.StatusNotFound, ErrEmpty)
		return
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// GET /api/v1/mesh/data
func (api *API) PeerGetMeshData(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, api.Peer.Data())
}

// POST /api/v1/mesh/data
func (api *API) PeerSetMeshData(w http.ResponseWriter, r *http.Request) {
	var newData map[string]interface{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	log.Debug("%s", body)

	if err = json.Unmarshal(body, &newData); err != nil {
		ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// this makes sure that the pwngrid server receives advertisements
	api.Client.SetData(map[string]interface{}{
		"advertisement": newData,
	})

	// update mesh advertisement data
	api.Peer.SetData(newData)

	JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}