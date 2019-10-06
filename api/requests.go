package api

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/pwngrid/crypto"
	"strings"
)

type UnitEnrollmentRequest struct {
	// name@SHA256(public_key)
	Identity string `json:"identity"`
	// BASE64(public_key.pem)
	PublicKey string `json:"public_key"`
	// BASE64(SIGN(identity, private_key))
	Signature string `json:"signature"`

	// SHA256(public_key)
	fingerprint string
	// parsed from public_key
	pubKey *rsa.PublicKey
}

func (enroll UnitEnrollmentRequest) Validate() error {
	// split the identity into name and fingerprint
	parts := strings.Split(enroll.Identity, "@")
	if len(parts) != 2 {
		return fmt.Errorf("error parsing the identity string: got %d parts", len(parts))
	}

	enroll.fingerprint = str.Trim(strings.ToLower(parts[1]))
	if len(enroll.fingerprint) != crypto.Hasher.Size()*2 {
		return fmt.Errorf("unexpected fingerprint length for %s", enroll.fingerprint)
	}

	// parse the public key as b64 pem
	pubKeyPEM, err := base64.StdEncoding.DecodeString(enroll.PublicKey)
	if err != nil {
		return fmt.Errorf("error decoding the public key: %v", err)
	}

	keys, err := crypto.FromPublicPEM(string(pubKeyPEM))
	if err != nil {
		return fmt.Errorf("error parsing the public key: %v", err)
	}

	enroll.PublicKey = string(keys.PublicPEM)

	if keys.FingerprintHex != enroll.fingerprint {
		return fmt.Errorf("fingerprint mismatch: expected:%s got:%s", keys.FingerprintHex, enroll.fingerprint)
	}

	data := []byte(enroll.Identity)
	signature, err := base64.StdEncoding.DecodeString(enroll.Signature)
	if err != nil {
		return fmt.Errorf("error decoding the signature: %v", err)
	}

	if err := keys.VerifyMessage(data, signature); err != nil {
		return fmt.Errorf("signature verification failed: %s", err)
	}

	return nil
}