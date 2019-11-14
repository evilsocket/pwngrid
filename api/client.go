package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/crypto"
	"github.com/evilsocket/pwngrid/models"
	"github.com/evilsocket/pwngrid/utils"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	ClientTimeout   = 60
	ClientTokenFile = "/tmp/pwngrid-api-enrollment.json"
)

const (
	Endpoint = "https://api.pwnagotchi.ai/api/v1"
)

type Client struct {
	sync.Mutex

	cli     *http.Client
	keys    *crypto.KeyPair
	token   string
	tokenAt time.Time
	data    map[string]interface{}
}

func NewClient(keys *crypto.KeyPair) *Client {
	cli := &Client{
		cli: &http.Client{
			Timeout: time.Duration(ClientTimeout) * time.Second,
		},
		keys: keys,
		data: make(map[string]interface{}),
	}

	if info, err := os.Stat(ClientTokenFile); err == nil {
		if time.Since(info.ModTime()) < models.TokenTTL {
			log.Debug("loading token from %s ...", ClientTokenFile)
			var data map[string]interface{}
			if raw, err := ioutil.ReadFile(ClientTokenFile); err == nil {
				if err := json.Unmarshal(raw, &data); err == nil {
					cli.token = data["token"].(string)
					cli.tokenAt = info.ModTime()
					log.Debug("token: %s", cli.token)
				} else {
					log.Warning("error decoding %s: %v", ClientTokenFile, err)
				}
			} else {
				log.Warning("error reading %s: %v", ClientTokenFile, err)
			}
		} else {
			log.Debug("token in %s is expired", ClientTokenFile)
		}
	}

	return cli
}

func (c *Client) enroll() error {
	identity := fmt.Sprintf("%s@%s", utils.Hostname(), c.keys.FingerprintHex)

	log.Debug("refreshing api token as %s ...", identity)

	signature, err := c.keys.SignMessage([]byte(identity))
	if err != nil {
		return err
	}

	signature64 := base64.StdEncoding.EncodeToString(signature)
	pubKeyPEM64 := base64.StdEncoding.EncodeToString(c.keys.PublicPEM)

	log.Debug("SIGN(%s) = %s", identity, signature64)

	enrollment := map[string]interface{}{
		"identity":   identity,
		"public_key": pubKeyPEM64,
		"signature":  signature64,
		"data":       c.data,
	}

	obj, err := c.request("POST", "/unit/enroll", enrollment, false)
	if err != nil {
		return err
	}

	c.tokenAt = time.Now()
	c.token = obj["token"].(string)
	log.Debug("new token: %s", c.token)

	if raw, err := json.Marshal(obj); err == nil {
		log.Debug("saving token to %s ...", ClientTokenFile)
		if err = ioutil.WriteFile(ClientTokenFile, raw, 0644); err != nil {
			log.Warning("error saving token to %s: %v", ClientTokenFile, err)
		}
	} else {
		log.Warning("error encoding token: %v", err)
	}

	return nil
}

func (c *Client) request(method string, path string, data interface{}, auth bool) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s%s", Endpoint, path)
	err := (error)(nil)
	started := time.Now()
	defer func() {
		if err == nil {
			log.Debug("%s %s (%s)", method, url, time.Since(started))
		} else {
			log.Error("%s %s (%s) %v", method, url, time.Since(started), err)
		}
	}()

	buf := new(bytes.Buffer)
	if data != nil {
		if err = json.NewEncoder(buf).Encode(data); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}

	if auth {
		if time.Since(c.tokenAt) >= models.TokenTTL {
			if err := c.enroll(); err != nil {
				return nil, err
			}
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	res, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var obj map[string]interface{}
	if err = json.Unmarshal(body, &obj); err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		err = fmt.Errorf("%d %s", res.StatusCode, obj["error"])
	}

	return obj, err
}

func (c *Client) SetData(newData map[string]interface{}) map[string]interface{} {
	c.Lock()
	defer c.Unlock()

	for key, val := range newData {
		if val == nil {
			delete(c.data, key)
		} else {
			c.data[key] = val
		}
	}

	return c.data
}

func (c *Client) Data() map[string]interface{} {
	c.Lock()
	defer c.Unlock()
	return c.data
}

func (c *Client) Request(method string, path string, data interface{}, auth bool) (map[string]interface{}, error) {
	c.Lock()
	defer c.Unlock()
	return c.request(method, path, data, auth)
}

func (c *Client) Get(path string, auth bool) (map[string]interface{}, error) {
	return c.Request("GET", path, nil, auth)
}

func (c *Client) Post(path string, what interface{}, auth bool) (map[string]interface{}, error) {
	return c.Request("POST", path, what, auth)
}

func (c *Client) PagedUnits(page int) (map[string]interface{}, error) {
	return c.Get(fmt.Sprintf("/units/?p=%d", page), false)
}

func (c *Client) Unit(fingerprint string) (map[string]interface{}, error) {
	return c.Get(fmt.Sprintf("/unit/%s", fingerprint), false)
}

func (c *Client) ReportAP(report interface{}) (map[string]interface{}, error) {
	return c.Post("/unit/report/ap", report, true)
}

func (c *Client) Inbox(page int) (map[string]interface{}, error) {
	return c.Get(fmt.Sprintf("/unit/inbox/?p=%d", page), true)
}

func (c *Client) InboxMessage(id int) (map[string]interface{}, error) {
	return c.Get(fmt.Sprintf("/unit/inbox/%d", id), true)
}

func (c *Client) MarkInboxMessage(id int, mark string) (map[string]interface{}, error) {
	return c.Get(fmt.Sprintf("/unit/inbox/%d/%s", id, mark), true)
}

func (c *Client) SendMessageTo(fingerprint string, msg Message) error {
	_, err := c.Post(fmt.Sprintf("/unit/%s/inbox", fingerprint), msg, true)
	return err
}
