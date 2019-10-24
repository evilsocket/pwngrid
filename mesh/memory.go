package mesh

import (
	"encoding/json"
	"fmt"
	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
	"io/ioutil"
	"math"
	"os"
	"path"
	"sync"
	"time"
)

type Memory struct {
	sync.Mutex
	path  string
	peers map[string]*Peer
}

func MemoryFromPath(path string) (err error, mem *Memory) {
	if path, err = fs.Expand(path); err != nil {
		return err, nil
	}

	mem = &Memory{
		path:  path,
		peers: make(map[string]*Peer),
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
			log.Error("error loading %s: %v", fileName, err)
			return nil
		}

		var peer jsonPeer
		if err = json.Unmarshal(data, &peer); err != nil {
			log.Error("error loading %s: %v", fileName, err)
			return nil
		}

		mem.peers[peer.Fingerprint] = peerFromJSON(peer)
		return nil
	})

	log.Debug("loaded %d known peers", len(mem.peers))

	return
}

func (mem *Memory) Size() int {
	mem.Lock()
	defer mem.Unlock()
	return len(mem.peers)
}

func (mem *Memory) Of(fingerprint string) *Peer {
	mem.Lock()
	defer mem.Unlock()

	if peer, found := mem.peers[fingerprint]; found {
		return peer
	}

	return nil
}

func (mem *Memory) List() []*Peer {
	mem.Lock()
	defer mem.Unlock()

	list := make([]*Peer, 0)
	for _, peer := range mem.peers {
		list = append(list, peer)
	}

	return list
}

func (mem *Memory) Track(fingerprint string, peer *Peer) error {
	mem.Lock()
	defer mem.Unlock()

	if encounter, found := mem.peers[fingerprint]; !found {
		// peer first encounter
		peer.Encounters = 1
		peer.MetAt = time.Now()
		peer.PrevSeenAt = peer.SeenAt
	} else {
		// we met this peer before
		if encounter.Encounters < math.MaxUint64 {
			encounter.Encounters++
		}
		peer.PrevSeenAt = encounter.SeenAt
		peer.MetAt = encounter.MetAt
		peer.Encounters = encounter.Encounters
	}

	peer.SeenAt = time.Now()

	// save/update peer data in memory
	mem.peers[fingerprint] = peer
	// save/update peer data on disk
	fileName := path.Join(mem.path, fmt.Sprintf("%s.json", fingerprint))
	if data, err := json.Marshal(peer); err != nil {
		return err
	} else if err := ioutil.WriteFile(fileName, data, os.ModePerm); err != nil {
		return err
	}

	return nil
}
