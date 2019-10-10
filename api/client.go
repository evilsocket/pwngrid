package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/crypto"
	"github.com/evilsocket/pwngrid/models"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	ClientTimeout = 60
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
}

func NewClient(keys *crypto.KeyPair) *Client {
	return &Client{
		cli: &http.Client{
			Timeout: time.Duration(ClientTimeout) * time.Second,
		},
		keys: keys,
	}
}

func (c *Client) enroll() error {
	name, err := os.Hostname()
	if err != nil {
		return err
	}

	if strings.HasSuffix(name, ".local") {
		name = strings.Replace(name, ".local", "", -1)
	}

	identity := fmt.Sprintf("%s@%s", name, c.keys.FingerprintHex)
	log.Info("refreshing api token as %s ...", identity)

	signature, err := c.keys.SignMessage([]byte(identity))
	if err != nil {
		return err
	}

	signature64 := base64.StdEncoding.EncodeToString(signature)
	pubKeyPEM64 := base64.StdEncoding.EncodeToString(c.keys.PublicPEM)

	log.Debug("SIGN(%s) = %s", identity, signature64)

	brain := map[string]interface{}{}
	if data, err := ioutil.ReadFile("/root/brain.json"); err == nil {
		if err = json.Unmarshal(data, &brain); err == nil {
			log.Debug("brain: %v", brain)
		}
	}

	enrollment := map[string]interface{}{
		"identity":   identity,
		"public_key": pubKeyPEM64,
		"signature":  signature64,
		"data": map[string]interface{}{
			"brain": brain,
		},
	}

	obj, err := c.request("POST", "/unit/enroll", enrollment, false)
	if err != nil {
		return err
	}

	c.tokenAt = time.Now()
	c.token = obj["token"].(string)
	log.Debug("new token: %s", c.token)

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

func (c *Client) Inbox(page int) (map[string]interface{}, error) {
	return c.Get(fmt.Sprintf("/unit/inbox/?p=%d", page), true)
}

func (c *Client) InboxMessage(id int) (map[string]interface{}, error) {
	obj, err := c.Get(fmt.Sprintf("/unit/inbox/%d", id), true)
	log.Info("%v", obj)
	return obj, err
}

func (c *Client) SendMessageTo(fingerprint string, msg Message) error {
	_, err := c.Post(fmt.Sprintf("/unit/%s/inbox", fingerprint), msg, true)
	return err
}
