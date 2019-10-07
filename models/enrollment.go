package models

import (
	"encoding/base64"
	"fmt"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/pwngrid/crypto"
	"strings"
)

type EnrollmentRequest struct {
	Identity    string                 `json:"identity"`   // name@SHA256(public_key)
	PublicKey   string                 `json:"public_key"` // BASE64(public_key.pem)
	Signature   string                 `json:"signature"`  // BASE64(SIGN(identity, private_key))
	Data        map[string]interface{} `json:"data"`       // misc data for the unit
	KeyPair     *crypto.KeyPair        `json:"-"`          // parsed from public_key
	Name        string                 `json:"-"`
	Fingerprint string                 `json:"-"` // SHA256(public_key)
	Address     string                 `json:"-"`
	Country     string                 `json:"-"`
}

func (enroll *EnrollmentRequest) Validate() error {
	// split the identity into name and fingerprint
	parts := strings.Split(enroll.Identity, "@")
	if len(parts) != 2 {
		return fmt.Errorf("error parsing the identity string: got %d parts", len(parts))
	}

	enroll.Name = str.Trim(parts[0])
	enroll.Fingerprint = str.Trim(strings.ToLower(parts[1]))
	if len(enroll.Fingerprint) != crypto.Hasher.Size()*2 {
		return fmt.Errorf("unexpected fingerprint length for %s", enroll.Fingerprint)
	}

	// parse the public key as b64 pem
	pubKeyPEM, err := base64.StdEncoding.DecodeString(enroll.PublicKey)
	if err != nil {
		return fmt.Errorf("error decoding the public key: %v", err)
	}

	enroll.KeyPair, err = crypto.FromPublicPEM(string(pubKeyPEM))
	if err != nil {
		return fmt.Errorf("error parsing the public key: %v", err)
	}

	enroll.PublicKey = string(enroll.KeyPair.PublicPEM)

	if enroll.KeyPair.FingerprintHex != enroll.Fingerprint {
		return fmt.Errorf("fingerprint mismatch: expected:%s got:%s", enroll.KeyPair.FingerprintHex, enroll.Fingerprint)
	}

	data := []byte(enroll.Identity)
	signature, err := base64.StdEncoding.DecodeString(enroll.Signature)
	if err != nil {
		return fmt.Errorf("error decoding the signature: %v", err)
	}

	if err := enroll.KeyPair.VerifyMessage(data, signature); err != nil {
		return fmt.Errorf("signature verification failed: %s", err)
	}

	return nil
}
