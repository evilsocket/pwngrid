package api

import (
	"encoding/json"
	"github.com/evilsocket/islazy/log"
	"io/ioutil"
	"net/http"
	"time"
)

var client = &http.Client {
	Timeout: 60 * time.Second,
}

func GetJSON(url string) (map[string]interface{}, error) {
	err := (error)(nil)
	started := time.Now()
	defer func() {
		log.Debug("GET %s (%s) %v", url, time.Since(started), err)
	}()

	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var obj map[string]interface{}

	if err = json.Unmarshal(body, &obj); err != nil {
		return nil, err
	}

	return obj, nil
}
