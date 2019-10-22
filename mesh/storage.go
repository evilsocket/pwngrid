package mesh

import (
	"encoding/json"
	"fmt"
	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"
)

type Storage struct {
	sync.Mutex
	path  string
	peers map[string]*jsonPeer
}

func StorageFromPath(path string) (err error, store *Storage) {
	if path, err = fs.Expand(path); err != nil {
		return err, nil
	}

	store = &Storage{
		path:  path,
		peers: make(map[string]*jsonPeer),
	}

	if !fs.Exists(path) {
		log.Debug("creating %s ...", path)
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			return
		}
	}

	err = fs.Glob(path, "*.json", func(fileName string) error {
		log.Debug("loading %s ...", fileName)
		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			return err
		}

		var peer jsonPeer
		if err = json.Unmarshal(data, &peer); err != nil {
			return err
		}

		store.peers[peer.Fingerprint] = &peer

		return nil
	})

	log.Debug("loaded %d known peers", len(store.peers))

	return
}

func (store *Storage) Size() int {
	store.Lock()
	defer store.Unlock()
	return len(store.peers)
}

func (store *Storage) Of(fingerprint string) *jsonPeer {
	store.Lock()
	defer store.Unlock()

	if peer, found := store.peers[fingerprint]; found {
		return peer
	}

	return nil
}

func (store *Storage) List() []*jsonPeer {
	store.Lock()
	defer store.Unlock()

	list := make([]*jsonPeer, 0)
	for _, peer := range store.peers {
		list = append(list, peer)
	}

	return list
}

func (store *Storage) Track(fingerprint string, peer *Peer) error {
	store.Lock()
	defer store.Unlock()

	if encounter, found := store.peers[fingerprint]; !found {
		// peer first encounter
		peer.Encounters = 1
		peer.MetAt = time.Now()
		peer.PrevSeenAt = peer.SeenAt
	} else {
		// we met this peer before
		encounter.Encounters++
		peer.PrevSeenAt = encounter.SeenAt
		peer.MetAt = encounter.MetAt
		peer.Encounters = encounter.Encounters
	}

	peer.SeenAt = time.Now()

	// save/update peer data in memory
	obj := peer.json()
	store.peers[fingerprint] = obj
	// save/update peer data on disk
	fileName := path.Join(store.path, fmt.Sprintf("%s.json", fingerprint))
	if data, err := json.Marshal(obj); err != nil {
		return err
	} else if err := ioutil.WriteFile(fileName, data, os.ModePerm); err != nil {
		return err
	}

	return nil
}
